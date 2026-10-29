#!/usr/bin/env python3
"""Fresh-process RSS calibration for preparation and minimal online loading."""
import argparse,hashlib,json,os,platform,subprocess,time
from pathlib import Path
ROOT=Path(__file__).resolve().parents[2];HERE=Path(__file__).resolve().parent

def sha(p):
 h=hashlib.sha256()
 with p.open('rb') as stream:
  for b in iter(lambda:stream.read(1<<20),b''):h.update(b)
 return h.hexdigest()
def main():
 p=argparse.ArgumentParser(description=__doc__);p.add_argument('--go',default='go');a=p.parse_args()
 out=HERE/'resources';out.mkdir(exist_ok=False)
 sources=sorted((ROOT/'encoding/liftrs-shared').rglob('*.go'))+sorted((ROOT/'encoding/liftrs-go').glob('*.go'))+sorted((ROOT/'encoding/liftrs-cpp').glob('*.*'))+[HERE/'resources.py',ROOT/'go.mod',ROOT/'go.sum']
 manifest={str(p.relative_to(ROOT)):sha(p) for p in sources};identity=hashlib.sha256(json.dumps(manifest,sort_keys=True).encode()).hexdigest()[:16]
 cache=ROOT/'.cache/liftrs/local-resources'/identity;cache.mkdir(parents=True,exist_ok=True)
 r={'revision':subprocess.check_output(['git','rev-parse','HEAD'],cwd=ROOT,text=True).strip(),'source_sha256':manifest,'cache_directory':str(cache.relative_to(ROOT)),'status':'running','cases':{},'environment':{'platform':platform.platform(),'go':subprocess.check_output([a.go,'version'],text=True).strip(),'host':'same 48 GiB Apple M5 Pro'},'policy':{'timeout_s':600,'go_soft_limit_bytes':9<<30,'existing_s13_cache':'read-only','measurements':'fresh process peak RSS includes load/preparation/runtime/profile; not isolated phase RSS'}}
 def save():(out/'results.json').write_text(json.dumps(r,indent=2)+'\n')
 save();(cache/'provenance.json').write_text(json.dumps({'revision':r['revision'],'source_sha256':manifest},indent=2)+'\n')
 binary=cache/'probe'
 with (out/'build.log').open('w') as log:subprocess.run([a.go,'test','-c','-tags','liftrs_cpp','-o',str(binary),'./encoding/liftrs-shared'],cwd=ROOT,stdout=log,stderr=subprocess.STDOUT,check=True)
 r['binary']={'path':str(binary.relative_to(ROOT)),'sha256':sha(binary)};save()
 for mode in ['basis-s10','baseline-s10','prototype-s10','export-s13','minimal-load-s13']:
  directory=out/mode;directory.mkdir()
  env=dict(os.environ,PIR_CALIBRATION_RESOURCE=mode,PIR_CALIBRATION_OUT=str(directory),PIR_CALIBRATION_CACHE=str(cache),PIR_CALIBRATION_S13=str(ROOT/'.cache/liftrs/s13-m2-t5/cyclic.json'))
  with (directory/'run.log').open('w') as log:
   process=subprocess.Popen([str(binary),'-test.run','^TestCalibrationResource$','-test.v'],env=env,cwd=ROOT,stdout=log,stderr=subprocess.STDOUT);start=time.monotonic();last=start-30
   try:
    while True:
     pid,status,usage=os.wait4(process.pid,os.WNOHANG)
     if pid:process.returncode=os.waitstatus_to_exitcode(status);break
     if time.monotonic()-start>600:raise TimeoutError(mode)
     if time.monotonic()-last>30:print(mode+' '+str(round(time.monotonic()-start))+'s',flush=True);last=time.monotonic()
     time.sleep(.5)
   except BaseException:
    process.kill();os.wait4(process.pid,0);raise
  case={'returncode':process.returncode,'wall_s':time.monotonic()-start,'peak_rss_bytes':usage.ru_maxrss*(1 if platform.system()=='Darwin' else 1024),'phase_data':'phases.json'};r['cases'][mode]=case;save()
  if process.returncode:raise SystemExit('resource probe failed')
  case['phases']=json.loads((directory/'phases.json').read_text());save();print(mode+' RSS '+str(round(case['peak_rss_bytes']/2**20,2))+' MiB',flush=True)
  for profile in directory.glob('*.pprof'):
   with (directory/(profile.stem+'-flat.txt')).open('w') as stream:subprocess.run([a.go,'tool','pprof','-top',str(binary),str(profile)],stdout=stream,stderr=subprocess.STDOUT,check=True)
 assert r['cases']['baseline-s10']['phases'][-1]['prepared_sha256']==r['cases']['prototype-s10']['phases'][-1]['prepared_sha256']
 r['cache_sha256']={str(p.relative_to(cache)):sha(p) for p in cache.rglob('*') if p.is_file()};r['evidence_sha256']={str(p.relative_to(out)):sha(p) for p in out.rglob('*') if p.is_file() and p!=out/'results.json'};r['status']='passed';save();print('Fresh-process resource calibration: PASS')
if __name__=='__main__':main()
