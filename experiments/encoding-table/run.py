#!/usr/bin/env python3
"""Fill unmeasured s<=16 encoding cells with existing guarded verification harnesses."""
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
GIB = 1 << 30
HELP = runpy.run_path(str(ROOT / 'experiments/hpc-cached-s15/run.py'))
GUARD = runpy.run_path(str(ROOT / 'experiments/hpc-s16/run.py'))
MEM = runpy.run_path(str(ROOT / 'experiments/cpu-streamed-constants/cache_memory.py'))
DIGESTS = {
    10: "3ab552f2ea40fe7b8ee39e2aabde2852f8fc536e9945fb7237fee188f357877e",
    14: "88d5ff39c5df8bdd4aa0f85fc19833ae673df95365b4fd862bb83936e5a0adf8",
    15: "59f223d593a3706de5d6692bff9b8378b34a07936a7bf33d4fca2382e20335a3",
    16: "f33dc66cb3b616697866a298ad04dd3d4853fb98bf840f3c81976955bc895874",
}
REFERENCES = {
    10: '.cache/liftrs/gpu-products/products-02/cpu-0',
    14: '.cache/liftrs/parallel-s14/s14-m2-t5',
    15: '.cache/liftrs/parallel-s15/s15-m2-t5',
    16: '.cache/liftrs/gpu-preparation/s16-m2-t5',
}
# Each cell is missing from the reference at61b1857. No repeated preprocessing.
CASES = {
    'cpu-compact-s10': (10, 'TestCPUCompactCached', {'PIR_CPU_COMPACT_MODE': 'compact'}, 'PIR_CPU_COMPACT'),
    'cpu-packed-s14': (14, 'TestGPUPackedCached', {'PIR_TABLE_CPU_PACKED': '1', 'PIR_GPU_PACKED_MODE': 'packed'}, 'PIR_GPU_PACKED'),
    'cpu-disk-s14': (14, 'TestCPPDiskOnlineWorkflow', {}, 'PIR_CPU_DISK_WORKFLOW'),
    'cpu-reuse-s15': (15, 'TestCPUMemoryCached', {}, 'PIR_CPU_MEMORY'),
    'cpu-compact-s15': (15, 'TestCPUCompactCached', {'PIR_CPU_COMPACT_MODE': 'compact'}, 'PIR_CPU_COMPACT'),
    'cpu-file-s15': (15, 'TestCPUFileCached', {'PIR_CPU_FILE_MODE': 'file'}, 'PIR_CPU_FILE'),
    'cpu-packed-s15': (15, 'TestGPUPackedCached', {'PIR_TABLE_CPU_PACKED': '1', 'PIR_GPU_PACKED_MODE': 'packed'}, 'PIR_GPU_PACKED'),
    'cpu-disk-s15': (15, 'TestCPPDiskOnlineWorkflow', {}, 'PIR_CPU_DISK_WORKFLOW'),
    'cpu-packed-s16': (16, 'TestGPUPackedCached', {'PIR_TABLE_CPU_PACKED': '1', 'PIR_GPU_PACKED_MODE': 'packed'}, 'PIR_GPU_PACKED'),
    'cpu-disk-s16': (16, 'TestCPPDiskOnlineWorkflow', {}, 'PIR_CPU_DISK_WORKFLOW'),
    'gpu-stream-s10': (10, 'TestCUDACachedEncoding', {'PIR_CUDA_STREAM': '1'}, 'PIR_CUDA_CACHED'),
    'gpu-packed-s10': (10, 'TestGPUPackedCached', {'PIR_GPU_PACKED_STREAM': '1', 'PIR_GPU_PACKED_MODE': 'packed'}, 'PIR_GPU_PACKED'),
    'gpu-original-s14': (14, 'TestCUDACachedEncoding', {}, 'PIR_CUDA_CACHED'),
    'gpu-compact-s14': (14, 'TestCUDACachedEncoding', {'PIR_CUDA_COMPACT': '1'}, 'PIR_CUDA_CACHED'),
    'gpu-host-s14': (14, 'TestCUDAEncodingWorkflow', {'PIR_CUDA_HOST_WORKFLOW': '1'}, 'PIR_CUDA_WORKFLOW'),
    'gpu-original-s15': (15, 'TestCUDACachedEncoding', {}, 'PIR_CUDA_CACHED'),
    'gpu-reuse-s15': (15, 'TestCUDACachedEncoding', {'PIR_CUDA_REUSE': '1'}, 'PIR_CUDA_CACHED'),
    'gpu-compact-s15': (15, 'TestCUDACachedEncoding', {'PIR_CUDA_COMPACT': '1'}, 'PIR_CUDA_CACHED'),
    'gpu-stream-s15': (15, 'TestCUDACachedEncoding', {'PIR_CUDA_STREAM': '1'}, 'PIR_CUDA_CACHED'),
    'gpu-packed-s15': (15, 'TestGPUPackedCached', {'PIR_GPU_PACKED_STREAM': '1', 'PIR_GPU_PACKED_MODE': 'packed'}, 'PIR_GPU_PACKED'),
    'gpu-host-s15': (15, 'TestCUDAEncodingWorkflow', {'PIR_CUDA_HOST_WORKFLOW': '1'}, 'PIR_CUDA_WORKFLOW'),
    'gpu-host-s16': (16, 'TestCUDAEncodingWorkflow', {'PIR_CUDA_HOST_WORKFLOW': '1'}, 'PIR_CUDA_WORKFLOW'),
}


