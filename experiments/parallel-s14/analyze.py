#!/usr/bin/env python3
"""Verify retained s14 evidence and summarize measured first-use phases."""
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
 name=sys.argv[1] if len(sys.argv)>1 else 'initial';out=HERE/name;r=json.loads((out/'results.json').read_text());assert r['status']=='passed'
 verify_sources(ROOT,r['source_sha256'],r['revision']);assert sha(ROOT/r['binary']['path'])==r['binary']['sha256']
 cache=ROOT/r['cache_directory']
 for p,want in r['cache_sha256'].items():assert sha(cache/p)==want,p
 for p,want in r['evidence_sha256'].items():assert sha(out/p)==want,p
 phases={x['phase']:x for x in r['phases']};code=phases.get('discovery',phases['load-online']);k,n=code['k'],code['n'];s=6 if r['smoke'] else 14
 assert n==1<<(2*s)
 enc=[x for x in r['phases'] if x['phase'].startswith('encode-')];checks=[x for x in r['phases'] if x['phase'].startswith('verify-') and x['phase']!='verify-packed']
 assert len(enc)==len(checks) and all(x['systematic_checked']==k and x['curves_checked']==16 and x['repeat_match'] for x in checks)
 assert len({x['canonical_sha256'] for x in checks})==1
 assert all(x['native_stages']['calls']==5 for x in enc)
 assert phases['pack-output']['packed_bytes']==(n*s+7)//8
 if 'verify-packed' in phases:assert phases['verify-packed']['readback_match']
 else:assert r['smoke'], 'large run missing packed readback'
 meta=json.loads((cache/'online-metadata.json').read_text());assert meta['k']==k and meta['n']==n and meta['identity_verified']
 med=statistics.median(x['elapsed_ns']/1e9 for x in enc)
 times={name:x['elapsed_ns']/1e9 for name,x in phases.items()};first=sum(times.get(x,0) for x in ['executor-setup','discovery','preparation','online-setup'])+med
 table=['| Phase | Measured time |','|---|---:|']
 for x in r['phases']:table.append(f'| {x["phase"]} | {x["elapsed_ns"]/1e9:.3f} s |')
 table+=['',f'Exact K={k:,}; N={n:,}; code rate K/N={k/n:.9f}.',f'Median profiled encoding: {med:.3f} s.',f'Measured discovery/preparation + executor/online setup + one median encoding: {first:.3f} s (excludes serialization/loading/verification; reused or skipped stages contribute zero).',f'Whole-process peak RSS: {r["process_peak_rss_bytes"]/2**30:.3f} GiB.',f'Packed output: {phases["pack-output"]["packed_bytes"]:,} bytes.',f'Canonical SHA-256: `{checks[0]["canonical_sha256"]}`.']
 (out/'table.md').write_text('\n'.join(table)+'\n');print('\n'.join(table))
if __name__=='__main__':main()
