#!/usr/bin/env python3
"""Guarded full s17 CLI launch. Requires an explicitly approved budget manifest."""
import argparse,datetime,hashlib,json,os,runpy,shutil,signal,subprocess,time
from pathlib import Path
ROOT=Path(__file__).resolve().parents[2];HERE=Path(__file__).resolve().parent
HELP=runpy.run_path(str(ROOT/'experiments/hpc-cached-s15/run.py'))
GUARD=runpy.run_path(str(ROOT/'experiments/hpc-s16/run.py'))
STOP=runpy.run_path(str(ROOT/'experiments/gpu-preprocessing-workflow/run.py'))['stop_process']
GIB=1<<30
class CheckpointStop(Exception):pass

def validate(plan):
 h=plan['host_route'];g=plan['gpu_route'];s=plan['storage'];d=plan['deadlines']
 assert plan['exact_k']==3876566203 and plan['n']==1<<34
 assert h['exclusive_node'] and h['allocated_cpu_cores']==128 and h['active_workers']==32
 assert h['requested_ram_bytes']==864*GIB and h['process_group_rss_cap_bytes']==800*GIB
 assert h['go_soft_memory_limit_bytes']==720*GIB and h['principal_array_cap_bytes']==128*GIB
 assert h['native_owned_cap_bytes']==128<<20 and h['final_code_payload_cap_bytes']==161*GIB
 assert g['gpus']==4 and g['nodes']==1 and g['per_device_owned_cap_bytes']==112*GIB
 assert g['aggregate_owned_cap_bytes']==448*GIB and plan['vector_layout']in('peer','local')
 assert s['new_retained_cap_bytes']==512*GIB and s['scratch_cap_bytes']==16*GIB
 assert s['minimum_initial_shared_free_bytes']==576*GIB
 assert s['projected_retained_payload_bytes']==12*plan['exact_k']+12*plan['n']+61
 assert d['worker_seconds']==97*3600<d['controller_seconds']==98*3600<d['slurm_seconds']==100*3600
 assert d['checkpoint_notice_seconds']==3600
 assert plan['recovery']['interval_seconds']==7200 and plan['recovery']['completed_generations_per_phase']==2
 assert plan['recovery']['periodic_write_overhead_target_fraction']==0.01

def checkpoint_notice_time(began,timeout,started,deadlines):
 return min(began+timeout,started+deadlines['controller_seconds'])-deadlines['checkpoint_notice_seconds']

def last_worker_event(path):
 if not path.exists():return None
 with path.open('rb')as f:
  f.seek(max(0,path.stat().st_size-65536));lines=f.read().decode(errors='replace').splitlines()
 for line in reversed(lines):
  try:
   row=json.loads(line)
   if isinstance(row,dict)and'phase'in row:return row
  except ValueError:pass
 return None

