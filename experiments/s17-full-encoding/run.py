#!/usr/bin/env python3
"""One guarded full CPU/GPU encoding, pinned to a committed source snapshot."""
import argparse
import datetime
import hashlib
import json
import os
from pathlib import Path
import runpy
import shutil
import signal
import subprocess
import time

ROOT = Path(__file__).resolve().parents[2]
HERE = Path(__file__).resolve().parent
HELP = runpy.run_path(str(ROOT/'experiments/hpc-cached-s15/run.py'))
GUARD = runpy.run_path(str(ROOT/'experiments/hpc-s16/run.py'))
STOP = runpy.run_path(str(ROOT/'experiments/gpu-preprocessing-workflow/run.py'))['stop_process']
MEM = runpy.run_path(str(ROOT/'experiments/cpu-streamed-constants/cache_memory.py'))
GIB = 1 << 30
PREPROCESSING = ROOT/'.cache/liftrs/gpu-s17-full-preprocessing/s17-m2-t5'


def policy(backend):
    return dict(cpus=32, host_request_bytes=(768 if backend == 'cpu' else 640)*GIB,
                rss_guard_bytes=(704 if backend == 'cpu' else 320)*GIB,
                cgroup_guard_bytes=(744 if backend == 'cpu' else 608)*GIB,
                go_soft_bytes=(640 if backend == 'cpu' else 288)*GIB,
                workspace_cap_bytes=320*GIB, device_cap_bytes=112*GIB if backend == 'gpu' else 0,
                gpus=4 if backend == 'gpu' else 0, retained_cap_bytes=64*GIB,
                scratch_cap_bytes=16*GIB, worker_seconds=46*3600,
                controller_seconds=47*3600, slurm_seconds=48*3600)


def compare_cpu_reference(record, reference):
    """Compare the same full code and deterministic database, without a rerun."""
    assert reference['status'] == 'passed' and reference['backend'] == 'cpu'
    assert record['code_manifest'] == reference['code_manifest']
    for phase in ('verify-canonical', 'pack-output'):
        assert record['phases']['phases'][phase]['sha256'] == reference['phases']['phases'][phase]['sha256'], phase


