#!/usr/bin/env python3
"""Verify recorded sources/cache/evidence and summarize profiled matrix encoding."""
import hashlib,json,statistics,sys
from pathlib import Path
ROOT=Path(__file__).resolve().parents[2];HERE=Path(__file__).resolve().parent
sys.path.insert(0,str(ROOT/'experiments'))
from source_evidence import verify_sources

def sha(p):
 h=hashlib.sha256()
 with p.open('rb') as f:
  for b in iter(lambda:f.read(1<<20),b''):h.update(b)
 return h.hexdigest()
def main():
 name=sys.argv[1] if len(sys.argv)>1 else 's13-prepared';out=HERE/name;r=json.loads((out/'results.json').read_text());assert r['status']=='passed'
 verify_sources(ROOT,r['source_sha256'],r['revision'])
 assert sha(ROOT/r['binary']['path'])==r['binary']['sha256']
 cache=ROOT/r['cache_directory']
 for p,want in r['cache_sha256'].items():assert sha(cache/p)==want,p
 for p,want in r['evidence_sha256'].items():assert sha(out/p)==want,p
 if 'input_cache' in r:assert sha(ROOT/r['input_cache']['path'])==r['input_cache']['sha256']
 phases=r['phases'];rows=[]
 for w in sorted({x['workers'] for x in phases if x['phase'].startswith('encode-')}):
  enc=[x for x in phases if x['phase'].startswith('encode-') and x['workers']==w]
  for x in enc:
   verify=next(p for p in phases if p['phase']==x['phase'].replace('encode-','verify-'))
   assert verify['full_hash_match']
   assert x['native_stages']['calls']==5
   assert sum(x['native_stages'][k] for k in ['taylor_ns','columns_transposes_ns','rows_ns'])<=x['elapsed_ns']
  med=lambda f:statistics.median(f(x)/1e9 for x in enc)
  seconds=med(lambda x:x['elapsed_ns']);times=[med(lambda x:x['native_stages'][k]) for k in ['taylor_ns','columns_transposes_ns','rows_ns']]
  rows.append((w,seconds,*times))
 baseline=153.409
 table=['| Workers | Encoding median | Taylor | Columns + transposes | Rows | Historical baseline / median |','|---:|---:|---:|---:|---:|---:|']
 for w,seconds,t,c,rows_s in rows:
  speed=f'{baseline/seconds:.2f}×' if 'input_cache' in r else '—'
  table.append(f'| {w} | {seconds:.3f} s | {t:.3f} s | {c:.3f} s | {rows_s:.3f} s | {speed} |')
 table+=['',f'Whole-process peak RSS: {r["process_peak_rss_bytes"]/2**20:.2f} MiB. Stage columns are separate medians and need not sum to the encoding median. Historical speedup is meaningful only for the s=13 fixture; small smoke timings are correctness evidence.']
 (out/'table.md').write_text('\n'.join(table)+'\n');print('\n'.join(table))
if __name__=='__main__':main()
