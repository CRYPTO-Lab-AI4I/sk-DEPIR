#!/usr/bin/env python3
"""Matched s14 CPU refactor comparison; submit and hand off, never monitor here."""
import argparse,datetime,difflib,hashlib,json,os,runpy,shutil,signal,subprocess,time
from pathlib import Path

ROOT=Path(__file__).resolve().parents[2];HERE=Path(__file__).resolve().parent
HELP=runpy.run_path(str(ROOT/'experiments/hpc-cached-s15/run.py'))
GUARD=runpy.run_path(str(ROOT/'experiments/hpc-s16/run.py'))
STOP=runpy.run_path(str(ROOT/'experiments/gpu-preprocessing-workflow/run.py'))['stop_process']
MEM=runpy.run_path(str(ROOT/'experiments/cpu-streamed-constants/cache_memory.py'))
PATCH=runpy.run_path(str(HERE/'baseline_progress.py'))['instrument']
BASELINE='671bf5e'
INPUT={'path':str(ROOT/'.cache/liftrs/parallel-s14/s14-m2-t5/online.bin'),
       'code_sha256':'88d5ff39c5df8bdd4aa0f85fc19833ae673df95365b4fd862bb83936e5a0adf8',
       'word_sha256':'b828d4dda538e95b78d914dce52fbfa7fc29a139e4781b0dad83f061f3e26d6c','s':14}
POLICY={'cpus':32,'host_request_bytes':32<<30,'rss_guard_bytes':24<<30,'cgroup_guard_bytes':30<<30,
        'new_cache_guard_bytes':16<<30,'scratch_guard_bytes':16<<30,'go_soft_bytes':20<<30,
        'worker_seconds':900,'controller_seconds':6600,'slurm_seconds':7200,'gpus':0}
ORDERS=(('old','new'),('new','old'),('old','new'),('new','old'))


def sha(path):return HELP['sha'](path)
def source_hashes(tree):
 return {str(p.relative_to(tree)):sha(p)for p in tree.rglob('*')if p.is_file()and(p.suffix in('.go','.cpp','.hpp','.h','.py','.sbatch')or p.name in('go.mod','go.sum','probe.go.txt'))}

def check():
 assert POLICY['rss_guard_bytes']<POLICY['cgroup_guard_bytes']<POLICY['host_request_bytes']
 assert POLICY['worker_seconds']<POLICY['controller_seconds']<POLICY['slurm_seconds']
 assert len(ORDERS)==4 and sum(order[0]=='old'for order in ORDERS)==2
 assert Path(INPUT['path']).is_file(),'existing verified s14 cache required; no regeneration'
 # Budget is within the already approved s14 encoding-table envelope.
 table=runpy.run_path(str(ROOT/'experiments/encoding-table/run.py'))['policy']('cpu-packed-s14')
 for key in('cpus','host_request_bytes','rss_guard_bytes','cgroup_guard_bytes','new_cache_guard_bytes','scratch_guard_bytes','worker_seconds','controller_seconds','slurm_seconds'):
  assert POLICY[key]<=table[key],key
 print('s14 input exists; no allocation/hash/benchmark; previous CPU budget envelope checked')