def validate():
    for backend in ('cpu', 'gpu'):
        p = policy(backend)
        assert p['rss_guard_bytes'] < p['cgroup_guard_bytes'] < p['host_request_bytes']
        assert p['go_soft_bytes'] < p['rss_guard_bytes']
        assert p['worker_seconds'] < p['controller_seconds'] < p['slurm_seconds']
        # These are below the previously admitted full preprocessing ceilings.
        assert p['host_request_bytes'] <= 864*GIB and p['device_cap_bytes'] <= 112*GIB
        assert p['retained_cap_bytes'] <= 512*GIB and p['scratch_cap_bytes'] <= 16*GIB
    previous = json.loads((ROOT/'experiments/gpu-s17-full-preprocessing/runs/resume-02/analysis.json').read_text())
    assert previous['status'] == 'passed' and previous['job_id'] == '75832'
    m = json.loads((PREPROCESSING/'code/manifest.json').read_text())
    assert m == previous['stage_manifests']['code']
    assert m['header']['K'] == 3876566203 and m['header']['N'] == 1 << 34
    assert (PREPROCESSING/'code/online.bin').stat().st_size == m['files']['online.bin']['bytes']
    return m


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('backend', choices=('cpu', 'gpu'))
    parser.add_argument('name', nargs='?')
    parser.add_argument('--revision')
    parser.add_argument('--check', action='store_true')
    parser.add_argument('--reference-run', help='reviewed CPU run whose complete GPU output hashes must match')
    args = parser.parse_args()
    manifest = validate()
    if args.check:
        print('verified s17 manifest/extent and per-backend budget ceilings checked; no payload read/allocation/launch')
        return
    if not args.name or not args.name.replace('-', '').isalnum() or not args.revision:
        parser.error('fresh name and explicit committed revision required')
    job = os.environ['SLURM_JOB_ID']
    budget = policy(args.backend)
    reference = None
    if args.reference_run:
        if args.backend != 'gpu' or not args.reference_run.replace('-', '').isalnum():
            parser.error('reference-run requires GPU and an alphanumeric/hyphen run name')
        reference = json.loads((HERE/'runs'/args.reference_run/'cpu'/'results.json').read_text())
        assert reference['status'] == 'passed' and reference['backend'] == 'cpu'
        assert reference['code_manifest'] == manifest
    assert int(os.environ.get('SLURM_CPUS_PER_TASK', '0')) >= budget['cpus']
    assert int(os.environ.get('SLURM_MEM_PER_NODE', '0')) >= budget['host_request_bytes']//(1 << 20)
    revision = subprocess.check_output(['git', 'rev-parse', args.revision+'^{commit}'], cwd=ROOT, text=True).strip()
    out = HERE/'runs'/args.name/args.backend
    out.mkdir(parents=True, exist_ok=False)
    cache = ROOT/'.cache/liftrs/s17-full-encoding'/args.name/args.backend
    cache.mkdir(parents=True, exist_ok=False)
    scratch = Path('/tmp')/('depir-s17-encoding-'+job)
    scratch.mkdir(exist_ok=False)
    assert shutil.disk_usage(cache).free > 96*GIB and shutil.disk_usage(scratch).free > 20*GIB
    env = {k: v for k, v in os.environ.items() if not k.startswith('PIR_')}
    env.update(GOTOOLCHAIN='local', GOPROXY='off', GOSUMDB='off',
               GOPATH=str(ROOT/'.cache/liftrs/toolchains/gopath'), GOCACHE=str(scratch/'go-cache'),
               TMPDIR=str(scratch), CGO_ENABLED='1', CXX='g++', GOMAXPROCS='32',
               GOFLAGS='-p=4', GOMEMLIMIT=str(budget['go_soft_bytes']), LIFTRS_CPU_PROGRESS='1')
    go = str(ROOT/'.cache/liftrs/toolchains/go1.27.3-linux-amd64/go/bin/go')
    began = time.monotonic()
    cancelled = []
    signal.signal(signal.SIGTERM, lambda *_: cancelled.append(True))
    record = dict(status='running', job_id=job, revision=revision, backend=args.backend,
                  started_utc=datetime.datetime.now(datetime.timezone.utc).isoformat(),
                  cache=str(cache), policy=budget, code_manifest=manifest, commands=[])

    # Pipeline arithmetic is frozen separately from this updated controller.
    controller = out/'controller.py.txt'
    shutil.copyfile(Path(__file__), controller)
    record['controller_sha256'] = HELP['sha'](controller)
    if reference is not None:
        (out/'cpu-reference.json').write_text(json.dumps(reference, indent=2)+'\n')
        record['reference_run'] = args.reference_run

    def save():
        temp = out/'results.json.tmp'
        temp.write_text(json.dumps(record, indent=2)+'\n')
        temp.replace(out/'results.json')

    def event(phase, kind, **fields):
        row = dict(utc=datetime.datetime.now(datetime.timezone.utc).isoformat(), phase=phase,
                   event=kind, elapsed_seconds=time.monotonic()-began, **fields)
        with (out/'progress.jsonl').open('a') as stream:
            stream.write(json.dumps(row)+'\n')
            stream.flush()
        print(json.dumps(row), flush=True)

    def run(label, argv, cwd=ROOT, timeout=600):
        start = time.monotonic()
        peak = 0
        reason = shutdown = None
        last = last_storage = start
        storage = (0, 0)
        event(label, 'started')
        with (out/(label+'.log')).open('w') as log:
            proc = subprocess.Popen(list(map(str, argv)), cwd=cwd, env=env,
                                    stdout=log, stderr=subprocess.STDOUT, start_new_session=True)
            try:
                while proc.poll() is None:
                    now = time.monotonic()
                    peak = max(peak, HELP['group_rss'](proc.pid))
                    cgroup = MEM['job_cgroup'](job)
                    if now-last_storage >= 30:
                        storage = (GUARD['tree_bytes'](cache)+GUARD['tree_bytes'](out),
                                   GUARD['tree_bytes'](scratch))
                        last_storage = now
                    if peak > budget['rss_guard_bytes']:
                        reason = 'RSS guard'
                    if cgroup and cgroup['current_bytes'] > budget['cgroup_guard_bytes']:
                        reason = 'cgroup memory guard'
                    if storage[0] > budget['retained_cap_bytes'] or storage[1] > budget['scratch_cap_bytes']:
                        reason = 'storage guard'
                    if shutil.disk_usage(cache).free < 32*GIB:
                        reason = 'shared filesystem headroom'
                    if now-start > timeout or now-began > budget['controller_seconds']:
                        reason = 'deadline'
                    if cancelled:
                        reason = 'termination signal'
                    if now-last >= 30:
                        event(label, 'heartbeat', peak_rss_bytes=peak, cgroup=cgroup,
                              retained_bytes=storage[0], scratch_bytes=storage[1], remaining_work='unknown')
                        last = now
                    if reason:
                        observation = dict(phase=label, reason=reason, peak_rss_bytes=peak,
                                           cgroup=cgroup, retained_bytes=storage[0], scratch_bytes=storage[1])
                        record['guard_observation'] = observation
                        event(label, 'guard-triggered', **{k: v for k, v in observation.items() if k != 'phase'})
                        save()
                        shutdown = STOP(proc)
                        break
                    time.sleep(2)
            finally:
                if proc.poll() is None and shutdown is None:
                    shutdown = STOP(proc)
        record['commands'].append(dict(label=label, argv=list(map(str, argv)),
            seconds=time.monotonic()-start, peak_rss_bytes=peak, exit_code=proc.returncode,
            guard=reason, shutdown=shutdown))
        save()
        event(label, 'failed' if reason or proc.returncode else 'completed', peak_rss_bytes=peak)
        if reason or proc.returncode:
            raise RuntimeError(label+': '+str(reason or proc.returncode))

    save()
    try:
        # Freeze tracked sources at submission, independent of later working edits.
        source = scratch/'repo'
        source.mkdir()
        archive = scratch/'source.tar'
        with archive.open('wb') as stream:
            subprocess.run(['git', 'archive', revision], cwd=ROOT, stdout=stream, check=True)
        run('source-extract', ['tar', '-xf', archive, '-C', source])
        record['source_sha256'] = {str(p.relative_to(source)): HELP['sha'](p)
            for p in source.rglob('*') if p.is_file() and
            (p.suffix in ('.go', '.cpp', '.hpp', '.h', '.cu', '.cuh', '.py', '.sbatch') or
             p.name in ('go.mod', 'go.sum', 'probe.go.txt'))}
        probe = source/'experiments/s17-full-encoding/probe.go.txt'
        record['probe_sha256'] = HELP['sha'](probe)
        shutil.copyfile(probe, source/'encoding/liftrs-shared/full_s17_probe_test.go')
        save()
        for label, argv in [('host', ['hostname']), ('cpu', ['lscpu']),
                            ('job', ['scontrol', 'show', 'job', job]), ('go', [go, 'version']),
                            ('compiler', ['g++', '--version'])]:
            run(label, argv)
        library = ''
        tags = 'liftrs_cpp'
        if args.backend == 'gpu':
            tags += ',liftrs_cuda'
            for label, argv in [('gpu', ['nvidia-smi', '--query-gpu=name,uuid,driver_version,memory.total,memory.free', '--format=csv']),
                                ('topology', ['nvidia-smi', 'topo', '-m'])]:
                run(label, argv)
            preflight = cache/'peer-probe'
            run('peer-probe-build', ['nvcc', '-std=c++17', '-O2', '-arch=native',
                '-Iencoding/liftrs-cuda', 'experiments/gpu-preprocessing-workflow/peer-probe.cu', '-o', preflight], source)
            run('peer-probe', [preflight], source, timeout=60)
            library = str(cache/'libliftrs-cuda.so')
            run('cuda-build', ['nvcc', '-std=c++17', '-O3', '-lineinfo', '-arch=native', '-shared',
                '-Xcompiler=-fPIC,-pthread', 'encoding/liftrs-cuda/matrix.cu',
                'encoding/liftrs-cuda/preparation.cu', '-o', library], source)
        binary = cache/'encoding.test'
        run('build', [go, 'test', '-tags', tags, '-c', '-o', binary, './encoding/liftrs-shared'], source)
        package = source/'encoding/liftrs-shared'
        smoke = cache/'smoke'
        env['PIR_S17_SMOKE_CACHE'] = str(smoke)
        run('smoke-fixture', [binary, '-test.run=^TestFullEncodingProbeSmoke$', '-test.v'], package)
        env.pop('PIR_S17_SMOKE_CACHE')
        small = json.loads((smoke/'cpu.json').read_text())
        small.update(Backend=args.backend, Cache=str(smoke/'run'), Library=library,
                     Devices=[0, 1, 2, 3] if library else [], DeviceCap=1 << 30 if library else 0)
        smoke_config = cache/'smoke-config.json'
        smoke_config.write_text(json.dumps(small, indent=2)+'\n')
        env['PIR_S17_ENCODING_CONFIG'] = str(smoke_config)
        run('smoke-pipeline', [binary, '-test.run=^TestFullS17Encoding$', '-test.v'], package)
        full = dict(Code=str(PREPROCESSING/'code/online.bin'),
                    Digest=manifest['files']['online.bin']['sha256'], Cache=str(cache),
                    Backend=args.backend, Library=library, Devices=[0, 1, 2, 3] if library else [],
                    S=17, T=5, Workers=32, MaxPayload=161*GIB,
                    Workspace=budget['workspace_cap_bytes'], DeviceCap=budget['device_cap_bytes'], Smoke=False)
        config = cache/'config.json'
        config.write_text(json.dumps(full, indent=2)+'\n')
        env['PIR_S17_ENCODING_CONFIG'] = str(config)
        run('encoding-pipeline', [binary, '-test.run=^TestFullS17Encoding$', '-test.v', '-test.timeout=46h'],
            package, timeout=budget['worker_seconds'])
        record['phases'] = json.loads((cache/'phases.json').read_text())
        assert record['phases']['phases']['pack-output']['bytes'] == 34*GIB
        assert (cache/'word-packed.bin').stat().st_size == 34*GIB
        if reference is not None:
            compare_cpu_reference(record, reference)
            record['cpu_comparison'] = 'complete canonical and packed SHA256 match reviewed CPU run'
        record['status'] = 'passed'
    except BaseException as error:
        record['status'] = 'failed'
        record['error'] = repr(error)
        raise
    finally:
        record['elapsed_seconds'] = time.monotonic()-began
        # Packed output already has producer streaming hash + full disk readback;
        # do not add a hidden34GiB exit pass. Seal its size/digest in phases.json.
        record['cache_sha256'] = {str(p.relative_to(cache)): HELP['sha'](p)
            for p in cache.rglob('*') if p.is_file() and p.name not in ('word-packed.bin', 'word-packed.bin.tmp')}
        record['evidence_sha256'] = {str(p.relative_to(out)): HELP['sha'](p)
            for p in out.rglob('*') if p.is_file() and p.name not in ('results.json', 'results.json.tmp')}
        save()


if __name__ == '__main__':
    main()
