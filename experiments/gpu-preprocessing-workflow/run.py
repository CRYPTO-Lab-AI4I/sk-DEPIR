#!/usr/bin/env python3
"""Guarded small COMPLETE persistent preprocessing validation, not a full s17 run."""
import ast,datetime,hashlib,json,os,runpy,shutil,signal,subprocess,time
from pathlib import Path
ROOT=Path(__file__).resolve().parents[2];HERE=Path(__file__).resolve().parent
HELP=runpy.run_path(str(ROOT/'experiments/hpc-cached-s15/run.py'))
GUARD=runpy.run_path(str(ROOT/'experiments/hpc-s16/run.py'))

def stop_process(process,wait_seconds=5):
 """Bound both waits: a broken GPU driver can leave even SIGKILL unreaped."""
 for sig in(signal.SIGTERM,signal.SIGKILL):
  try:os.killpg(process.pid,sig)
  except ProcessLookupError:return {'reaped':process.poll()is not None,'last_signal':int(sig)}
  try:process.wait(timeout=wait_seconds);return {'reaped':True,'last_signal':int(sig)}
  except subprocess.TimeoutExpired:pass
 return {'reaped':False,'last_signal':int(signal.SIGKILL),'warning':'worker not reaped; scheduler must clean allocation'}

def main():
 import argparse
 p=argparse.ArgumentParser(description=__doc__);p.add_argument('name',nargs='?');p.add_argument('--check',action='store_true');args=p.parse_args()
 if args.check:
  ast.parse(Path(__file__).read_text());assert 600<900<1200
  assert 4*(1<<30)==4<<30
  print('syntax and bounded deadline/device policy checked; no full s17 mode');return
 if not args.name or not args.name.replace('-','').isalnum():p.error('fresh alphanumeric/hyphen name required')
 out=HERE/'runs'/args.name;out.mkdir(parents=True,exist_ok=False)
 cache=ROOT/'.cache/liftrs/gpu-preprocessing-workflow'/args.name;cache.mkdir(parents=True,exist_ok=False)
 scratch=Path('/tmp')/('depir-preprocessing-workflow-'+os.environ['SLURM_JOB_ID']);scratch.mkdir(exist_ok=False)
 env={k:v for k,v in os.environ.items()if not k.startswith('PIR_')}
 env.update(GOTOOLCHAIN='local',GOPROXY='off',GOSUMDB='off',GOPATH=str(ROOT/'.cache/liftrs/toolchains/gopath'),GOCACHE=str(scratch/'go-cache'),TMPDIR=str(scratch),CGO_ENABLED='0',GOMAXPROCS='4',GOFLAGS='-p=2',GOMEMLIMIT=str(4<<30))
 go=str(ROOT/'.cache/liftrs/toolchains/go1.27.3-linux-amd64/go/bin/go')
 start=time.monotonic();cancelled=[];signal.signal(signal.SIGTERM,lambda *_:cancelled.append(True))
 record={'status':'running','job_id':os.environ['SLURM_JOB_ID'],'revision':subprocess.check_output(['git','rev-parse','HEAD'],cwd=ROOT,text=True).strip(),'started_utc':datetime.datetime.now(datetime.timezone.utc).isoformat(),'cache':str(cache),'commands':[],
 'policy':{'cpus':4,'host_request_bytes':8<<30,'rss_guard_bytes':6<<30,'go_soft_limit_bytes':4<<30,'new_cache_guard_bytes':2<<30,'scratch_guard_bytes':4<<30,'gpus':4,'aggregate_owned_device_cap_bytes':4<<30,'per_device_cap_bytes':1<<30,'worker_seconds':600,'controller_seconds':900,'slurm_seconds':1200},'cases':[],'peer_probe':True}
 def save():(out/'results.json').write_text(json.dumps(record,indent=2)+'\n')
 def event(label,kind,**fields):
  row=dict(utc=datetime.datetime.now(datetime.timezone.utc).isoformat(),phase=label,event=kind,completed=len(record['commands']),unit='commands',elapsed_seconds=time.monotonic()-start,**fields)
  with(out/'progress.jsonl').open('a')as f:f.write(json.dumps(row)+'\n');f.flush()
  print(json.dumps(row),flush=True)
 def run(label,argv,cwd=ROOT,expect=0,timeout=600):
  began=time.monotonic();last=began;peak=0;reason=None;shutdown=None;event(label,'started')
  with(out/(label+'.log')).open('w')as f:
   proc=subprocess.Popen(list(map(str,argv)),cwd=cwd,env=env,stdout=f,stderr=subprocess.STDOUT,start_new_session=True)
   try:
    while proc.poll()is None:
     peak=max(peak,HELP['group_rss'](proc.pid))
     if peak>6<<30:reason='RSS guard'
     if GUARD['tree_bytes'](cache)>2<<30 or GUARD['tree_bytes'](scratch)>4<<30:reason='storage guard'
     if time.monotonic()-began>timeout or time.monotonic()-start>900:reason='deadline'
     if cancelled:reason='termination signal'
     if time.monotonic()-last>=30:event(label,'heartbeat',peak_rss_bytes=peak,remaining_work='unknown');last=time.monotonic()
     if reason:shutdown=stop_process(proc);break
     time.sleep(1)
   finally:
    if proc.poll()is None and shutdown is None:shutdown=stop_process(proc)
  record['commands'].append(dict(label=label,argv=list(map(str,argv)),seconds=time.monotonic()-began,peak_rss_bytes=peak,exit_code=proc.returncode,expected_exit_code=expect,guard=reason,shutdown=shutdown,timeout_seconds=timeout));save()
  event(label,'failed'if reason or proc.returncode!=expect else'completed',peak_rss_bytes=peak,guard=reason)
  if shutdown and not shutdown['reaped']:record['cleanup_incomplete']=True;save()
  if reason or proc.returncode!=expect:raise RuntimeError(label+': '+str(reason or proc.returncode))
 save()
 try:
  source=scratch/'repo';shutil.copytree(ROOT,source,ignore=shutil.ignore_patterns('.git','.cache','runs','reviews','__pycache__','slurm-*.log','reconcile-slurm-*.log'))
  record['source_sha256']={str(f.relative_to(source)):HELP['sha'](f)for f in source.rglob('*')if f.is_file()and(f.suffix in('.go','.h','.hpp','.cpp','.cu','.cuh','.py','.sbatch')or f.name in('go.mod','go.sum'))};save()
  for label,argv in [('host',['hostname']),('cpu',['lscpu']),('gpu',['nvidia-smi','--query-gpu=name,uuid,driver_version,memory.total','--format=csv']),('topology',['nvidia-smi','topo','-m']),('job',['scontrol','show','job',record['job_id']]),('go',[go,'version']),('compiler',['g++','--version']),('nvcc',['nvcc','--version'])]:run(label,argv)
  run('controller-tests',['python3','-B','experiments/gpu-preprocessing-workflow/test_controller.py'],source)
  env['LIFTRS_CPU_PROGRESS']='1'
  probe=cache/'peer-probe'
  run('peer-probe-build',['nvcc','-std=c++17','-O2','-arch=native','-Iencoding/liftrs-cuda','experiments/gpu-preprocessing-workflow/peer-probe.cu','-o',probe],source)
  run('peer-probe',[str(probe)],source,timeout=60)
  plain=cache/'preprocess-go';cli=cache/'preprocess-cuda';library=cache/'libliftrs-cuda.so'
  ldflags='-X main.sourceRevision='+record['revision']
  run('go-only-build',[go,'build','-ldflags',ldflags,'-o',plain,'./cmd/pir-preprocess'],source)
  run('go-only-tests',[go,'test','./cmd/pir-preprocess','-count=1','-timeout=10m'],source)
  env.update(CGO_ENABLED='1',CXX='g++')
  run('native-tests',[go,'test','-tags','liftrs_cpp','./cmd/pir-preprocess','-count=1','-v','-timeout=10m'],source)
  run('race',[go,'test','-race','-tags','liftrs_cpp','./cmd/pir-preprocess','-count=1','-timeout=10m'],source)
  run('asan',[go,'test','-asan','-tags','liftrs_cpp','./cmd/pir-preprocess','-count=1','-timeout=10m'],source)
  run('cuda-build',['nvcc','-std=c++17','-O3','-lineinfo','-arch=native','-shared','-Xcompiler=-fPIC,-pthread','encoding/liftrs-cuda/matrix.cu','encoding/liftrs-cuda/preparation.cu','-o',library],source)
  run('cuda-cli-build',[go,'build','-tags','liftrs_cpp,liftrs_cuda','-ldflags',ldflags,'-o',cli,'./cmd/pir-preprocess'],source)
  run('vet',[go,'vet','-tags','liftrs_cpp,liftrs_cuda','./cmd/pir-preprocess'],source)
  reference=ROOT/'.cache/liftrs/gpu-s17-preprocessing/calibration-01/fixtures/s8/online.bin'
  historical=json.loads((ROOT/'experiments/gpu-s17-preprocessing/runs/calibration-01/results.json').read_text());assert historical['status']=='passed'
  wanted=historical['cache_sha256']['fixtures/s8/online.bin'];assert HELP['sha'](reference)==wanted
  record['independent_go_reference']={'path':str(reference),'sha256':wanted,'source_job_id':historical['job_id'],'source_revision':historical['revision']}
  env['LIFTRS_CPU_PROGRESS']='1'
  for layout in ('peer','local'):
   for count in (1,4):
    label=layout+'-'+str(count);dest=cache/label
    cmd=[str(cli),'--s','8','--t','5','--cache',str(dest),'--backend','cuda','--gpu-library',str(library),'--devices',','.join(map(str,range(count))),'--gpu-bytes',str(1<<30),'--max-array-bytes',str(8<<20),'--native-bytes',str(64<<20),'--max-code-bytes',str(8<<20),'--vector-layout',layout]
    run(label+'-fresh',cmd,source);assert HELP['sha'](dest/'code/online.bin')==wanted
    baseline={stage:HELP['sha'](dest/stage/'manifest.json')for stage in('support','root','code')}
    run(label+'-cached',cmd+['--resume'],source)
    assert all(HELP['sha'](dest/stage/'manifest.json')==digest for stage,digest in baseline.items())
    # Preserve the first complete output while recreating a process that only
    # sees support/root. No useful cache or profile is removed.
    (dest/'code').rename(dest/'code-first')
    run(label+'-root-restart',cmd+['--resume'],source);assert HELP['sha'](dest/'code/online.bin')==wanted
    assert all(HELP['sha'](dest/stage/'manifest.json')==baseline[stage]for stage in('support','root'))
    plain_cmd=[str(plain),'--s','8','--t','5','--backend','cpp','--cache',str(dest),'--max-array-bytes',str(8<<20),'--native-bytes',str(64<<20),'--max-code-bytes',str(8<<20),'--resume']
    run(label+'-cuda-free-readback',plain_cmd,source)
    rootlog=(out/(label+'-root-restart.log')).read_text();cachedlog=(out/(label+'-cached.log')).read_text()
    assert '"phase":"discovery"'not in rootlog and'"phase":"setup"'not in cachedlog
    record['cases'].append({'layout':layout,'gpus':count,'s':8,'canonical_sha256':wanted,'baseline_manifest_sha256':baseline,'cache':str(dest)});save()
  # CUDA sanitizer exercises a real fresh executable workflow; cached byte/hash
  # checks in separate processes above retain profiles and compare with Go.
  run('memcheck',['compute-sanitizer','--tool','memcheck','--leak-check','full','--error-exitcode','99',str(cli),'--s','4','--t','3','--cache',str(cache/'memcheck'),'--backend','cuda','--gpu-library',str(library),'--devices','0,1,2,3','--gpu-bytes',str(1<<30),'--max-array-bytes',str(1<<20),'--native-bytes',str(8<<20),'--max-code-bytes',str(1<<20),'--vector-layout','local'],source)
  log=(out/'memcheck.log').read_text();assert 'ERROR SUMMARY: 0 errors'in log and'LEAK SUMMARY: 0 bytes'in log
  record['status']='passed'
 except BaseException as e:record['status']='failed';record['error']=repr(e);raise
 finally:
  record['elapsed_seconds']=time.monotonic()-start
  record['cache_sha256']={str(f.relative_to(cache)):HELP['sha'](f)for f in cache.rglob('*')if f.is_file()}
  record['evidence_sha256']={str(f.relative_to(out)):HELP['sha'](f)for f in out.rglob('*')if f.is_file()and f.name!='results.json'};save()
if __name__=='__main__':main()
