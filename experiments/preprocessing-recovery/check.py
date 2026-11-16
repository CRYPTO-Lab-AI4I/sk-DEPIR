#!/usr/bin/env python3
"""Run one bounded recovery check; longer full workloads belong in Slurm."""
import argparse,datetime,hashlib,json,os,platform,subprocess
from pathlib import Path
ROOT=Path(__file__).resolve().parents[2]
HERE=Path(__file__).resolve().parent
parser=argparse.ArgumentParser(description=__doc__)
parser.add_argument('mode',choices=('unit','race','asan','vet','controller'))
args=parser.parse_args()
go=ROOT/'.cache/liftrs/toolchains/go1.27.3-linux-amd64/go/bin/go'
env=dict(os.environ,GOTOOLCHAIN='local',GOPROXY='off',GOSUMDB='off',GOPATH=str(ROOT/'.cache/liftrs/toolchains/gopath'),GOCACHE=str(ROOT/'.cache/liftrs/preprocessing-workflow-build-cache'),GOMAXPROCS='4',GOFLAGS='-p=4')
packages=['./cmd/pir-preprocess','./encoding/liftrs-shared']
if args.mode=='controller':commands=[['python3','-B','experiments/gpu-s17-full-preprocessing/test_controller.py'],['python3','-B','experiments/gpu-s17-full-preprocessing/run.py','--check']]
elif args.mode=='vet':commands=[[str(go),'vet','-tags','liftrs_cpp',*packages]]
else:
 flags=[]if args.mode=='unit'else['-'+args.mode]
 pattern='Test.*Recovery|TestPersistent'+('|TestPreprocessing'if args.mode=='unit'else'')
 commands=[[str(go),'test',*flags,'-tags','liftrs_cpp',*packages,'-run',pattern,'-count=1']]
 if args.mode=='asan':env['ASAN_OPTIONS']='detect_leaks=0' # sandbox ptrace prevents LeakSanitizer
record={'mode':args.mode,'utc':datetime.datetime.now(datetime.timezone.utc).isoformat(),'revision':subprocess.check_output(['git','rev-parse','HEAD'],cwd=ROOT,text=True).strip(),'host':platform.node(),'commands':[]}
paths=list((ROOT/'encoding/liftrs-shared').glob('*recovery*.go'))+list((ROOT/'cmd/pir-preprocess').glob('*.go'))+[ROOT/'encoding/liftrs-shared/preprocessing.go',ROOT/'encoding/liftrs-shared/preprocessing_core.go',ROOT/'encoding/liftrs-shared/orbit_polynomial.go',Path(__file__)]
record['source_sha256']={str(p.relative_to(ROOT)):hashlib.sha256(p.read_bytes()).hexdigest()for p in paths}
for command in commands:
 try:
  done=subprocess.run(command,cwd=ROOT,env=env,capture_output=True,text=True,timeout=25)
  row={'argv':command,'exit_code':done.returncode,'stdout':done.stdout,'stderr':done.stderr}
 except subprocess.TimeoutExpired as error:row={'argv':command,'exit_code':'timeout','detail':str(error)}
 record['commands'].append(row)
record['status']='passed'if all(r['exit_code']==0 for r in record['commands'])else'failed'
(HERE/(args.mode+'-checks.json')).write_text(json.dumps(record,indent=2)+'\n')
print(json.dumps(record,indent=2))
raise SystemExit(record['status']!='passed')
