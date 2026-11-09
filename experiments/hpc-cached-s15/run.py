#!/usr/bin/env python3
"""Validate Linux portability and reproduce the unchanged cached s15 probe.

Run inside the bounded Slurm allocation in run.sbatch. The isolated source copy
and stage cache protect reference CSVs and preprocessing from output-writing tests.
"""
import hashlib
import json
import os
from pathlib import Path
import shutil
import signal
import subprocess
import sys
import time

ROOT = Path(__file__).resolve().parents[2]
HERE = Path(__file__).resolve().parent


def sha(path):
    h = hashlib.sha256()
    with path.open('rb') as f:
        for block in iter(lambda: f.read(8 << 20), b''):
            h.update(block)
    return h.hexdigest()


def group_rss(group):
    """Sum Linux resident process pages in the command's private process group."""
    pages = 0
    for path in Path('/proc').glob('[0-9]*/stat'):
        try:
            fields = path.read_text().rsplit(')', 1)[1].split()
            if int(fields[2]) == group:
                pages += int((path.parent / 'statm').read_text().split()[1])
        except (FileNotFoundError, ProcessLookupError, PermissionError):
            continue
    return pages * os.sysconf('SC_PAGE_SIZE')


def verify(phases, reference):
    """Require the retained full-word digest and all existing validation phases."""
    by_name = {p['phase']: p for p in phases}
    assert by_name['verify-0']['canonical_sha256'] == reference['canonical_sha256']
    assert by_name['verify-0']['systematic_checked'] == 216914047
    assert by_name['verify-0']['curves_checked'] == 16
    assert by_name['pack-output']['sha256'] == reference['sha256']
    assert by_name['pack-output']['packed_bytes'] == 2013265920
    assert by_name['verify-packed']['readback_match']
    assert by_name['verify-packed']['positions_checked'] == 1027


