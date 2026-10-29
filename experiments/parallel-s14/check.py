#!/usr/bin/env python3
"""Verify new experimental s14 path and both isolated repository build paths."""
import argparse,json,os,shutil,subprocess,tempfile,time
from pathlib import Path
ROOT=Path(__file__).resolve().parents[2];HERE=Path(__file__).resolve().parent
p=argparse.ArgumentParser(description=__doc__);p.add_argument('--go',default='go');a=p.parse_args();out=HERE/'final-validation';out.mkdir(exist_ok=False);r={'status':'running','checks':[]}
with (out/'checks.log').open('w') as log:
 def run(cmd,cwd=ROOT,env=None):
  log.write('$ '+' '.join(cmd)+'\n');log.flush();start=time.monotonic();x=subprocess.run(cmd,cwd=cwd,env=env,stdout=log,stderr=subprocess.STDOUT,timeout=600);r['checks'].append({'command':cmd,'isolated':cwd!=ROOT,'exit_code':x.returncode,'wall_s':time.monotonic()-start});(out/'verification.json').write_text(json.dumps(r,indent=2)+'\n');assert x.returncode==0
 run([a.go,'test','-race','-tags','liftrs_cpp','./encoding/liftrs-shared','-run','TestScalingPreparationCPP|TestScalingCPPField14|TestCPPMatrix|TestScalingOnline','-count=1'])
 run([a.go,'vet','-tags','liftrs_cpp','./encoding/...','./pirconfig','./planner'])
 with tempfile.TemporaryDirectory(prefix='depir-s14-full-') as tmp:
  copy=Path(tmp)/'repo';shutil.copytree(ROOT,copy,ignore=shutil.ignore_patterns('.git','.cache','__pycache__','.DS_Store'))
  run(['make','-C','encoding/rme_legacy','rme'],cwd=copy)
  run([a.go,'test','./...'],cwd=copy,env=dict(os.environ,CGO_ENABLED='0'))
  run([a.go,'test','-tags','liftrs_cpp','./...'],cwd=copy)
r['status']='passed';(out/'verification.json').write_text(json.dumps(r,indent=2)+'\n');print('s14 final validation PASS')
