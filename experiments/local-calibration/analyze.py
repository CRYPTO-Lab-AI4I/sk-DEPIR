#!/usr/bin/env python3
"""Verify local calibration evidence and summarize medians without extrapolating a full encoder."""
import hashlib,json,statistics,sys
from pathlib import Path
ROOT=Path(__file__).resolve().parents[2];HERE=Path(__file__).resolve().parent
sys.path.insert(0,str(HERE.parent))
from source_evidence import verify_sources

def sha(p):
 h=hashlib.sha256()
 with p.open('rb') as stream:
  for b in iter(lambda:stream.read(1<<20),b''):h.update(b)
 return h.hexdigest()
def verify(out):
 r=json.loads((out/'results.json').read_text());assert r['status']=='passed';verify_sources(ROOT,r['source_sha256'],r['revision']);assert sha(ROOT/r['binary']['path'])==r['binary']['sha256']
 for name,want in r['cache_sha256'].items():assert sha(ROOT/r['cache_directory']/name)==want,name
 for name,want in r['evidence_sha256'].items():assert sha(out/name)==want,name
 return r

def main():
 out=HERE/(sys.argv[1] if len(sys.argv)>1 else 'initial');r=verify(out)
 if out.name=='resources':
  rows=['| Fresh process | Main phase (s) | Whole-process peak RSS (MiB) |','|---|---:|---:|']
  for name,c in r['cases'].items():rows.append(f'| {name} | {c["phases"][0]["elapsed_ns"]/1e9:.3f} | {c["peak_rss_bytes"]/2**20:.2f} |')
  assert r['cases']['baseline-s10']['phases'][-1]['prepared_sha256']==r['cases']['prototype-s10']['phases'][-1]['prepared_sha256']
 else:
  events=[e for e in r['events'] if 'phase'in e];assert r['events'][-1]['verified']
  med=lambda xs:statistics.median(e['elapsed_ns']/1e9 for e in xs)
  cpu={w:[e for e in events if e['phase'].startswith(f'cpu-w{w}-')] for w in [1,2,4,8,16]};baseline=med(cpu[1])
  rows=['| CPU workers | Median batch (s) | Speedup vs 1 |','|---:|---:|---:|']
  for w,cases in cpu.items():
   if cases:rows.append(f'| {w} | {med(cases):.3f} | {baseline/med(cases):.2f}× |')
  if r.get('small_tiles'):
   (out/'table.md').write_text('\n'.join(rows)+'\n');print('\n'.join(rows));print('Source/cache/binary/evidence checks: PASS');return
  rows+=['','| Transfer block | Read MiB/s | Write MiB/s (excludes fsync) | Median fsync (ms) |','|---:|---:|---:|---:|']
  for block in [4096,64<<10,1<<20,4<<20]:
   read=[e for e in events if e['phase'].startswith(f'read-b{block}-')];write=[e for e in events if e['phase'].startswith(f'write-b{block}-')];sync=[e for e in events if e['phase'].startswith(f'sync-b{block}-')]
   rows.append(f'| {block} B | {read[0]["bytes"]/2**20/med(read):.1f} | {write[0]["bytes"]/2**20/med(write):.1f} | {med(sync)*1000:.2f} |')
  rows+=['','| Pipeline slots | Median read/compute/write (s) | Buffer payload (MiB) |','|---:|---:|---:|']
  for slots in [1,2,4]:
   cases=[e for e in events if e['phase'].startswith(f'pipeline-slots{slots}-')];rows.append(f'| {slots} | {med(cases):.3f} | {cases[0]["buffer_payload_bytes"]/2**20:.1f} |')
 (out/'table.md').write_text('\n'.join(rows)+'\n');print('\n'.join(rows));print('Source/cache/binary/evidence checks: PASS')
if __name__=='__main__':main()
