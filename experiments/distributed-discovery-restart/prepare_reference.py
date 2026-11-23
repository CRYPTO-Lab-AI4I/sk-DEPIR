#!/usr/bin/env python3
"""Freeze verified s10 native cache hashes without rerunning preprocessing."""
from pathlib import Path
import argparse,hashlib,json,pprint
p=argparse.ArgumentParser(description=__doc__);p.add_argument('cache',type=Path);p.add_argument('header',type=Path);p.add_argument('output',type=Path);a=p.parse_args()
h=json.loads(a.header.read_text());assert (h['Spec']['FieldDegree'],h['Spec']['CurveDegree'],h['K'],h['N'])==(10,5,146641,1048576)
root=Path(__file__).resolve().parents[2];record=json.loads((root/'experiments/gpu-products/runs/products-02/results.json').read_text());assert record['status']=='passed'and record['job_id']=='72839'
for name in ('online.bin','g.bin','support.bin'):
 data=(a.cache/name).read_bytes();assert hashlib.sha256(data).hexdigest()==record['cache_sha256']['cpu-0/'+name]
code=(a.cache/'online.bin').read_bytes();generator=(a.cache/'g.bin').read_bytes();support=(a.cache/'support.bin').read_bytes()
assert len(code)==h['Bytes']and len(generator)==4*(h['N']-h['K'])and len(support)==4*h['K']
indices=[int.from_bytes(support[i:i+4],'little')for i in range(0,len(support),4)];assert indices==sorted(set(indices))and max(indices)<h['N']-1
archive=b''.join(v.to_bytes(8,'little')for v in indices);owners=[0]*8
for v in indices:owners[min(v,v*1024%(h['N']-1))%8]+=1
files={name:dict(name=name,bytes=len(data),sha256=hashlib.sha256(data).hexdigest())for name,data in [('s10-code.bin',code),('s10-generator.bin',generator),('s10-archive.raw',archive)]}
polynomial=code[h['POffset']:h['SpectrumOffset']];assert len(polynomial)==4*(h['K']+1)
stages={'root':dict(prefix=h['K']+1,bytes=len(polynomial),sha256=hashlib.sha256(polynomial).hexdigest())}
previous=1
while previous<h['N']-h['K']:
 previous=min(2*previous,h['N']-h['K']);data=generator[:4*previous];stages[f'newton-{previous}']=dict(prefix=previous,bytes=len(data),sha256=hashlib.sha256(data).hexdigest())
r=dict(provenance='Native CPU cache72839/source715bac2;seals and full CPU/GPU/readback previously passed;header revalidated;not authentication',header=h,samples=None,files=files,stages=stages,root_counts=[v+1 for v in owners])
assert max(r['root_counts'])<=h['N']//8
# Full final inverse prefixes equal the Newton truncations by uniqueness of P^-1.
a.output.write_text('"""Frozen verified native s10 logical reference;see prepare_reference.py."""\nREFERENCE = '+pprint.pformat(r,sort_dicts=True,width=100)+'\n')
