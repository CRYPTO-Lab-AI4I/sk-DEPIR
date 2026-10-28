#!/usr/bin/env python3
"""Check s=13 provenance and generate a stage table, including partial runs."""
import hashlib,json,statistics,runpy
from pathlib import Path
ROOT=Path(__file__).resolve().parents[2];HERE=Path(__file__).resolve().parent
sha=lambda p:hashlib.sha256(p.read_bytes()).hexdigest()
r=json.loads((HERE/'results.json').read_text())
runpy.run_path(str(ROOT/'experiments/source_evidence.py'))['verify_sources'](ROOT,r['source_sha256'],r['revision'])
for name in ['events','stderr','progress']:
 if name in r:assert sha(HERE/r[name])==r[name+'_sha256'],name
lines=['# Cyclic encoding at s=13','','Status: '+r['status']+'. Times below are measured, not forecasts.','','| Stage | Time | Sampled peak heap | Sampled peak Go runtime |','|---|---:|---:|---:|']
for name,p in r.get('phases',{}).items():
 lines.append(f'| {name} | {p["elapsed_ns"]/1e9:.2f} s | {p["sampled_peak_heap_bytes"]/2**30:.3f} GiB | {p["sampled_peak_go_sys_bytes"]/2**30:.3f} GiB |')
if r['status']=='completed':
 code=r['code'];v=r['verification'];m=r['measurement']
 assert (code['s'],code['m'],code['t'],code['n'])==(13,2,5,67108864)
 assert v['k']==code['k'] and v['systematic_positions_checked']==code['k'] and v['curve_checks']==16
 assert v['packed_bytes']==109051904 and len(v['canonical_sha256'])==len(v['packed_sha256'])==64
 enc=[p['elapsed_ns']/1e9 for name,p in r['phases'].items() if name.startswith('encode_')]
 assert len(enc)==r['policy']['encoding_repeats'] and m['encoding_s']==statistics.median(enc)
 first=m['discovery_s']+m['preparation_s']+m['encoding_s'];assert first==m['first_use_s']
 lines+=['',f'Exact K={code["k"]:,}, N={code["n"]:,}; code rate={code["k"]/code["n"]:.8f}, N/K={code["n"]/code["k"]:.4f}.',f'First use: {first/60:.2f} min. Median encoding: {m["encoding_s"]:.2f} s.',f'Whole-case wall time: {r["wall_seconds"]/60:.2f} min; process peak RSS: {r["process_peak_rss_bytes"]/2**30:.3f} GiB.',f'Packed output: {v["packed_bytes"]:,} bytes (104 MiB).']
(HERE/'table.md').write_text('\n'.join(lines)+'\n')
print('\n'.join(lines))
