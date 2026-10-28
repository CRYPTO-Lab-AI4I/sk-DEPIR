#!/usr/bin/env python3
"""Collect separate s=9 CPU profiles; timings include profiler overhead."""
import argparse
import datetime
import hashlib
import json
import os
from pathlib import Path
import platform
import subprocess
import tempfile

ROOT = Path(__file__).resolve().parents[2]
OUT = Path(__file__).resolve().parent

def command(args):
    return subprocess.check_output([str(x) for x in args], cwd=ROOT, text=True)

def digest(path):
    return hashlib.sha256(path.read_bytes()).hexdigest()

def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--go', default='go')
    parser.add_argument('--seconds', type=float, default=4)
    args = parser.parse_args()
    if args.seconds <= 0:
        parser.error('--seconds must be positive')
    sources = sorted((ROOT/'encoding/liftrs-go').glob('*.go')) + [OUT/'main.go', OUT/'run.py']
    inputs = sorted((ROOT/'experiments/tablefree-encoding/artifacts').glob('s09-*.json'))
    result = {'timestamp_utc': datetime.datetime.now(datetime.timezone.utc).isoformat(),
              'revision': command(['git', 'rev-parse', 'HEAD']).strip(),
              'working_tree_dirty': bool(command(['git', 'status', '--porcelain']).strip()),
              'environment': {'platform': platform.platform(), 'python': platform.python_version(),
                              'cpu_count': os.cpu_count(), 'go': command([args.go, 'version']).strip(),
                              'gomaxprocs_environment': os.environ.get('GOMAXPROCS', 'default')},
              'minimum_seconds': args.seconds,
              'source_sha256': {str(p.relative_to(ROOT)): digest(p) for p in sources},
              'input_sha256': {str(p.relative_to(ROOT)): digest(p) for p in inputs}, 'profiles': []}
    with tempfile.TemporaryDirectory(prefix='depir-profile-') as directory:
        binary = Path(directory)/'profile'
        command([args.go, 'build', '-o', binary, './experiments/encoder-profile'])
        for stage in ['prepare', 'encode']:
            for method in ['paper', 'cyclic-fft']:
                name = stage+'-'+method
                profile = OUT/(name+'.pprof')
                measured = json.loads(command([binary, '-stage', stage, '-method', method,
                                               '-duration', str(args.seconds)+'s', '-profile', profile]))
                for order in ['flat', 'cum']:
                    text = command([args.go, 'tool', 'pprof', '-top', '-nodecount=35', '-'+order, binary, profile])
                    (OUT/(name+'-'+order+'.txt')).write_text(text)
                measured['profile'] = profile.name
                measured['profile_sha256'] = digest(profile)
                if stage == 'encode':
                    checkpoint = json.loads((ROOT/'experiments/tablefree-encoding/results.json').read_text())
                    # Both encoding outputs must agree with the saved canonical hash.
                    def hashes(value):
                        if isinstance(value, dict):
                            for key, item in value.items():
                                if key == 'output_sha256':
                                    yield item
                                else:
                                    yield from hashes(item)
                        elif isinstance(value, list):
                            for item in value:
                                yield from hashes(item)
                    if measured['sha256'] not in set(hashes(checkpoint)):
                        raise RuntimeError('encoding output differs from checkpoint')
                result['profiles'].append(measured)
    (OUT/'results.json').write_text(json.dumps(result, indent=2)+'\n')

if __name__ == '__main__':
    main()