def main():
 parser=argparse.ArgumentParser(description=__doc__);parser.add_argument('name',nargs='?');parser.add_argument('--check',action='store_true');args=parser.parse_args()
 check()
 if args.check:return
 if not args.name or not args.name.replace('-','').isalnum():parser.error('fresh run name required')
 job=os.environ.get('SLURM_JOB_ID');assert job,'submit through Slurm'
 assert int(os.environ.get('SLURM_CPUS_PER_TASK','0'))>=32
 assert int(os.environ.get('SLURM_MEM_PER_NODE','0'))>=32768
 out=HERE/'runs'/args.name;out.mkdir(parents=True,exist_ok=False)
 cache=ROOT/'.cache/liftrs/cpu-encoding-regression'/args.name;cache.mkdir(parents=True,exist_ok=False)
 scratch=Path('/tmp')/('depir-cpu-regression-'+job);scratch.mkdir(exist_ok=False)
 assert shutil.disk_usage(cache).free>32<<30 and shutil.disk_usage(scratch).free>20<<30
 env={k:v for k,v in os.environ.items()if not k.startswith('PIR_')}
 env.update(CGO_ENABLED='1',CXX='g++',GOMAXPROCS='32',GOFLAGS='-p=4',GOTOOLCHAIN='local',GOPROXY='off',GOSUMDB='off',GOMEMLIMIT=str(POLICY['go_soft_bytes']),GOPATH=str(ROOT/'.cache/liftrs/toolchains/gopath'),GOCACHE=str(scratch/'go-cache'),TMPDIR=str(scratch),LIFTRS_CPU_PROGRESS='1')
 go=str(ROOT/'.cache/liftrs/toolchains/go1.27.3-linux-amd64/go/bin/go')
 started=time.monotonic();cancelled=[];signal.signal(signal.SIGTERM,lambda *_:cancelled.append(True))
 record={'status':'running','job_id':job,'revision':subprocess.check_output(['git','rev-parse','HEAD'],cwd=ROOT,text=True).strip(),'baseline_revision':BASELINE,'started_utc':datetime.datetime.now(datetime.timezone.utc).isoformat(),'cache':str(cache),'input':INPUT,'policy':POLICY,'orders':ORDERS,'cpu_affinity':sorted(os.sched_getaffinity(0)),'commands':[]}
 def save():
  p=out/'results.json.tmp';p.write_text(json.dumps(record,indent=2)+'\n');p.replace(out/'results.json')
 def event(phase,kind,**fields):
  row=dict(utc=datetime.datetime.now(datetime.timezone.utc).isoformat(),phase=phase,event=kind,elapsed_seconds=time.monotonic()-started,**fields)
  with(out/'progress.jsonl').open('a')as f:f.write(json.dumps(row)+'\n');f.flush()
  print(json.dumps(row),flush=True)
 def run(label,argv,cwd=ROOT,timeout=POLICY['worker_seconds']):
  began=time.monotonic();last=began;peak=0;reason=None;shutdown=None;last_storage=0;retained=temporary=0
  event(label,'started');log=out/(label+'.log')
  with log.open('w')as stream:
   proc=subprocess.Popen(list(map(str,argv)),cwd=cwd,env=env,stdout=stream,stderr=subprocess.STDOUT,start_new_session=True)
   try:
    while proc.poll()is None:
     now=time.monotonic();peak=max(peak,HELP['group_rss'](proc.pid));group=MEM['job_cgroup'](job)
     if now-last_storage>=30:
      retained=GUARD['tree_bytes'](cache)+GUARD['tree_bytes'](out);temporary=GUARD['tree_bytes'](scratch);last_storage=now
     if peak>POLICY['rss_guard_bytes']:reason='RSS guard'
     if group and group['current_bytes']>POLICY['cgroup_guard_bytes']:reason='cgroup guard'
     if retained>POLICY['new_cache_guard_bytes']or temporary>POLICY['scratch_guard_bytes']:reason='storage guard'
     if now-began>timeout or now-started>POLICY['controller_seconds']:reason='deadline'
     if cancelled:reason='termination signal'
     if now-last>=30:
      event(label,'heartbeat',peak_rss_bytes=peak,cgroup=group,retained_bytes=retained,scratch_bytes=temporary,remaining_work='unknown');last=now
     if reason:shutdown=STOP(proc);break
     time.sleep(1)
   finally:
    if proc.poll()is None and shutdown is None:shutdown=STOP(proc)
  record['commands'].append({'label':label,'argv':list(map(str,argv)),'seconds':time.monotonic()-began,'peak_rss_bytes':peak,'exit_code':proc.returncode,'guard':reason,'shutdown':shutdown});save()
  event(label,'failed'if reason or proc.returncode else'completed',peak_rss_bytes=peak)
  if reason or proc.returncode:raise RuntimeError(label+': '+str(reason or proc.returncode))
 save()
 try:
  source=scratch/'current'
  shutil.copytree(ROOT,source,ignore=shutil.ignore_patterns('.git','.cache','runs','reviews','__pycache__','slurm-*.log','reconcile-slurm-*.log'))
  record['source_sha256']=source_hashes(source);save()
  baseline=scratch/'baseline';shutil.copytree(source,baseline)
  dirs=('encoding/liftrs-shared','encoding/liftrs-cpp')
  for directory in dirs:shutil.rmtree(baseline/directory)
  archive=scratch/'baseline.tar'
  with archive.open('wb')as f:subprocess.run(['git','archive',BASELINE,*dirs],cwd=ROOT,stdout=f,check=True)
  run('baseline-extract',['tar','-xf',archive,'-C',baseline])
  pristine={str(p.relative_to(baseline)):sha(p)for d in dirs for p in(baseline/d).rglob('*')if p.is_file()}
  PATCH(baseline,source)
  record['baseline_pristine_sha256']=pristine
  modified={str(p.relative_to(baseline)):sha(p)for d in dirs for p in(baseline/d).rglob('*')if p.is_file()}
  record['baseline_modified_sha256']=modified
  record['baseline_patch_changed']=[p for p in modified if modified[p]!=pristine.get(p)]
  assert set(record['baseline_patch_changed'])=={'encoding/liftrs-cpp/c_api.cpp','encoding/liftrs-cpp/matrix.hpp','encoding/liftrs-cpp/progress.hpp'}
  diff=[]
  for name in record['baseline_patch_changed']:
   original=subprocess.run(['git','show',BASELINE+':'+name],cwd=ROOT,capture_output=True,text=True)
   if original.returncode and name!='encoding/liftrs-cpp/progress.hpp':raise RuntimeError(original.stderr)
   diff.extend(difflib.unified_diff(original.stdout.splitlines(True),(baseline/name).read_text().splitlines(True),fromfile='pristine/'+name,tofile='instrumented/'+name))
  (out/'baseline-progress.diff').write_text(''.join(diff));save()
  for label,argv in [('host',['hostname']),('cpu',['lscpu']),('job',['scontrol','show','job',job]),('go',[go,'version']),('compiler',['g++','--version'])]:run(label,argv)
  probe=source/'experiments/cpu-encoding-regression/probe.go.txt';record['probe_sha256']=sha(probe)
  binaries={}
  for version,tree in [('old',baseline),('new',source)]:
   shutil.copyfile(probe,tree/'encoding/liftrs-shared/cpu_regression_probe_test.go')
   binary=cache/(version+'.test');binaries[version]=binary
   run(version+'-build',[go,'test','-tags','liftrs_cpp','-c','-o',binary,'./encoding/liftrs-shared'],tree)
  fixture=cache/'fixture';env['PIR_REGRESSION_FIXTURE']=str(fixture)
  run('fixture',[binaries['new'],'-test.run=^TestCPURegressionFixture$','-test.v'],source/'encoding/liftrs-shared')
  env.pop('PIR_REGRESSION_FIXTURE');env['PIR_REGRESSION_INPUT']=str(fixture/'input.json')
  for version,tree in [('old',baseline),('new',source)]:
   run('small-'+version,[binaries[version],'-test.run=^TestCPURegressionSample$','-test.v','-test.timeout=10m'],tree/'encoding/liftrs-shared')
  input_path=cache/'s14-input.json';input_path.write_text(json.dumps(INPUT,indent=2)+'\n');env['PIR_REGRESSION_INPUT']=str(input_path)
  # Historical full-word provenance is read once; every sample separately checks
  # the code digest and canonical word hash. Never regenerate preprocessing here.
  meta=json.loads((Path(INPUT['path']).parent/'word-metadata.json').read_text())
  assert meta['canonical_sha256']==INPUT['word_sha256']and meta['seed']==71
  for repeat,order in enumerate(ORDERS):
   for version in order:
    tree=baseline if version=='old'else source
    run(f'sample-{repeat}-{version}',[binaries[version],'-test.run=^TestCPURegressionSample$','-test.v','-test.timeout=15m'],tree/'encoding/liftrs-shared')
  for version in('old','new'):
   profiles=cache/('profiles-'+version);profiles.mkdir();env['PIR_REGRESSION_PROFILE']=str(profiles)
   tree=baseline if version=='old'else source
   run('profile-'+version,[binaries[version],'-test.run=^TestCPURegressionSample$','-test.v','-test.timeout=15m'],tree/'encoding/liftrs-shared')
   env.pop('PIR_REGRESSION_PROFILE')
   for phase in('load-resident','open-file','native-plan','encoding'):
    run('profile-'+version+'-'+phase,[go,'tool','pprof','-top',binaries[version],profiles/(phase+'.pprof')])
  record['status']='passed'
 except BaseException as error:record['status']='failed';record['error']=repr(error);raise
 finally:
  record['elapsed_seconds']=time.monotonic()-started
  record['cache_sha256']={str(p.relative_to(cache)):sha(p)for p in cache.rglob('*')if p.is_file()}
  record['evidence_sha256']={str(p.relative_to(out)):sha(p)for p in out.rglob('*')if p.is_file()and p.name not in('results.json','results.json.tmp')};save()

if __name__=='__main__':main()
