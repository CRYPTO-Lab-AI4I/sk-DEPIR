#!/usr/bin/env python3
"""Check separated encoder paths, legacy compatibility and historical evidence."""
import argparse,hashlib,json,os,shutil,subprocess,tempfile,time
from pathlib import Path
ROOT=Path(__file__).resolve().parents[2];HERE=Path(__file__).resolve().parent
p=argparse.ArgumentParser(description=__doc__);p.add_argument('--go',default='go');args=p.parse_args()
result={'revision':subprocess.check_output(['git','rev-parse','HEAD'],cwd=ROOT,text=True).strip(),'checks':[],'status':'running','source_sha256':{str(p.relative_to(ROOT)):hashlib.sha256(p.read_bytes()).hexdigest() for p in list(ROOT.glob('encoding/**/*.go'))+list(ROOT.glob('pir/*.go'))+[HERE/'check.py']}}
def save():(HERE/'results.json').write_text(json.dumps(result,indent=2)+'\n')
save()
with (HERE/'checks.log').open('w') as log:
 def run(cmd,cwd=ROOT,env=None,output=None):
  log.write('$ '+' '.join(cmd)+'\n');log.flush();start=time.monotonic()
  target=output.open('w') if output else log
  try:r=subprocess.run(cmd,cwd=cwd,env=env,stdout=target,stderr=log)
  finally:
   if output:target.close()
  result['checks'].append({'command':cmd,'isolated_copy':cwd!=ROOT,'exit_code':r.returncode,'elapsed_s':time.monotonic()-start});save()
  if r.returncode:result['status']='failed';save();raise SystemExit('check failed; see checks.log')
 run([args.go,'test','./encoding/liftrs-shared','./encoding/liftrs-go'],env=dict(os.environ,CGO_ENABLED='0'))
 run([args.go,'test','-tags','liftrs_cpp','./encoding/liftrs-shared','./encoding/liftrs-cpp'])
 run([args.go,'vet','-tags','liftrs_cpp','./encoding/...','./pir/...'])
 with tempfile.TemporaryDirectory(prefix='depir-encoding-layout-') as directory:
  temp=Path(directory);copy=temp/'repo'
  shutil.copytree(ROOT,copy,ignore=shutil.ignore_patterns('.git','.cache','__pycache__','.DS_Store'))
  # The legacy binary is ignored and deliberately excluded by Git; build a fresh
  # one in the isolated copy rather than testing the retained local executable.
  run(['make','-C','encoding/rme_legacy','-B','rme'],cwd=copy)
  run([args.go,'test','./...'],cwd=copy)
  run([args.go,'test','-tags','liftrs_cpp','./...'],cwd=copy)
  binary=temp/'encoder';cache=temp/'small-cache'
  run([args.go,'run','./experiments/encoder-profile-s13','-s','4','-t','2','-cache',str(cache),'-profiles',str(HERE/'smoke-preparation'),'-source-manifest','encoding-layout'],output=HERE/'smoke-preparation.jsonl')
  run([args.go,'build','-tags','liftrs_cpp','-o',str(binary),'./experiments/native-encoding'])
  for mode in ['go','cpp']:
   events=HERE/(mode+'-events.jsonl')
   run([str(binary),'-backend',mode,'-s','4','-t','2','-repeats','1','-cache',str(cache),'-profiles',str(HERE/('smoke-'+mode))],output=events)
   verified=next(e for e in map(json.loads,events.read_text().splitlines()) if e['event']=='verified')
   # Immutable hashes from the pre-move seed-71 s=4 fixture, not agreement
   # between two newly changed implementations alone.
   assert verified['canonical_sha256']=='4421d6289c0b7a6f994afd8293a31dd017213a41584c56121bb9d447ac1137d4'
   assert verified['packed_sha256']=='bd7ef77fc1fff8d78829594bc46fd9342833754a3367e10b6bc88fab901ca0ee'
   assert verified['systematic_positions_checked']==97 and verified['curve_checks']==16
  run(['python3','experiments/native-encoding/analyze.py','s13'],env=dict(os.environ,PYTHONPYCACHEPREFIX='/tmp/depir-pycache'))
  run(['python3','experiments/kernel-languages/analyze.py'],env=dict(os.environ,PYTHONPYCACHEPREFIX='/tmp/depir-pycache'))
result['status']='passed';save();print('Layout, full suites, legacy build, both fixture hashes and historical evidence: PASS')
