#!/usr/bin/env python3
"""Bounded one-GPU exact s18 dimension, no full preprocessing arrays."""
import argparse
import datetime
import hashlib
import json
import os
from pathlib import Path
import runpy
import signal
import subprocess
import time

ROOT = Path(__file__).resolve().parents[2]
HERE = Path(__file__).resolve().parent
HELP = runpy.run_path(str(ROOT/'experiments/hpc-cached-s15/run.py'))
GUARD = runpy.run_path(str(ROOT/'experiments/hpc-s16/run.py'))
STOP = runpy.run_path(str(ROOT/'experiments/gpu-preprocessing-workflow/run.py'))['stop_process']
MODEL = runpy.run_path(str(HERE/'memory.py'))['write_model']
POLICY = dict(cpus=4, gpus=1, owned_device_cap_bytes=512 << 20,
              requested_host_bytes=8 << 30, rss_guard_bytes=6 << 30,
              retained_cap_bytes=2 << 30, scratch_cap_bytes=4 << 30,
              worker_seconds=600, controller_seconds=900, slurm_seconds=1200)
KNOWN = {10: 146641, 16: 918672421, 17: 3876566203}


def completed_dimension(path):
    rows = [json.loads(line) for line in Path(path).read_text().splitlines()
            if line.startswith('{')]
    complete = [row for row in rows if row.get('event') == 'completed' and
                row.get('phase') == 'exact-weighted-dimension']
    assert len(complete) == 1
    row = complete[0]
    assert row['completed'] == row['total'] == 3**row['s']
    assert row['n'] == 1 << (2*row['s']) and 0 < row['k'] < row['n']
    assert row['owned_device_bytes'] <= POLICY['owned_device_cap_bytes']
    return row


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('name', nargs='?')
    parser.add_argument('--revision')
    parser.add_argument('--check', action='store_true')
    args = parser.parse_args()
    assert POLICY['worker_seconds'] < POLICY['controller_seconds'] < POLICY['slurm_seconds']
    assert 8 <= POLICY['owned_device_cap_bytes'] and POLICY['gpus'] == 1
    if args.check:
        print('oneGPU/8-byte explicit payload, bounded deadlines; no allocation or fulls18 launch')
        return
    if not args.name or not args.name.replace('-', '').isalnum() or not args.revision:
        parser.error('fresh name and committed source revision required')
    job = os.environ['SLURM_JOB_ID']
    revision = subprocess.check_output(['git', 'rev-parse', args.revision+'^{commit}'], cwd=ROOT, text=True).strip()
    out = HERE/'runs'/args.name
    out.mkdir(parents=True, exist_ok=False)
    cache = ROOT/'.cache/liftrs/s18-preprocessing-admission'/args.name
    cache.mkdir(parents=True, exist_ok=False)
    scratch = Path('/tmp')/('depir-s18-count-'+job)
    scratch.mkdir(exist_ok=False)
    env = {k: v for k, v in os.environ.items() if not k.startswith('PIR_')}
    env.update(GOTOOLCHAIN='local', GOPROXY='off', GOSUMDB='off',
               GOPATH=str(ROOT/'.cache/liftrs/toolchains/gopath'),
               GOCACHE=str(scratch/'go-cache'), TMPDIR=str(scratch),
               GOMAXPROCS='4', GOFLAGS='-p=2', GOMEMLIMIT=str(4 << 30), CGO_ENABLED='0')
    go = str(ROOT/'.cache/liftrs/toolchains/go1.27.3-linux-amd64/go/bin/go')
    record = dict(status='running', job_id=job, revision=revision, cache=str(cache),
                  policy=POLICY, started_utc=datetime.datetime.now(datetime.timezone.utc).isoformat(),
                  dimensions=[], commands=[])
    began = time.monotonic()
    cancelled = []
    signal.signal(signal.SIGTERM, lambda *_: cancelled.append(True))

    def save():
        temp = out/'results.json.tmp'
        temp.write_text(json.dumps(record, indent=2)+'\n')
        temp.replace(out/'results.json')

    def event(phase, kind, **fields):
        row = dict(utc=datetime.datetime.now(datetime.timezone.utc).isoformat(),
                   phase=phase, event=kind, elapsed_seconds=time.monotonic()-began, **fields)
        with (out/'progress.jsonl').open('a') as stream:
            stream.write(json.dumps(row)+'\n')
            stream.flush()
        print(json.dumps(row), flush=True)

    def run(label, argv, cwd=ROOT):
        started = last = time.monotonic()
        peak = 0
        reason = shutdown = None
        event(label, 'started')
        with (out/(label+'.log')).open('w') as log:
            proc = subprocess.Popen(list(map(str, argv)), cwd=cwd, env=env,
                                    stdout=log, stderr=subprocess.STDOUT, start_new_session=True)
            try:
                while proc.poll() is None:
                    now = time.monotonic()
                    peak = max(peak, HELP['group_rss'](proc.pid))
                    if peak > POLICY['rss_guard_bytes']:
                        reason = 'RSS guard'
                    if GUARD['tree_bytes'](cache)+GUARD['tree_bytes'](out) > POLICY['retained_cap_bytes'] or GUARD['tree_bytes'](scratch) > POLICY['scratch_cap_bytes']:
                        reason = 'storage guard'
                    if now-started > POLICY['worker_seconds'] or now-began > POLICY['controller_seconds']:
                        reason = 'deadline'
                    if cancelled:
                        reason = 'termination signal'
                    if now-last >= 30:
                        event(label, 'heartbeat', peak_rss_bytes=peak, remaining_work='unknown')
                        last = now
                    if reason:
                        shutdown = STOP(proc)
                        break
                    time.sleep(1)
            finally:
                if proc.poll() is None and shutdown is None:
                    shutdown = STOP(proc)
        record['commands'].append(dict(label=label, argv=list(map(str, argv)),
            seconds=time.monotonic()-started, peak_rss_bytes=peak, exit_code=proc.returncode,
            guard=reason, shutdown=shutdown))
        save()
        event(label, 'failed' if reason or proc.returncode else 'completed')
        if reason or proc.returncode:
            raise RuntimeError(label+': '+str(reason or proc.returncode))

    save()
    try:
        source = scratch/'repo'
        source.mkdir()
        archive = scratch/'source.tar'
        with archive.open('wb') as stream:
            subprocess.run(['git', 'archive', revision], cwd=ROOT, stdout=stream, check=True)
        run('source-extract', ['tar', '-xf', archive, '-C', source])
        record['source_sha256'] = {str(p.relative_to(source)): HELP['sha'](p)
            for p in source.rglob('*') if p.is_file() and
            (p.suffix in ('.go', '.cpp', '.h', '.hpp', '.cu', '.cuh', '.py', '.sbatch') or
             p.name in ('go.mod', 'go.sum', 'reference.go.txt'))}
        save()
        for label, argv in [('host', ['hostname']), ('cpu', ['lscpu']),
                            ('job', ['scontrol', 'show', 'job', job]), ('go', [go, 'version']),
                            ('gpu', ['nvidia-smi', '--query-gpu=name,uuid,driver_version,memory.total', '--format=csv'])]:
            run(label, argv)
        ref = cache/'reference'
        reference_source = source/'experiments/s18-preprocessing-admission/reference.go.txt'
        renamed = source/'experiments/s18-preprocessing-admission/reference_probe.go'
        renamed.write_bytes(reference_source.read_bytes())
        record['reference_source_sha256'] = HELP['sha'](reference_source)
        run('reference-build', [go, 'build', '-o', ref, renamed], source)
        run('reference', [ref, cache/'reference.pprof'], source)
        tool, checker = cache/'count', cache/'class-test'
        run('class-build', ['g++', '-std=c++17', '-O2', 'experiments/s18-preprocessing-admission/class_test.cpp', '-o', checker], source)
        run('class-check', [checker], source)
        run('cuda-build', ['nvcc', '-std=c++17', '-O3', '-lineinfo', '-arch=native',
                          'experiments/s18-preprocessing-admission/count.cu', '-o', tool], source)
        references = [json.loads(line) for line in (out/'reference.log').read_text().splitlines() if line.startswith('{')]
        for i, row in enumerate(references):
            label = 'reference-'+str(i)
            run(label, [tool, str(row['s']), str(row['t'])], source)
            counted = completed_dimension(out/(label+'.log'))
            assert counted['k'] == row['k'], 'independent Go dimension differs'
        run('memcheck', ['compute-sanitizer', '--tool', 'memcheck', '--leak-check', 'full',
                        '--error-exitcode', '99', tool, '8', '5'], source)
        memcheck = (out/'memcheck.log').read_text()
        assert 'ERROR SUMMARY: 0 errors' in memcheck and 'LEAK SUMMARY: 0 bytes' in memcheck
        for s in (10, 16, 17, 18):
            label = 'dimension-s'+str(s)
            run(label, [tool, str(s), '5'], source)
            counted = completed_dimension(out/(label+'.log'))
            if s in KNOWN:
                assert counted['k'] == KNOWN[s], 'verified historical dimension differs'
            record['dimensions'].append(counted)
            save()
        count = next(r for r in record['dimensions'] if r['s'] == 18)
        MODEL(out/'memory-model.json', 18, count['k'])
        record['status'] = 'passed'
    except BaseException as error:
        record['status'] = 'failed'
        record['error'] = repr(error)
        raise
    finally:
        record['elapsed_seconds'] = time.monotonic()-began
        record['cache_sha256'] = {str(p.relative_to(cache)): HELP['sha'](p)
            for p in cache.rglob('*') if p.is_file()}
        record['evidence_sha256'] = {str(p.relative_to(out)): HELP['sha'](p)
            for p in out.rglob('*') if p.is_file() and p.name not in ('results.json', 'results.json.tmp')}
        save()


if __name__ == '__main__':
    main()
