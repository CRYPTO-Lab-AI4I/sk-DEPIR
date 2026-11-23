#!/usr/bin/env python3
"""Pin logical outputs from independently verified four-worker command82181."""
from pathlib import Path
import argparse
import hashlib
import json
import pprint
p=argparse.ArgumentParser(description=__doc__);p.add_argument('cache',type=Path);p.add_argument('output',type=Path);a=p.parse_args()
source=a.cache/'command-cold';record=json.loads((source/'s8-publication.json').read_text());assert record['header']['K']==7381 and record['identity_verified']is True
files={}
for item in record['files']:
 data=(source/item['name']).read_bytes();assert len(data)==item['bytes']and hashlib.sha256(data).hexdigest()==item['sha256'];files[item['name']]=item
stages={}
for path in source.glob('s8-*-commit.json'):
 r=json.loads(path.read_text());assert r['ranks']==4
 data=b''.join((source/f's8-rank-{rank}-{r["stage"]}.bin').read_bytes()for rank in range(4));assert len(data)==4*r['prefix']
 stages[r['stage']]=dict(prefix=r['prefix'],bytes=len(data),sha256=hashlib.sha256(data).hexdigest())
assert len(stages)==17
archive=(source/'s8-archive.raw').read_bytes();owners=[0]*8
for i in range(0,len(archive),8):
 v=int.from_bytes(archive[i:i+8],'little');owners[min(v,v*256%65535)%8]+=1
result=dict(provenance='Verified command82181, producing a9fb7c2;canonical/native/pilot comparisons passed;not authentication',header=record['header'],files=files,stages=stages,root_counts=[n+1 for n in owners])
a.output.write_text('"""Frozen logical-output reference;regenerate with prepare_reference.py."""\nREFERENCE = '+pprint.pformat(result,sort_dicts=True,width=100)+'\n')
