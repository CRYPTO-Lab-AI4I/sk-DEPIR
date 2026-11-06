#!/usr/bin/env python3
"""Measure modest allocation replicas; no s=15 encoding/preparation is run."""
import argparse
import hashlib
import json
import os
import platform
import subprocess
from pathlib import Path

ROOT=Path(__file__).resolve().parents[2]
HERE=Path(__file__).resolve().parent
p=argparse.ArgumentParser(description=__doc__)
p.add_argument('--go',default='go')
p.add_argument('--run-name',default='replicas-final')
a=p.parse_args()
out=HERE/a.run_name;out.mkdir(exist_ok=False)
configs=json.loads((ROOT/'pirconfig/configs_catalogue.json').read_text())['configurations']
c=next(x for x in configs if x['code']['s']==15)
k,n=c['capacity']['k'],c['capacity']['n']
sources=[HERE/'run.py',HERE/'probe.go',ROOT/'go.mod',ROOT/'pirconfig/configs_catalogue.json']
manifest={str(p.relative_to(ROOT)):hashlib.sha256(p.read_bytes()).hexdigest() for p in sources}
for path,digest in manifest.items():
 if not (ROOT/'experiments/source-snapshots'/(digest+'.txt')).exists():
  original=subprocess.run(['git','show','HEAD:'+path],cwd=ROOT,capture_output=True)
  if original.returncode or hashlib.sha256(original.stdout).hexdigest()!=digest:
   (ROOT/'experiments/source-snapshots'/(digest+'.txt')).write_bytes((ROOT/path).read_bytes())
identity=hashlib.sha256(json.dumps(manifest,sort_keys=True).encode()).hexdigest()[:16]
cache=ROOT/'.cache/liftrs/preparation-memory'/identity;cache.mkdir(parents=True,exist_ok=True)
binary=cache/'probe'
record={'revision':subprocess.check_output(['git','rev-parse','HEAD'],cwd=ROOT,text=True).strip(),'source_sha256':manifest,'environment':{'platform':platform.platform(),'go':subprocess.check_output([a.go,'version'],text=True).strip(),'host_reference':'experiments/encoding-timing/environment-end.json'},'kind':'allocation_replica_measurements_not_full_preparation','policy':{'scale':64,'max_entries':10000000,'go_soft_memory_bytes':1<<30,'profiling':'CPU and post-GC heap for every replica'},'cases':[],'status':'running'}
def save():(out/'results.json').write_text(json.dumps(record,indent=2)+'\n')
save()
with (out/'build.log').open('w') as log:
 subprocess.run([a.go,'build','-o',str(binary),'./experiments/preparation-memory'],cwd=ROOT,stdout=log,stderr=subprocess.STDOUT,check=True)
record['binary']={'path':str(binary.relative_to(ROOT)),'sha256':hashlib.sha256(binary.read_bytes()).hexdigest()}
# Two map densities expose capacity/occupancy variation; factor/bitset replicas
# preserve the same N/K scaling. Each process is independent and below 1 GiB.
cases=[('map',k//64),('map',(k//64)*3//4),('factors',(k+32767)//128),('bitset',k//64)]
for index,(kind,count) in enumerate(cases):
 folder=out/(str(index)+'-'+kind)
 command=[str(binary),'-kind',kind,'-count',str(count),'-domain',str(n//64),'-out',str(folder)]
 with (out/(str(index)+'-events.log')).open('w') as log:
  process=subprocess.Popen(command,cwd=ROOT,stdout=log,stderr=subprocess.STDOUT)
  _,status,usage=os.wait4(process.pid,0)
  code=os.waitstatus_to_exitcode(status);assert code==0
 measurement=json.loads((folder/'measurement.json').read_text())
 measurement.update(command=command,process_peak_rss_bytes=usage.ru_maxrss*(1 if platform.system()=='Darwin' else 1024))
 record['cases'].append(measurement);save()
 for profile,metric in [('cpu.pprof','cpu'),('heap.pprof','inuse_space')]:
  with (folder/(metric+'-flat.txt')).open('w') as log:
   subprocess.run([a.go,'tool','pprof','-top','-sample_index='+metric,str(binary),str(folder/profile)],stdout=log,stderr=subprocess.STDOUT,check=True)
record['status']='passed';record['evidence_sha256']={str(p.relative_to(out)):hashlib.sha256(p.read_bytes()).hexdigest() for p in out.rglob('*') if p.is_file() and p.name!='results.json'};save()
print('Allocation replicas PASS')