def policy(case):
    s = CASES[case][0]
    ram = 16 if s == 10 else 32 if s < 16 else 128
    return {'cpus': 32, 'host_request_bytes': ram * GIB,
            'rss_guard_bytes': min(96, ram * 3 // 4) * GIB,
            'cgroup_guard_bytes': min(112, ram - 2) * GIB,
            'new_cache_guard_bytes': (64 if case == 'cpu-disk-s16' else 16) * GIB,
            'scratch_guard_bytes': 16 * GIB, 'device_cap_bytes': 0 if case.startswith('cpu-') else (512 << 20) if '-host-' in case else 40 * GIB,
            'controller_seconds': 6600, 'worker_seconds': 5400, 'slurm_seconds': 7200}


def check():
    assert len(CASES) == 22
    assert all(s in REFERENCES for s, _, _, _ in CASES.values())
    assert all(int(case.rsplit('s', 1)[1]) == row[0] for case, row in CASES.items())
    assert policy('cpu-disk-s16')['new_cache_guard_bytes'] == 64 * GIB
    assert all(policy(c)['new_cache_guard_bytes'] == 16 * GIB for c in CASES if c != 'cpu-disk-s16')
    assert all(policy(c)['rss_guard_bytes'] < policy(c)['host_request_bytes'] for c in CASES)


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('batch', nargs='?')
    parser.add_argument('case', nargs='?', choices=CASES)
    parser.add_argument('--check', action='store_true')
    args = parser.parse_args()
    check()
    if args.check:
        print('22 supported table cells; bounded storage/RAM policies checked')
        return
    if not args.batch or not args.batch.replace('-', '').isalnum() or not args.case:
        parser.error('fresh batch name and case required')
    case = args.case
    s, test, switches, prefix = CASES[case]
    job = os.environ['SLURM_JOB_ID']
    limits = policy(case)
    out = HERE / 'runs' / args.batch / case
    out.mkdir(parents=True, exist_ok=False)
    cache = ROOT / '.cache/liftrs/encoding-table' / args.batch / case
    cache.mkdir(parents=True, exist_ok=False)
    scratch = Path('/tmp') / ('sk-depir-table-' + job)
    scratch.mkdir(exist_ok=False)
    reference = ROOT / REFERENCES[s]
    for name in ('online.bin', 'online-metadata.json'):
        (cache / name).symlink_to(reference / name)
    env = {k: v for k, v in os.environ.items() if not k.startswith('PIR_')}
    env.update(GOTOOLCHAIN='local', GOPROXY='off', GOSUMDB='off',
               GOPATH=str(ROOT / '.cache/liftrs/toolchains/gopath'), GOCACHE=str(scratch / 'go-cache'),
               TMPDIR=str(scratch), CGO_ENABLED='1', CXX='g++', GOMAXPROCS='32', GOFLAGS='-p=4',
               PIR_TABLE_S=str(s), **switches)
    env[prefix + '_OUT'] = str(out / 'workflow')
    env[prefix + '_CACHE'] = str(cache / 'workflow')
    env[prefix + '_REFERENCE'] = str(reference)
    work = cache / 'workflow'
    work.mkdir()
    for name in ('online.bin', 'online-metadata.json'):
        (work / name).symlink_to(reference / name)
    (out / 'workflow').mkdir()
    go = str(ROOT / '.cache/liftrs/toolchains/go1.27.3-linux-amd64/go/bin/go')
    binary = cache / 'encoding.test'
    record = {'status': 'running', 'job_id': job, 'case': case, 's': s, 'm': 2, 't': 5,
              'machine_class': 'hpc', 'started_utc': datetime.datetime.now(datetime.timezone.utc).isoformat(),
              'revision': subprocess.check_output(['git', 'rev-parse', 'HEAD'], cwd=ROOT, text=True).strip(),
              'policy': limits, 'cache': str(cache), 'commands': []}
    began = time.monotonic()
    cancelled = []
    signal.signal(signal.SIGTERM, lambda *_: cancelled.append(True))
    observer = MEM['FileCache']()

    def save():
        tmp = out / 'results.json.tmp'
        tmp.write_text(json.dumps(record, indent=2) + '\n')
        tmp.replace(out / 'results.json')

    def run(label, argv, cwd=ROOT):
        start = time.monotonic()
        peak = peak_cache = peak_combined = 0
        reason = None
        last_storage = 0
        with (out / (label + '.log')).open('w') as log, (out / 'progress.jsonl').open('a') as progress:
            proc = subprocess.Popen(argv, cwd=cwd, env=env, stdout=log, stderr=subprocess.STDOUT, start_new_session=True)
            try:
                while proc.poll() is None:
                    now = time.monotonic()
                    rss = HELP['group_rss'](proc.pid)
                    peak = max(peak, rss)
                    cg = MEM['job_cgroup'](job)
                    if now - last_storage > 10:
                        new = GUARD['tree_bytes'](cache)
                        temp = GUARD['tree_bytes'](scratch)
                        paths = [reference / 'online.bin'] + [p for p in cache.rglob('*') if p.is_file() and not p.is_symlink()]
                        resident = sum(observer.sample(paths).values())
                        peak_cache = max(peak_cache, resident)
                        peak_combined = max(peak_combined, rss + resident)
                        last_storage = now
                        if new > limits['new_cache_guard_bytes'] or temp > limits['scratch_guard_bytes']:
                            reason = 'new cache/scratch cap'
                    if peak > limits['rss_guard_bytes']: reason = 'RSS guard'
                    if cg and cg['current_bytes'] > limits['cgroup_guard_bytes']: reason = 'cgroup RAM guard'
                    if GUARD['mem_available']() < 16 * GIB: reason = 'node RAM headroom'
                    if now - began > limits['controller_seconds']: reason = 'controller deadline'
                    if now - start > limits['worker_seconds']: reason = 'worker deadline'
                    if cancelled: reason = 'termination signal'
                    progress.write(json.dumps({'command': label, 'rss_bytes': rss, 'tracked_cache_bytes': peak_cache,
                                               'cgroup': cg, 'guard': reason}) + '\n')
                    progress.flush()
                    if reason:
                        GUARD['stop_group'](proc)
                        break
                    time.sleep(2)
            finally:
                if proc.poll() is None: GUARD['stop_group'](proc)
        record['commands'].append({'label': label, 'argv': list(map(str, argv)), 'seconds': time.monotonic() - start,
                                   'peak_rss_bytes': peak, 'peak_tracked_cache_bytes': peak_cache,
                                   'peak_rss_plus_tracked_cache_bytes': peak_combined,
                                   'exit_code': proc.returncode, 'guard': reason})
        save()
        if reason or proc.returncode: raise RuntimeError(label + ': ' + str(reason or proc.returncode))

    save()
    try:
        if shutil.disk_usage(cache).free < limits['new_cache_guard_bytes'] + 20 * GIB: raise RuntimeError('initial storage guard')
        if shutil.disk_usage(scratch).free < 20 * GIB: raise RuntimeError('initial scratch guard')
        meta = json.loads((reference / 'online-metadata.json').read_text())
        assert meta['sha256'] == DIGESTS[s]
        assert meta['identity_verified'] and (meta['s'], meta['t'], meta['n']) == (s, 5, 1 << (2 * s))
        assert (reference / 'online.bin').stat().st_size == meta['bytes']
        record['input'] = {'path': str(reference), 'metadata': meta}
        # The loader verifies the entire trusted digest before execution; this avoids a second large read here.
        source = scratch / 'repo'
        shutil.copytree(ROOT, source, ignore=shutil.ignore_patterns('.git', '.cache', 'runs', '__pycache__', 'slurm-*.log'))
        record['source_sha256'] = {str(p.relative_to(source)): HELP['sha'](p) for p in source.rglob('*')
                                   if p.is_file() and (p.suffix in ('.go', '.cpp', '.h', '.hpp', '.cu', '.py', '.sbatch') or p.name in ('go.mod', 'go.sum'))}
        for label, argv in [('hostname', ['hostname']), ('cpu', ['lscpu']), ('compiler', ['g++', '--version']),
                            ('go', [go, 'version']), ('job', ['scontrol', 'show', 'job', job])]:
            run(label, argv)
        if case.startswith('gpu-'):
            run('gpu', ['nvidia-smi', '--query-gpu=name,uuid,driver_version,memory.total', '--format=csv'])
            visible = env['CUDA_VISIBLE_DEVICES']
            assert ',' not in visible
            free = int(subprocess.check_output(['nvidia-smi', '-i', visible, '--query-gpu=memory.free', '--format=csv,noheader,nounits'], text=True).strip()) << 20
            if free < limits['device_cap_bytes'] + GIB: raise RuntimeError('device headroom guard')
            library = cache / 'libliftrs-cuda.so'
            run('cuda-build', ['nvcc', '-std=c++17', '-O3', '-lineinfo', '-arch=native', '-shared', '-Xcompiler=-fPIC',
                               'encoding/liftrs-cuda/matrix.cu', '-o', str(library)], source)
            env['PIR_GPU_LIBRARY'] = str(library)
        run('build', [go, 'test', '-tags', 'liftrs_cpp,liftrs_cuda', '-c', '-o', str(binary), './encoding/liftrs-shared'], source)
        run('controls', [str(binary), '-test.run=^TestScalingTableControls$', '-test.v'], source)
        run('workflow', [str(binary), '-test.run=^' + test + '$', '-test.v', '-test.timeout=90m'], source)
        phases = list((out / 'workflow').rglob('phases.json'))
        assert phases, 'workflow must execute, not skip'
        for path in phases:
            rows = json.loads(path.read_text())
            enc = [r for r in rows if r['phase'].startswith(('encode-', 'encode-output-'))]
            verified = [r for r in rows if r['phase'].startswith('verify-') and 'systematic_checked' in r]
            assert len(enc) == len(verified) == (1 if s == 16 else 2)
            assert all(r['systematic_checked'] == meta['k'] and r['curves_checked'] == 16 for r in verified)
            assert len({r['canonical_sha256'] for r in verified}) == 1
        record['phase_files'] = [str(p.relative_to(out)) for p in phases]
        record['status'] = 'passed'
    except BaseException as error:
        record['status'] = 'failed'
        record['error'] = repr(error)
        raise
    finally:
        observer.close()
        record['cache_sha256'] = {str(p.relative_to(cache)): HELP['sha'](p) for p in cache.rglob('*') if p.is_file() and not p.is_symlink()}
        record['evidence_sha256'] = {str(p.relative_to(out)): HELP['sha'](p) for p in out.rglob('*')
                                     if p.is_file() and p.name not in ('results.json', 'results.json.tmp')}
        record['elapsed_seconds'] = time.monotonic() - began
        save()


if __name__ == '__main__':
    main()