def main():
    if len(sys.argv) != 2 or not os.environ.get('SLURM_JOB_ID'):
        raise SystemExit('Usage: submit run.sbatch RUN_NAME through Slurm')
    name = sys.argv[1]
    if not name or any(c not in 'abcdefghijklmnopqrstuvwxyz0123456789-_' for c in name):
        raise SystemExit('Run name must contain lowercase letters, digits, - or _')
    out = HERE / 'runs' / name
    out.mkdir(parents=True, exist_ok=False)
    job = os.environ['SLURM_JOB_ID']
    scratch = Path('/tmp') / ('sk-depir-hpc-' + job)
    scratch.mkdir(exist_ok=False)
    durable = ROOT / '.cache/liftrs/hpc-cached-s15' / name
    durable.mkdir(parents=True, exist_ok=False)
    env = dict(os.environ, GOTOOLCHAIN='local', CGO_ENABLED='1', CXX='g++',
               GOMAXPROCS='16', GOFLAGS='-p=4', GOPROXY='off', GOSUMDB='off',
               GOPATH=str(ROOT / '.cache/liftrs/toolchains/gopath'),
               GOCACHE=str(scratch / 'go-cache'), TMPDIR=str(scratch))
    go = ROOT / '.cache/liftrs/toolchains/go1.27.3-linux-amd64/go/bin/go'
    record = {'status': 'running', 'job_id': job, 'scratch': str(scratch),
              'durable_cache': str(durable), 'commands': [],
              'slurm': {k: v for k, v in os.environ.items() if k.startswith('SLURM_')},
              'revision': subprocess.check_output(['git', 'rev-parse', 'HEAD'], cwd=ROOT, text=True).strip()}

    def save():
        temporary = out / 'results.json.tmp'
        temporary.write_text(json.dumps(record, indent=2) + '\n')
        temporary.replace(out / 'results.json')

    def run(label, command, cwd=ROOT, required=True, timeout=600):
        start = time.monotonic()
        with (out / (label + '.log')).open('w') as log:
            p = subprocess.Popen(command, cwd=cwd, env=env, stdout=log,
                                 stderr=subprocess.STDOUT, start_new_session=True)
            peak = 0
            guard = None
            while p.poll() is None:
                peak = max(peak, group_rss(p.pid))
                if peak > 64 << 30 or time.monotonic() - start > timeout:
                    guard = '64 GiB sampled process-group RSS' if peak > 64 << 30 else 'command timeout'
                    os.killpg(p.pid, signal.SIGTERM)
                    try:
                        p.wait(timeout=10)
                    except subprocess.TimeoutExpired:
                        os.killpg(p.pid, signal.SIGKILL)
                        p.wait()
                    break
                time.sleep(1)
        record['commands'].append({'label': label, 'argv': list(map(str, command)),
                                   'exit_code': p.returncode,
                                   'sampled_group_peak_rss_bytes': peak,
                                   'guard_stop': guard,
                                   'elapsed_seconds': time.monotonic() - start})
        save()
        if guard or (required and p.returncode):
            raise RuntimeError(label + ' failed; see retained log')

    save()
    try:
        for label, command in [
            ('hostname', ['hostname']), ('cpu', ['lscpu']), ('ram', ['free', '-b']),
            ('storage', ['df', '-T', '-B1', str(ROOT), str(scratch)]),
            ('os', ['cat', '/etc/os-release']), ('compiler', ['g++', '--version']),
            ('go', [str(go), 'version']),
            ('allocation', ['scontrol', 'show', 'job', job]),
            ('gpu-visible', ['nvidia-smi', '--query-gpu=name,memory.total,driver_version', '--format=csv']),
            ('perf', ['perf', 'stat', '-e', 'task-clock', 'true'])]:
            try:
                run(label, command, required=label not in ('gpu-visible', 'perf'))
            except FileNotFoundError as error:
                if label not in ('gpu-visible', 'perf'):
                    raise
                record[label + '_unavailable'] = str(error)
        # Read the job's actual cgroup and affinity, not only visible host RAM.
        record['proc_status'] = Path('/proc/self/status').read_text()
        group = Path('/proc/self/cgroup').read_text()
        record['cgroup'] = group
        for line in group.splitlines():
            if line.startswith('0::'):
                base = Path('/sys/fs/cgroup') / line[3:].lstrip('/')
                record['cgroup_limits'] = {
                    str(p): p.read_text().strip()
                    for directory in [base, *base.parents]
                    if str(directory).startswith('/sys/fs/cgroup')
                    for field in ('memory.max', 'memory.high', 'cpu.max', 'cpuset.cpus.effective')
                    if (p := directory / field).is_file()}
        source = scratch / 'repo'
        shutil.copytree(ROOT, source, ignore=shutil.ignore_patterns('.git', '.cache', 'runs', '__pycache__'))
        record['source_sha256'] = {str(p.relative_to(source)): sha(p)
                                   for p in source.rglob('*') if p.is_file()
                                   and (p.suffix in ('.go', '.cpp', '.h', '.hpp', '.py', '.sbatch')
                                        or p.name in ('go.mod', 'go.sum', 'Makefile'))}
        # Existing full suites write CSV fixtures; all run in the isolated copy.
        run('race', [str(go), 'test', '-race', '-tags', 'liftrs_cpp', './encoding/liftrs-shared',
                     '-run', 'TestCPPMatrix|TestCPPOnline|TestScalingCPPLargeFields|TestScalingOnlineLargeShape',
                     '-count=1', '-timeout=8m'], cwd=source)
        run('vet', [str(go), 'vet', '-tags', 'liftrs_cpp', './encoding/...', './pirconfig', './planner'], cwd=source)
        run('legacy-build', ['make', '-C', 'encoding/rme_legacy', 'rme'], cwd=source)
        env['CGO_ENABLED'] = '0'
        run('suite-go', [str(go), 'test', './...', '-timeout=8m'], cwd=source)
        env['CGO_ENABLED'] = '1'
        run('suite-native', [str(go), 'test', '-tags', 'liftrs_cpp', './...', '-timeout=8m'], cwd=source)
        binary = durable / 'probe'
        run('build-probe', [str(go), 'test', '-c', '-tags', 'liftrs_cpp',
                            '-o', str(binary), './encoding/liftrs-shared'], cwd=source)
        record['binary_sha256'] = sha(binary)
        cache = scratch / 'cache'
        cache.mkdir()
        reference = ROOT / '.cache/liftrs/parallel-s15/s15-m2-t5'
        manifest = json.loads((ROOT / 'experiments/parallel-s15/initial/results.json').read_text())['cache_sha256']
        start = time.monotonic()
        for filename in ('online.bin', 'online-metadata.json'):
            shutil.copy2(reference / filename, cache / filename)
            assert sha(cache / filename) == manifest[filename], filename
        record['stage_and_verify_seconds'] = time.monotonic() - start
        env.update(PIR_S15_OUT=str(out), PIR_S15_CACHE=str(cache), PIR_S15_ONLINE_ONLY='1')
        run('cached-encoding', ['/usr/bin/time', '-v', str(binary), '-test.run=^TestS15FirstUse$',
                                '-test.v', '-test.timeout=35m'], cwd=source, timeout=2160)
        phases = json.loads((out / 'phases.json').read_text())
        verify(phases, json.loads((reference / 'word-metadata.json').read_text()))
        record['validation'] = 'canonical/packed digests, all systematic positions, 16 curves, 1027 packed reads pass'
        # Persist new output, leaving original fixture and preprocessing untouched.
        start = time.monotonic()
        for filename in ('word-packed.bin', 'word-metadata.json'):
            shutil.copy2(cache / filename, durable / filename)
        record['output_retention_seconds'] = time.monotonic() - start
        record['status'] = 'finalizing'
        save()
        record['evidence_sha256'] = {p.name: sha(p) for p in out.iterdir()
                                      if p.is_file() and p.name not in ('results.json', 'results.json.tmp')}
        record['retained_output_sha256'] = sha(durable / 'word-packed.bin')
        assert record['retained_output_sha256'] == manifest['word-packed.bin']
        record['status'] = 'passed'
        save()
    except BaseException as error:
        record['status'] = 'failed'
        record['error'] = repr(error)
        save()
        raise
    print('HPC cached s15 validation passed', flush=True)


if __name__ == '__main__':
    main()
