#!/usr/bin/env python3
"""Compare prepared paper/cyclic encoders; public artifacts contain no database.

Build once, prepare each code once, then load each complete method artifact in
an independent process and encode three times without repeating preparation.
See report.md for measurement boundaries and limitations.
"""
import argparse
import hashlib
import json
import os
from pathlib import Path
import platform
import statistics
import subprocess
import tempfile
from datetime import datetime, timezone

ROOT = Path(__file__).resolve().parents[2]
HERE = Path(__file__).resolve().parent

def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--go', default=os.environ.get('GO', 'go'))
    args = parser.parse_args()
    artifacts = HERE / 'artifacts'
    artifacts.mkdir(exist_ok=True)
    def invoke(executable, mode, s, t):
        command = [executable, '-mode', mode, '-s', str(s), '-t', str(t),
                   '-artifacts', str(artifacts), '-repeats', '3']
        # wait4 returns each child's own RSS high-water mark. Unlike collecting
        # all children with getrusage, the compiler's RSS cannot contaminate it.
        with tempfile.TemporaryFile() as stdout, tempfile.TemporaryFile() as stderr:
            process = subprocess.Popen(command, cwd=ROOT, stdout=stdout, stderr=stderr)
            _, status, usage = os.wait4(process.pid, 0)
            process.returncode = os.waitstatus_to_exitcode(status)
            stdout.seek(0); stderr.seek(0)
            if process.returncode:
                raise RuntimeError(stderr.read().decode())
            data = json.load(stdout)
        data['process_peak_rss_bytes'] = usage.ru_maxrss * (1 if platform.system() == 'Darwin' else 1024)
        return data
    with tempfile.TemporaryDirectory(prefix='sk-depir-encoder-build-') as tmp:
        executable = str(Path(tmp) / 'compare')
        subprocess.run([args.go, 'build', '-o', executable, './experiments/encoder-comparison'],
                       cwd=ROOT, check=True)
        scenarios = []
        for s, t in [(4, 2), (5, 3), (6, 4), (7, 6), (8, 8)]:
            setup = invoke(executable, 'prepare', s, t)
            methods = [invoke(executable, method, s, t) for method in ['paper', 'cyclic', 'cyclic-fft']]
            assert len({m['output_sha256'] for m in methods}) == 1
            scenarios.append({'preprocessing': setup, 'methods': methods})
            print(f's={s}: equal outputs; median encode ' + ', '.join(
                f"{m['method']} {statistics.median(m['encode_ns'])/1e6:.2f} ms" for m in methods), flush=True)
    sources = list((ROOT / 'encoding/liftrs-go').glob('*.go')) + [HERE/'main.go', HERE/'run.py', ROOT/'go.mod', ROOT/'paper.pdf']
    result = {'timestamp_utc': datetime.now(timezone.utc).isoformat(),
              'base_revision': subprocess.check_output(['git','rev-parse','HEAD'], cwd=ROOT, text=True).strip(),
              'working_tree_dirty': bool(subprocess.check_output(['git','status','--porcelain'], cwd=ROOT, text=True).strip()),
              'platform': platform.platform(), 'processor': platform.processor(),
              'cpu_count': os.cpu_count(), 'gomaxprocs': os.environ.get('GOMAXPROCS','default'),
              'python_version': platform.python_version(),
              'source_sha256': {str(p.relative_to(ROOT)):hashlib.sha256(p.read_bytes()).hexdigest() for p in sources},
              'artifact_sha256': {p.name:hashlib.sha256(p.read_bytes()).hexdigest() for p in sorted(artifacts.glob('*.json'))},
              'scenarios': scenarios}
    (HERE/'results.json').write_text(json.dumps(result, indent=2)+'\n')
    lines = ['# Prepared encoder comparison', '',
             'Measured CPU times; same codeword for both methods. Median of three encodings, with preparation excluded. P=paper, C=cyclic Karatsuba, F=cyclic cached FFT.', '',
             '| s / t | N | K | Discovery ms | Prepare P / C / F ms | Artifact P / C / F KiB | Load P / C / F ms | Encode P / C / F ms | Peak RSS P / C / F MiB |',
             '|---|---:|---:|---:|---:|---:|---:|---:|---:|']
    for case in scenarios:
        p=case['preprocessing']; a,b,c=case['methods']
        pair=lambda key,scale: f"{a[key]/scale:.2f} / {b[key]/scale:.2f} / {c[key]/scale:.2f}"
        lines.append(f"| {p['s']} / {p['t']} | {p['n']:,} | {p['k']:,} | {p['discovery_ns']/1e6:.2f} | {p['paper_prepare_ns']/1e6:.2f} / {p['cyclic_prepare_ns']/1e6:.2f} / {p['cyclic_fft_prepare_ns']/1e6:.2f} | {p['paper_artifact_bytes']/1024:.1f} / {p['cyclic_artifact_bytes']/1024:.1f} / {p['cyclic_fft_artifact_bytes']/1024:.1f} | {pair('artifact_load_ns',1e6)} | {statistics.median(a['encode_ns'])/1e6:.2f} / {statistics.median(b['encode_ns'])/1e6:.2f} / {statistics.median(c['encode_ns'])/1e6:.2f} | {pair('process_peak_rss_bytes',2**20)} |")
    lines += ['', 'RSS is the complete method process peak, including loading and I/O; it is not encoding-only live heap. See results.json for allocations, artifact reads, sync writes and cached readbacks.']
    (HERE/'table.md').write_text('\n'.join(lines)+'\n')

if __name__ == '__main__':
    main()