def main():
 p=argparse.ArgumentParser(description=__doc__);p.add_argument('name',nargs='?');p.add_argument('--check',action='store_true');p.add_argument('--resume',action='store_true')
 p.add_argument('--budget',type=Path,default=HERE/'budget.json');args=p.parse_args()
 budget_raw=args.budget.read_bytes();plan=json.loads(budget_raw);validate(plan)
 if args.check:print('full plan/format sizes and guard order checked; approved='+str(plan.get('approved',False))+'; no allocation or launch');return
 if not plan.get('approved')or not plan.get('approval_note'):p.error('explicit numeric user budget approval must be recorded before full launch')
 if not args.name or not args.name.replace('-','').isalnum():p.error('fresh result name required')
 if not os.environ.get('SLURM_JOB_ID'):p.error('submit through the exclusive-node Slurm script')
 assert int(os.environ.get('SLURM_JOB_NUM_NODES','0'))==1
 assert int(os.environ.get('SLURM_CPUS_ON_NODE','0'))>=128,'exclusive128-core allocation required'
 assert int(os.environ.get('SLURM_MEM_PER_NODE','0'))>=864*1024,'864GiB host allocation required'
 count_analysis=ROOT/'experiments/s17-preprocessing-admission/runs/counts-01/analysis.json'
 count_raw=count_analysis.read_bytes();counted=json.loads(count_raw);assert counted['status']=='passed'
 assert next(r['k']for r in counted['dimensions']if r['s']==17)==plan['exact_k']
 if shutil.disk_usage(ROOT).free<plan['storage']['minimum_initial_shared_free_bytes']:raise RuntimeError('initial shared storage headroom unavailable; quota still must be agreed')
 if shutil.disk_usage('/tmp').free<32*GIB:raise RuntimeError('initial local scratch headroom unavailable')
 if GUARD['mem_available']()<832*GIB:raise RuntimeError('exclusive node lacks initial832GiB available host memory')
 out=HERE/'runs'/args.name;out.mkdir(parents=True,exist_ok=False)
 (out/'budget.json').write_bytes(budget_raw);(out/'count-admission-analysis.json').write_bytes(count_raw)
 cache=ROOT/'.cache/liftrs/gpu-s17-full-preprocessing';cache.mkdir(exist_ok=True,parents=True)
 stages=cache/'s17-m2-t5'
 if not args.resume and any((stages/n).exists()for n in('support','root','code')):raise RuntimeError('completed stages exist; explicit --resume required')
 scratch=Path('/tmp')/('depir-s17-full-'+os.environ['SLURM_JOB_ID']);scratch.mkdir(exist_ok=False)
 env={k:v for k,v in os.environ.items()if not k.startswith('PIR_')}
 env.update(GOTOOLCHAIN='local',GOPROXY='off',GOSUMDB='off',GOPATH=str(ROOT/'.cache/liftrs/toolchains/gopath'),GOCACHE=str(scratch/'go-cache'),TMPDIR=str(scratch),CGO_ENABLED='1',CXX='g++',GOMAXPROCS='32',GOFLAGS='-p=4',GOMEMLIMIT=str(plan['host_route']['go_soft_memory_limit_bytes']),LIFTRS_CPU_PROGRESS='1')
 go=str(ROOT/'.cache/liftrs/toolchains/go1.27.3-linux-amd64/go/bin/go')
 record={'status':'running','job_id':os.environ['SLURM_JOB_ID'],'revision':subprocess.check_output(['git','rev-parse','HEAD'],cwd=ROOT,text=True).strip(),'started_utc':datetime.datetime.now(datetime.timezone.utc).isoformat(),'cache':str(cache),'code_cache':str(stages),'plan':plan,'budget_sha256':hashlib.sha256(budget_raw).hexdigest(),'count_analysis_sha256':hashlib.sha256(count_raw).hexdigest(),'commands':[]}
 start=time.monotonic();cancelled=[];signal.signal(signal.SIGTERM,lambda *_:cancelled.append(True))
 def save():
  temp=out/'results.json.tmp';temp.write_text(json.dumps(record,indent=2)+'\n');temp.replace(out/'results.json')
 def event(label,kind,**fields):
  row=dict(utc=datetime.datetime.now(datetime.timezone.utc).isoformat(),phase=label,event=kind,elapsed_seconds=time.monotonic()-start,**fields)
  with(out/'progress.jsonl').open('a')as f:f.write(json.dumps(row)+'\n');f.flush()
  print(json.dumps(row),flush=True)
 def run(label,argv,cwd=ROOT,timeout=600):
  began=time.monotonic();last=began;peak=0;reason=None;shutdown=None;sizes=(0,0);last_size=0;checkpoint_notice=False
  event(label,'started');log_path=out/(label+'.log')
  with log_path.open('w')as log:
   proc=subprocess.Popen(list(map(str,argv)),cwd=cwd,env=env,stdout=log,stderr=subprocess.STDOUT,start_new_session=True)
   try:
    while proc.poll()is None:
     now=time.monotonic();peak=max(peak,HELP['group_rss'](proc.pid))
     if now-last_size>=30:sizes=(GUARD['tree_bytes'](cache)+GUARD['tree_bytes'](out),GUARD['tree_bytes'](scratch));last_size=now
     if shutil.disk_usage(cache).free<64*GIB:reason='shared filesystem headroom'
     if peak>plan['host_route']['process_group_rss_cap_bytes']:reason='RSS guard'
     if sizes[0]>plan['storage']['new_retained_cap_bytes']or sizes[1]>plan['storage']['scratch_cap_bytes']:reason='storage guard'
     if GUARD['mem_available']()<32*GIB:reason='node available-memory headroom'
     if now-began>timeout or now-start>plan['deadlines']['controller_seconds']:reason='deadline'
     if cancelled:reason='termination signal'
     if label=='preprocessing'and not reason and not checkpoint_notice and now>=checkpoint_notice_time(began,timeout,start,plan['deadlines']):
      try:os.kill(proc.pid,signal.SIGUSR1)
      except ProcessLookupError:pass
      checkpoint_notice=True;event(label,'checkpoint-requested',remaining_work='unknown',note='save and stop at next recoverable boundary; hard guard remains active')
     # This is job-side validation of newly published support, never agent monitoring.
     support=stages/'support/manifest.json'
     if label=='preprocessing'and support.exists():
      m=json.loads(support.read_text())
      if m['support']['parameters']['k']!=plan['exact_k']:reason='support dimension differs from verified count'
     if now-last>=30:
      event(label,'heartbeat',peak_rss_bytes=peak,retained_bytes=sizes[0],scratch_bytes=sizes[1],worker_last_event=last_worker_event(log_path),remaining_work='unknown');last=now
     if reason:shutdown=STOP(proc);break
     time.sleep(2)
   finally:
    if proc.poll()is None and shutdown is None:shutdown=STOP(proc)
  record['commands'].append(dict(label=label,argv=list(map(str,argv)),seconds=time.monotonic()-began,peak_rss_bytes=peak,exit_code=proc.returncode,guard=reason,shutdown=shutdown));save()
  event(label,'checkpointed'if label=='preprocessing'and not reason and proc.returncode==75 else('failed'if reason or proc.returncode else'completed'),peak_rss_bytes=peak)
  if label=='preprocessing'and not reason and proc.returncode==75:raise CheckpointStop('worker stopped at a persisted boundary')
  if reason or proc.returncode:raise RuntimeError(label+': '+str(reason or proc.returncode))
 save()
 try:
  source=scratch/'repo';shutil.copytree(ROOT,source,ignore=shutil.ignore_patterns('.git','.cache','runs','reviews','__pycache__','slurm-*.log','reconcile-slurm-*.log'))
  record['source_sha256']={str(f.relative_to(source)):HELP['sha'](f)for f in source.rglob('*')if f.is_file()and(f.suffix in('.go','.h','.hpp','.cpp','.cu','.cuh','.py','.sbatch')or f.name in('go.mod','go.sum'))};save()
  for label,argv in [('host',['hostname']),('cpu',['lscpu']),('job',['scontrol','show','job',record['job_id']]),('gpu',['nvidia-smi','--query-gpu=name,uuid,driver_version,memory.total,memory.free','--format=csv']),('topology',['nvidia-smi','topo','-m']),('go',[go,'version']),('compiler',['g++','--version']),('nvcc',['nvcc','--version'])]:run(label,argv)
  # Small independent preflight, before reserving large device payloads.
  probe=cache/('peer-probe-'+record['job_id']);library=cache/('libliftrs-cuda-'+record['job_id']+'.so');cli=cache/('preprocess-'+record['job_id'])
  run('probe-build',['nvcc','-std=c++17','-O2','-arch=native','-Iencoding/liftrs-cuda','experiments/gpu-preprocessing-workflow/peer-probe.cu','-o',probe],source)
  run('peer-probe',[probe],source,timeout=60)
  run('cuda-build',['nvcc','-std=c++17','-O3','-lineinfo','-arch=native','-shared','-Xcompiler=-fPIC,-pthread','encoding/liftrs-cuda/matrix.cu','encoding/liftrs-cuda/preparation.cu','-o',library],source)
  run('cli-build',[go,'build','-tags','liftrs_cpp,liftrs_cuda','-ldflags','-X main.sourceRevision='+record['revision'],'-o',cli,'./cmd/pir-preprocess'],source)
  # Recovery hooks share CPU/GPU control flow; independently exercise fresh
  # native stop/restart before starting the large workload.
  run('recovery-preflight',[go,'test','-tags','liftrs_cpp','./cmd/pir-preprocess','./encoding/liftrs-shared','-run','Test.*Recovery','-count=1'],source)
  h=plan['host_route'];g=plan['gpu_route']
  command=[str(cli),'--s','17','--t','5','--workers',str(h['active_workers']),'--backend','cuda','--devices','0,1,2,3','--gpu-library',str(library),'--gpu-bytes',str(g['per_device_owned_cap_bytes']),'--max-array-bytes',str(h['principal_array_cap_bytes']),'--native-bytes',str(h['native_owned_cap_bytes']),'--max-code-bytes',str(h['final_code_payload_cap_bytes']),'--cache',str(stages),'--vector-layout',plan['vector_layout'],'--checkpoint-interval','2h']
  if args.resume:command.append('--resume')
  run('preprocessing',command,source,timeout=plan['deadlines']['worker_seconds'])
  support=json.loads((stages/'support/manifest.json').read_text());assert support['support']['parameters']['k']==plan['exact_k']
  record['stage_manifests']={name:json.loads((stages/name/'manifest.json').read_text())for name in('support','root','code')}
  record['status']='passed'
 except CheckpointStop as error:record['status']='checkpointed';record['note']=str(error)
 except BaseException as error:record['status']='failed';record['error']=repr(error);raise
 finally:
  record['elapsed_seconds']=time.monotonic()-start
  # CLI hashes/readbacks large payloads once; do not silently rehash235GiB on exit.
  # Preserve manifests and raw logs. Large-file independent reconciliation is a
  # separately admitted post-completion operation, not hidden benchmark time.
  record['cache_sha256']={str(f.relative_to(cache)):HELP['sha'](f)for f in cache.rglob('*')if f.is_file()and f.name not in('support.bin','root.bin','online.bin','generator.bin','progress.bin')}
  record['evidence_sha256']={str(f.relative_to(out)):HELP['sha'](f)for f in out.rglob('*')if f.is_file()and f.name not in('results.json','results.json.tmp')};save()
if __name__=='__main__':main()
