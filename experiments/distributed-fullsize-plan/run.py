#!/usr/bin/env python3
"""Read-only s17 root inventory; no GPU, large allocation or new data cache."""
import argparse
import hashlib
import json
import os
from pathlib import Path
import resource
import signal
import shutil
import subprocess
import time

ROOT = Path(__file__).resolve().parents[2]
HERE = Path(__file__).resolve().parent
SUPPORT = ROOT / '.cache/liftrs/gpu-s17-full-preprocessing/s17-m2-t5/support/support.bin'
SHA = '1d5251ecf13148683a28bff04b7e3f1c5f71cd8e3094eb083e147cb28c2e2419'
COUNT = 3876566203


def limits():
    # Allocation is 4 GiB; this single-threaded 8 MiB-buffer scanner needs far
    # less. Address-space and output-file limits apply to compiler/child alike.
    resource.setrlimit(resource.RLIMIT_AS, (1 << 30, 1 << 30))
    resource.setrlimit(resource.RLIMIT_FSIZE, (32 << 20, 32 << 20))


def guarded(command, log, seconds):
    # Kill the whole compiler/scanner process group, including /usr/bin/time's
    # child, so a controller timeout cannot leave the payload scan running.
    with subprocess.Popen(command, stdout=log, stderr=subprocess.STDOUT,
                          start_new_session=True, preexec_fn=limits) as child:
        try:
            code = child.wait(timeout=seconds)
        except subprocess.TimeoutExpired:
            os.killpg(child.pid, signal.SIGKILL)
            child.wait()
            raise
        if code:
            raise subprocess.CalledProcessError(code, command)


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('name')
    parser.add_argument('--check', action='store_true')
    args = parser.parse_args()
    if not args.name or any(c not in 'abcdefghijklmnopqrstuvwxyz0123456789-' for c in args.name):
        raise ValueError('invalid run name')
    if not shutil.which('g++') or not Path('/usr/bin/time').is_file():
        raise RuntimeError('compiler/time unavailable')
    if not SUPPORT.is_file() or SUPPORT.stat().st_size != COUNT*8:
        raise RuntimeError('retained support missing/wrong size')
    manifest = json.loads((SUPPORT.parent / 'manifest.json').read_text())
    # Full payload digest is checked by the scanner during its one read pass.
    if SHA not in json.dumps(manifest):
        raise RuntimeError('retained manifest digest mismatch')
    if args.check:
        print('preflight passed: existing support size/manifest and tools; no payload scan')
        return
    output = HERE / 'runs' / args.name
    output.mkdir(parents=True, exist_ok=False)
    scratch = Path(os.environ.get('SLURM_TMPDIR', '/tmp')) / ('depir-owner-' + args.name + '-' + str(os.getpid()))
    scratch.mkdir(exist_ok=False)
    began = time.monotonic()
    source = HERE / 'scan_support.cpp'
    frozen = scratch / source.name
    shutil.copyfile(source, frozen)
    metadata = dict(revision=subprocess.check_output(['git', 'rev-parse', 'HEAD'], cwd=ROOT, text=True).strip(),
                    source_sha256=hashlib.sha256(frozen.read_bytes()).hexdigest(),
                    input=str(SUPPORT.relative_to(ROOT)), input_sha256=SHA,
                    input_bytes=COUNT*8, node=os.uname().nodename,
                    job_id=os.environ.get('SLURM_JOB_ID'),
                    resources='1 HPC node,2 CPU cores,4 GiB requested RAM,no GPU;1 GiB child address-space,32 MiB file cap;720s controller,15min Slurm')
    (output/'metadata.json').write_text(json.dumps(metadata, indent=2)+'\n')
    binary = scratch/'scan_support'
    try:
        with (output/'build.log').open('w') as log:
            guarded(['g++', '-std=c++17', '-O3', str(frozen), '-lcrypto', '-o', str(binary)], log, 60)
        with (output/'progress.log').open('w') as log:
            guarded(['/usr/bin/time', '-v', '-o', str(output/'resources.log'), str(binary),
                     '17', '8', str(COUNT), SHA, str(SUPPORT), str(output/'owners.json')],
                    log, max(1, 720-(time.monotonic()-began)))
        result = json.loads((output/'owners.json').read_text())
        if result['count'] != COUNT or result['sha256'] != SHA:
            raise RuntimeError('scan identity mismatch')
        for ranks, counts in result['owners'].items():
            if len(counts) != int(ranks) or sum(counts) != COUNT:
                raise RuntimeError('owner histogram mismatch')
        metadata['status'] = 'passed'
    finally:
        metadata['controller_seconds'] = time.monotonic()-began
        (output/'metadata.json').write_text(json.dumps(metadata, indent=2)+'\n')
        # Retain scratch and compiler binary for diagnosis/reuse; never alter
        # or remove the original support archive or preprocessing profiles.
        print(json.dumps(dict(results=str(output), scratch=str(scratch), **metadata)), flush=True)


if __name__ == '__main__':
    main()
