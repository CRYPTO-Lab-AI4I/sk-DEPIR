#!/usr/bin/env python3
"""Run opt-in step-3 probes, retain preprocessing/profiles, verify both build paths."""
import argparse,hashlib,json,os,platform,shutil,subprocess,tempfile,time
from pathlib import Path
ROOT=Path(__file__).resolve().parents[2];HERE=Path(__file__).resolve().parent

def sha(path):return hashlib.sha256(path.read_bytes()).hexdigest()
def main():
 p=argparse.ArgumentParser(description=__doc__);p.add_argument('--go',default='go');p.add_argument('--run-name',default='initial');p.add_argument('--s13',action='store_true');a=p.parse_args()
 if not a.run_name or any(c not in 'abcdefghijklmnopqrstuvwxyz0123456789-_' for c in a.run_name):raise SystemExit('invalid run name')
 out=HERE/a.run_name;out.mkdir(exist_ok=False)
 sources=sorted((ROOT/'encoding/liftrs-shared').rglob('*.go'))+sorted((ROOT/'encoding/liftrs-go').glob('*.go'))+sorted((ROOT/'encoding/liftrs-cpp').glob('*.*'))+[HERE/'run.py',ROOT/'go.mod',ROOT/'go.sum']
 r={'revision':subprocess.check_output(['git','rev-parse','HEAD'],cwd=ROOT,text=True).strip(),'source_sha256':{str(p.relative_to(ROOT)):sha(p) for p in sources},'status':'running','checks':[],'environment':{'platform':platform.platform(),'cpu':'Apple M5 Pro, 48 GiB (same host as s13 baseline)','go':subprocess.check_output([a.go,'version'],text=True).strip()},'policy':{'profiling':True,'go_soft_memory_limit':'2GiB','timeout_seconds':600,'workloads':'one s13 cached native workspace encode' if a.s13 else 'setup s8..15; preparation/native workspace s8,9; matrix and route s4','large_baseline_rerun':False}}
 identity=hashlib.sha256(json.dumps(r['source_sha256'],sort_keys=True).encode()).hexdigest()[:16]
 cache=ROOT/'.cache/liftrs/scaling-step3'/identity;cache.mkdir(parents=True,exist_ok=True)
 r['cache_directory']=str(cache.relative_to(ROOT));(cache/'provenance.json').write_text(json.dumps({'revision':r['revision'],'source_sha256':r['source_sha256']},indent=2)+'\n')
 def save():(out/'results.json').write_text(json.dumps(r,indent=2)+'\n')
 save()
 with (out/'checks.log').open('w') as log:
  def run(cmd,cwd=ROOT,env=None):
   log.write('$ '+' '.join(cmd)+'\n');log.flush();start=time.monotonic();x=subprocess.run(cmd,cwd=cwd,env=env,stdout=log,stderr=subprocess.STDOUT,timeout=600)
   r['checks'].append({'command':cmd,'isolated_copy':cwd!=ROOT,'exit_code':x.returncode,'wall_s':time.monotonic()-start});save()
   if x.returncode:raise RuntimeError('check failed; see checks.log')
  binary=cache/'probe';run([a.go,'test','-c','-tags','liftrs_cpp','-o',str(binary),'./encoding/liftrs-shared']);r['binary']={'path':str(binary.relative_to(ROOT)),'sha256':sha(binary)};save()
  env=dict(os.environ,PIR_SCALING_PROBE=str(out),PIR_SCALING_CACHE=str(cache),GOMEMLIMIT='2GiB')
  if a.s13:env['PIR_SCALING_S13_CACHE']=str(ROOT/'.cache/liftrs/s13-m2-t5/cyclic.json');env['GOMEMLIMIT']='9GiB'
  run([str(binary),'-test.run','^TestScalingS13$' if a.s13 else '^TestScalingProbe$','-test.v'],env=env)
  run([a.go,'test','-tags','liftrs_cpp','./encoding/liftrs-shared','-run','TestScaling','-count=1'])
  run([a.go,'test','./encoding/liftrs-shared','-run','TestScaling','-count=1'],env=dict(os.environ,CGO_ENABLED='0'))
  run([a.go,'vet','-tags','liftrs_cpp','./encoding/...'])
  with tempfile.TemporaryDirectory(prefix='depir-scaling-checks-') as tmp:
   copy=Path(tmp)/'repo';shutil.copytree(ROOT,copy,ignore=shutil.ignore_patterns('.git','.cache','__pycache__','.DS_Store'))
   run(['make','-C','encoding/rme_legacy','rme'],cwd=copy)
   run([a.go,'test','./...'],cwd=copy)
   run([a.go,'test','-tags','liftrs_cpp','./...'],cwd=copy)
  for profile in sorted(out.glob('*.pprof')):
   for suffix,flags in [('flat',['-top']),('cum',['-top','-cum'])]:
    with (out/(profile.stem+'-'+suffix+'.txt')).open('w') as stream:subprocess.run([a.go,'tool','pprof',*flags,str(binary),str(profile)],stdout=stream,stderr=subprocess.STDOUT,check=True)
 r['status']='passed';r['cache_sha256']={str(p.relative_to(cache)):sha(p) for p in cache.rglob('*') if p.is_file()};r['evidence_sha256']={p.name:sha(p) for p in out.iterdir() if p.is_file() and p.name!='results.json'};save();print('Scaling prototypes and full validation: PASS')
if __name__=='__main__':main()
