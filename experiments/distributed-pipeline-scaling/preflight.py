#!/usr/bin/env python3
"""CPU-only export regression: historical exact bytes and larger native admission."""
from pathlib import Path
import argparse
import hashlib
import json
import os
import runpy
import subprocess
HERE=Path(__file__).resolve().parent
p=argparse.ArgumentParser(description=__doc__)
p.add_argument('binary',type=Path);p.add_argument('destination',type=Path)
p.add_argument('--root',type=Path,default=HERE.parents[1]);a=p.parse_args()
root=a.root.resolve();dest=a.destination.resolve();dest.mkdir(exist_ok=False)
binary=a.binary.resolve();assert binary.is_file()and os.access(binary,os.X_OK)
commands=[]
for label,cases,curve in (('legacy','4,6','2'),('native','8','5')):
    directory=dest/label;directory.mkdir()
    env=dict(os.environ,PIR_PIPELINE_CASES=cases,PIR_PIPELINE_CURVE=curve,PIR_PIPELINE_SMALL='0',PIR_PIPELINE_ROOT_EXPORT='1',PIR_PIPELINE_REFERENCE_EXPORT=str(directory),PIR_PIPELINE_REFERENCE_FROM=str(directory))
    command=[str(binary),'-test.run=^TestExportPilotReferences$|^TestPilotHierarchyReferences$','-test.v','-test.timeout=20s']
    result=subprocess.run(command,env=env,text=True,stdout=subprocess.PIPE,stderr=subprocess.STDOUT,timeout=25)
    (dest/(label+'.log')).write_text(result.stdout);commands.append(dict(argv=command,cases=cases,curve=curve,exit_code=result.returncode))
    if result.returncode:raise RuntimeError(label+' export/admission failed;see '+str(dest/(label+'.log')))
verified=0
for name in ('distributed-coordinator-stream/reference-inputs.json','distributed-owned-roots/root-references.json'):
    manifest=json.loads((root/'experiments'/name).read_text())
    for filename,record in manifest['files'].items():
        if not filename.startswith(('s4-','s6-')):continue
        data=(dest/'legacy'/filename).read_bytes()
        assert len(data)==record['bytes']and hashlib.sha256(data).hexdigest()==record['sha256'],filename
        verified+=len(data)
H=runpy.run_path(str(HERE/'run.py'));manifest=H['reference_manifest'](dest/'native',root)
info=json.loads((dest/'native/s8-reference.json').read_text())['parameters']
report=dict(status='passed',commands=commands,historical_bytes_verified=verified,native_parameters=info,root_references=manifest)
(dest/'checks.json').write_text(json.dumps(report,indent=2)+'\n')
print(json.dumps(report,indent=2))
