#!/usr/bin/env python3
"""Compute s17 format/array models from a trusted exact-dimension record, not peaks."""
import argparse,hashlib,json
from pathlib import Path

def model(k):
 n=1<<34
 if not 0<k<=n//2:raise ValueError('current s17 code requires0<K<=N/2')
 files={'support':8*k,'root':4*(k+1),'canonical_code':53+4*(k+1)+8*n,'generator':4*(n-k)}
 return {'s':17,'n':n,'k':k,'file_bytes':files,'retained_file_bytes':sum(files.values()),'retained_bound_bytes':18*n+61,
 'device_product_array_bytes':3*8*n,'discovery_principal_host_bytes':16*k+8*n+2*(n//8),
 'final_retained_principal_host_bytes':8*k+12*n+4,'peak_host_rss_bytes':None,'full_latency_seconds':None,
 'limitations':'Array/file models only; extra product/Newton temporaries, GC retention, runtime, driver, filesystem cache, staging, logs/profiles, partial stages and headroom additional. Not admission.'}

def main():
 p=argparse.ArgumentParser(description=__doc__);p.add_argument('dimension_manifest',nargs='?');p.add_argument('--check',action='store_true');args=p.parse_args()
 if args.check:
  n=1<<34
  for k in(1,918672421,n//2):
   r=model(k);assert r['retained_file_bytes']==12*k+12*n+61 and r['retained_file_bytes']<=r['retained_bound_bytes']
  assert model(n//2)['retained_bound_bytes']==288*(1<<30)+61
  print('file model and strict288GiB+61byte bound checked');return
 if not args.dimension_manifest:p.error('trusted completed dimension manifest required')
 path=Path(args.dimension_manifest);raw=path.read_bytes();record=json.loads(raw)
 if record['version']!=1 or record['stage']!='dimension' or record['spec']!={'FieldDegree':17,'Dimension':2,'CurveDegree':5}or record['dimension']['n']!=1<<34:raise ValueError('wrong dimension record')
 result=model(record['dimension']['k']);result['dimension_manifest_sha256']=hashlib.sha256(raw).hexdigest();result['count_producer']=record['producer']
 print(json.dumps(result,indent=2))
if __name__=='__main__':main()
