#!/usr/bin/env python3
"""Verify recorded language-comparison evidence and regenerate the result table."""
import hashlib,json,runpy,statistics
from pathlib import Path
ROOT=Path(__file__).resolve().parents[2];HERE=Path(__file__).resolve().parent

def sha(path):
 h=hashlib.sha256()
 with path.open('rb') as f:
  for block in iter(lambda:f.read(1<<20),b''):h.update(block)
 return h.hexdigest()
r=json.loads((HERE/'comparison/results.json').read_text());assert r['status']=='completed'
verify=runpy.run_path(str(ROOT/'experiments/source_evidence.py'))['verify_sources']
verify(ROOT,r['source_sha256'],r['revision'])
cache=ROOT/'.cache/liftrs/language-s13'
for name,h in r['fixture_sha256'].items():assert sha(cache/name)==h,name
for name,h in r['binary_sha256'].items():assert sha(ROOT/name)==h,name
lines=['# Go/C++ kernel and transform comparison','','Three repetitions, medians. Kernel times cover 134,217,728 element products per call; FFT times cover one forward transform. Allocation/input I/O/validation are excluded.','','| Work | Elements per batch | Go | C++ | C++ speedup |','|---|---:|---:|---:|---:|']
for c in r['cases']:
 for language,data in c['implementations'].items():
  path=ROOT/data['output_directory']/'output.bin';assert sha(path)==data['sha256'],c['label']
 assert c['implementations']['go']['sha256']==c['implementations']['cpp']['sha256']
 a,b=c['implementations']['go'],c['implementations']['cpp'];assert a['table_bytes']==b['table_bytes'] and a['passes']==b['passes']
 label=c['mode']+' '+c['kind']+' '+c['pattern'];lines.append(f'| {label} | {c["size"]:,} | {a["median_ns"]/1e9:.4f} s | {b["median_ns"]/1e9:.4f} s | {c["cpp_speedup"]:.2f}× |')
(HERE/'table.md').write_text('\n'.join(lines)+'\n');print('\n'.join(lines))
