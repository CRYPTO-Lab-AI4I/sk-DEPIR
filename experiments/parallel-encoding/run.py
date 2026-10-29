#!/usr/bin/env python3
"""Profile full parallel matrix encoding, reusing verified preprocessing."""
import argparse,hashlib,json,os,platform,subprocess,time
from pathlib import Path
ROOT=Path(__file__).resolve().parents[2];HERE=Path(__file__).resolve().parent

def sha(p):
 h=hashlib.sha256()
 with p.open('rb') as f:
  for b in iter(lambda:f.read(1<<20),b''):h.update(b)
 return h.hexdigest()
def main():
 p=argparse.ArgumentParser(description=__doc__);p.add_argument('--go',default='go');p.add_argument('--run-name',default='s13');p.add_argument('--smoke',action='store_true');p.add_argument('--workers',default='8,16');p.add_argument('--repeats',type=int,default=3);a=p.parse_args()
 if not a.run_name or any(c not in 'abcdefghijklmnopqrstuvwxyz0123456789-_' for c in a.run_name):raise SystemExit('invalid name')
 out=HERE/a.run_name;out.mkdir(exist_ok=False)
 sources=sorted((ROOT/'encoding/liftrs-shared').rglob('*.go'))+sorted((ROOT/'encoding/liftrs-go').glob('*.go'))+sorted((ROOT/'encoding/liftrs-cpp').glob('*.*'))+[HERE/'main.go',HERE/'run.py',ROOT/'go.mod',ROOT/'go.sum']
 manifest={str(p.relative_to(ROOT)):sha(p) for p in sources};identity=hashlib.sha256(json.dumps(manifest,sort_keys=True).encode()).hexdigest()[:16]
 cache=ROOT/'.cache/liftrs/parallel-encoding'/identity;cache.mkdir(parents=True,exist_ok=True)
 r={'revision':subprocess.check_output(['git','rev-parse','HEAD'],cwd=ROOT,text=True).strip(),'source_sha256':manifest,'status':'running','cache_directory':str(cache.relative_to(ROOT)),'environment':{'platform':platform.platform(),'go':subprocess.check_output([a.go,'version'],text=True).strip(),'compiler':subprocess.check_output([os.environ.get('CXX','clang++'),'--version'],text=True).strip(),'host':'Apple M5 Pro, 18 logical CPUs, 48 GiB; same local calibration host'},'policy':{'timeout_s':1800,'go_soft_memory_bytes':9<<30,'profiling':True,'preprocessing':'retained read-only s13 minimal file; small fixture retained on smoke','baseline':'153.409 s historical native median, not rerun','native_stage_times':'elapsed, not CPU samples; columns includes blocked transposes'}}
 def save():(out/'results.json').write_text(json.dumps(r,indent=2)+'\n')
 save();(cache/'provenance.json').write_text(json.dumps({'revision':r['revision'],'source_sha256':manifest},indent=2)+'\n')
 binary=cache/'probe'
 with (out/'build.log').open('w') as f:subprocess.run([a.go,'build','-tags','liftrs_cpp','-o',str(binary),'./experiments/parallel-encoding'],cwd=ROOT,stdout=f,stderr=subprocess.STDOUT,check=True)
 r['binary']={'path':str(binary.relative_to(ROOT)),'sha256':sha(binary)}
 if a.smoke:source=cache/'small'
 else:
  matches=list((ROOT/'.cache/liftrs/local-resources').glob('*/s13-online.bin'))
  if len(matches)!=1:raise RuntimeError('select unique retained s13 file')
  source=matches[0];r['input_cache']={'path':str(source.relative_to(ROOT)),'bytes':source.stat().st_size,'sha256':sha(source)}
 command=[str(binary),'-out',str(out),'-cache',str(source),'-s','8' if a.smoke else '13','-workers',a.workers,'-repeats',str(a.repeats)];r['command']=command;save()
 with (out/'events.log').open('w') as stdout,(out/'stderr.txt').open('w') as stderr:
  process=subprocess.Popen(command,cwd=ROOT,stdout=stdout,stderr=stderr);start=time.monotonic();last=start-30
  try:
   while True:
    pid,status,usage=os.wait4(process.pid,os.WNOHANG)
    if pid:process.returncode=os.waitstatus_to_exitcode(status);break
    if time.monotonic()-start>1800:raise TimeoutError('30-minute case limit')
    if time.monotonic()-last>30:
     print(str(round(time.monotonic()-start))+'s: '+(out/'events.log').read_text().splitlines()[-1] if (out/'events.log').stat().st_size else 'encoding started',flush=True);last=time.monotonic()
    time.sleep(.5)
  except BaseException as err:
   process.kill();os.wait4(process.pid,0);r['status']='failed';r['error']=repr(err);save();raise
 r.update(status='passed' if process.returncode==0 else 'failed',returncode=process.returncode,wall_s=time.monotonic()-start,process_peak_rss_bytes=usage.ru_maxrss*(1 if platform.system()=='Darwin' else 1024));save()
 if process.returncode:raise SystemExit('encoding failed; partial evidence retained')
 r['phases']=json.loads((out/'phases.json').read_text())
 assert all(x['full_hash_match'] for x in r['phases'] if x['phase'].startswith('verify-'))
 for profile in sorted(out.glob('*.pprof')):
  with (out/(profile.stem+'-flat.txt')).open('w') as f:subprocess.run([a.go,'tool','pprof','-top',str(binary),str(profile)],stdout=f,stderr=subprocess.STDOUT,check=True)
 r['cache_sha256']={str(p.relative_to(cache)):sha(p) for p in cache.rglob('*') if p.is_file()};r['evidence_sha256']={p.name:sha(p) for p in out.iterdir() if p.is_file() and p.name!='results.json'};save();print('parallel encoding PASS',flush=True)
if __name__=='__main__':main()
