#!/usr/bin/env python3
"""Profile independent C++ tile scaling, explicit-cache file I/O and overlap."""
import argparse,hashlib,json,os,platform,subprocess,time
from pathlib import Path
ROOT=Path(__file__).resolve().parents[2];HERE=Path(__file__).resolve().parent

def sha(path):
 h=hashlib.sha256()
 with path.open('rb') as stream:
  for b in iter(lambda:stream.read(1<<20),b''):h.update(b)
 return h.hexdigest()
def main():
 p=argparse.ArgumentParser(description=__doc__);p.add_argument('--go',default='go');p.add_argument('--run-name',default='initial');p.add_argument('--quick',action='store_true');p.add_argument('--small-tiles',action='store_true');a=p.parse_args()
 if not a.run_name or any(c not in 'abcdefghijklmnopqrstuvwxyz0123456789-_' for c in a.run_name):raise SystemExit('invalid run name')
 out=HERE/a.run_name;out.mkdir(exist_ok=False)
 sources=sorted((ROOT/'encoding/liftrs-shared').rglob('*.go'))+sorted((ROOT/'encoding/liftrs-go').glob('*.go'))+sorted((ROOT/'encoding/liftrs-cpp').glob('*.*'))+sorted(HERE.glob('*.go'))+[HERE/'run.py',ROOT/'go.mod',ROOT/'go.sum']
 manifest={str(p.relative_to(ROOT)):sha(p) for p in sources};identity=hashlib.sha256(json.dumps(manifest,sort_keys=True).encode()).hexdigest()[:16]
 cache=ROOT/'.cache/liftrs/local-calibration'/identity/a.run_name;cache.mkdir(parents=True,exist_ok=True)
 r={'revision':subprocess.check_output(['git','rev-parse','HEAD'],cwd=ROOT,text=True).strip(),'source_sha256':manifest,'environment':{'platform':platform.platform(),'go':subprocess.check_output([a.go,'version'],text=True).strip(),'compiler':subprocess.check_output([os.environ.get('CXX','clang++'),'--version'],text=True).strip(),'cpu':subprocess.check_output(['sysctl','-n','machdep.cpu.brand_string'],text=True).strip(),'logical_cpus':int(subprocess.check_output(['sysctl','-n','hw.logicalcpu'],text=True)),'memory_bytes':int(subprocess.check_output(['sysctl','-n','hw.memsize'],text=True))},'cache_directory':str(cache.relative_to(ROOT)),'quick':a.quick,'small_tiles':a.small_tiles,'status':'running','policy':{'timeout_s':900,'go_soft_memory_bytes':2<<30,'retained_storage_budget_bytes':2<<30,'file_sync':'fsync, not F_FULLFSYNC','native_profiles':'Go CPU profiles include cgo calls; kernels unchanged, prior detailed native sample retained'}}
 def save():(out/'results.json').write_text(json.dumps(r,indent=2)+'\n')
 save();(cache/'provenance.json').write_text(json.dumps({'revision':r['revision'],'source_sha256':manifest},indent=2)+'\n')
 binary=cache/'probe'
 with (out/'build.log').open('w') as log:subprocess.run([a.go,'build','-tags','liftrs_cpp','-o',str(binary),'./experiments/local-calibration'],cwd=ROOT,stdout=log,stderr=subprocess.STDOUT,check=True)
 r['binary']={'path':str(binary.relative_to(ROOT)),'sha256':sha(binary)};save()
 command=[str(binary),'-out',str(out),'-cache',str(cache)]+(['-quick'] if a.quick else [])+(['-cpu-only','-tile-elements','8192','-tiles','2048'] if a.small_tiles else []);r['command']=command;save()
 with (out/'events.jsonl').open('w') as stdout,(out/'stderr.txt').open('w') as stderr:
  process=subprocess.Popen(command,cwd=ROOT,stdout=stdout,stderr=stderr);start=time.monotonic();last=start-30;count=0
  try:
   while True:
    pid,status,usage=os.wait4(process.pid,os.WNOHANG)
    events=[json.loads(line) for line in (out/'events.jsonl').read_text().splitlines(keepends=True) if line.endswith('\n')]
    for e in events[count:]:
     if 'phase'in e:print(e['phase']+': '+str(round(e['elapsed_ns']/1e9,3))+'s',flush=True)
    count=len(events)
    if pid:process.returncode=os.waitstatus_to_exitcode(status);break
    if time.monotonic()-start>900:raise TimeoutError('15-minute calibration limit')
    if time.monotonic()-last>30:print('progress '+str(round(time.monotonic()-start))+'s',flush=True);last=time.monotonic()
    time.sleep(.5)
  except BaseException as err:
   process.kill();_,status,usage=os.wait4(process.pid,0);process.returncode=os.waitstatus_to_exitcode(status);r['error']=repr(err);r['status']='failed';save();raise
 r.update(status='passed' if process.returncode==0 else 'failed',returncode=process.returncode,wall_s=time.monotonic()-start,process_peak_rss_bytes=usage.ru_maxrss*(1 if platform.system()=='Darwin' else 1024),events=events)
 save()
 if process.returncode:raise SystemExit('calibration failed; partial evidence retained')
 assert events[-1]['verified'] and events[-1]['native_inverse']
 for profile in out.glob('*.pprof'):
  with (out/(profile.stem+'-flat.txt')).open('w') as stream:subprocess.run([a.go,'tool','pprof','-top',str(binary),str(profile)],stdout=stream,stderr=subprocess.STDOUT,check=True)
 r['cache_sha256']={str(p.relative_to(cache)):sha(p) for p in cache.rglob('*') if p.is_file()};r['retained_bytes']=sum(p.stat().st_size for p in cache.rglob('*') if p.is_file());assert r['retained_bytes']<2<<30
 r['evidence_sha256']={p.name:sha(p) for p in out.iterdir() if p.is_file() and p.name!='results.json'};save();print('Local calibration: PASS')
if __name__=='__main__':main()
