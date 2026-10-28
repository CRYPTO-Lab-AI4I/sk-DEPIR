#!/usr/bin/env python3
"""Verify retained step-3 manifests and regenerate concise diagnostic tables."""
import hashlib,json,sys
from pathlib import Path
ROOT=Path(__file__).resolve().parents[2];HERE=Path(__file__).resolve().parent
sys.path.insert(0,str(HERE.parent))
from source_evidence import verify_sources

def sha(p):return hashlib.sha256(p.read_bytes()).hexdigest()
def main():
 name=sys.argv[1] if len(sys.argv)>1 else 'initial';out=HERE/name;r=json.loads((out/'results.json').read_text());assert r['status']=='passed'
 verify_sources(ROOT,r['source_sha256'],r['revision'])
 assert sha(ROOT/r['binary']['path'])==r['binary']['sha256']
 for p,digest in r['evidence_sha256'].items():assert sha(out/p)==digest,p
 for p,digest in r['cache_sha256'].items():assert sha(ROOT/r['cache_directory']/p)==digest,p
 if (out/'s13-phases.json').exists():
  d={x['phase']:x for x in json.loads((out/'s13-phases.json').read_text())}
  verified=d['s13-verify'];assert verified['matches_original_baseline'] and verified['systematic_positions']==11926435
  assert verified['canonical_sha256']=='6120f0388f6035dd480e391b26ef93c04ed368a77f2f93c02f6d4f56b99a7b79'
  rows=['| Stage | Profiled time (s) | Allocated MiB | Heap at stage end (MiB) |','|---|---:|---:|---:|']
  for key,x in d.items():rows.append(f'| {key} | {x["elapsed_ns"]/1e9:.6f} | {x["allocated_bytes"]/2**20:.2f} | {x["heap_after_bytes"]/2**20:.2f} |')
 else:
  d={x['phase']:x for x in json.loads((out/'phases.json').read_text())}
  rows=['| s | Dense setup allocated MiB | Sparse setup allocated MiB | Sparse description KiB | Same description |','|---:|---:|---:|---:|---|']
  for s in range(8,16):
   dense=d.get(f'dense-setup-s{s}');sparse=d[f'sparse-setup-s{s}'];assert sparse['dense_comparison']==(s<=13)
   rows.append(f'| {s} | '+(f'{dense["allocated_bytes"]/2**20:.2f}' if dense else 'not run')+f' | {sparse["allocated_bytes"]/2**20:.2f} | {sparse["description_bytes"]/1024:.2f} | '+('yes' if dense else 'not compared; structural/Horner tests')+' |')
  rows+=['','| s | Baseline preparation (s) | Prototype preparation (s) | Baseline allocated MiB | Prototype allocated MiB |','|---:|---:|---:|---:|---:|']
  for s in [8,9]:
   a,b=d[f'baseline-preparation-s{s}'],d[f'prototype-preparation-s{s}'];assert b['pg_spectrum_equal']
   rows.append(f'| {s} | {a["elapsed_ns"]/1e9:.3f} | {b["elapsed_ns"]/1e9:.3f} | {a["allocated_bytes"]/2**20:.2f} | {b["allocated_bytes"]/2**20:.2f} |')
  for case in d['matrix-disk-s4']['cases']:assert case['roundtrip'] and case['file_bytes']==1024 and case['tile_payload_bytes']<=256
  for case in d['routing-s4']['cases']:assert case['canonical_equal'] and case['logical_bucket_and_output_bytes']==5104
 (out/'table.md').write_text('\n'.join(rows)+'\n');print('\n'.join(rows));print('Source/binary/cache/profile/evidence checks: PASS')
if __name__=='__main__':main()
