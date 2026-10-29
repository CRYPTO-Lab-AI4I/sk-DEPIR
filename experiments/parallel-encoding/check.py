#!/usr/bin/env python3
"""Run integration checks, isolating legacy tests that write CSV fixtures."""
import argparse,hashlib,json,os,shutil,subprocess,tempfile,time
from pathlib import Path
ROOT=Path(__file__).resolve().parents[2];HERE=Path(__file__).resolve().parent

def main():
 p=argparse.ArgumentParser(description=__doc__);p.add_argument('--go',default='go');a=p.parse_args();out=HERE/'validation';out.mkdir(exist_ok=False)
 result={'status':'running','revision':subprocess.check_output(['git','rev-parse','HEAD'],cwd=ROOT,text=True).strip(),'checks':[],'script_sha256':hashlib.sha256(Path(__file__).read_bytes()).hexdigest()}
 def save():(out/'verification.json').write_text(json.dumps(result,indent=2)+'\n')
 with (out/'checks.log').open('w') as log:
  def run(command,cwd=ROOT,env=None):
   log.write('$ '+' '.join(command)+'\n');log.flush();start=time.monotonic();r=subprocess.run(command,cwd=cwd,env=env,stdout=log,stderr=subprocess.STDOUT,timeout=600)
   result['checks'].append({'command':command,'isolated_copy':cwd!=ROOT,'exit_code':r.returncode,'wall_s':time.monotonic()-start});save()
   if r.returncode:raise RuntimeError('check failed: see checks.log')
  run([a.go,'test','-race','-tags','liftrs_cpp','./encoding/liftrs-shared','-run','TestCPP(Ordered|Matrix)|TestScalingOnline','-count=1'])
  run([a.go,'vet','-tags','liftrs_cpp','./encoding/...','./experiments/parallel-encoding'])
  # Standalone native regression/ABI tests exercise the unchanged v1 format too.
  with tempfile.TemporaryDirectory(prefix='depir-parallel-check-') as tmp:
   for source in ['test.cpp','c_api_test.cpp']:
    binary=str(Path(tmp)/source.replace('.cpp',''));command=[os.environ.get('CXX','clang++'),'-std=c++17','-O1','-g','-fsanitize=address,undefined','-fno-omit-frame-pointer','-pthread','-Iencoding/liftrs-cpp','encoding/liftrs-cpp/'+source]
    if source=='c_api_test.cpp':command+=['encoding/liftrs-cpp/c_api.cpp']
    run(command+['-o',binary])
    args=[binary]
    if source=='c_api_test.cpp':
     fixture=ROOT/'.cache/liftrs/language-s13/transform.bin';result['native_fixture']={'path':str(fixture.relative_to(ROOT)),'sha256':hashlib.sha256(fixture.read_bytes()).hexdigest()};args.append(str(fixture))
    run(args)
   copy=Path(tmp)/'repo';shutil.copytree(ROOT,copy,ignore=shutil.ignore_patterns('.git','.cache','__pycache__','.DS_Store'))
   run(['make','-C','encoding/rme_legacy','rme'],cwd=copy)
   run([a.go,'test','./...'],cwd=copy,env=dict(os.environ,CGO_ENABLED='0'))
   run([a.go,'test','-tags','liftrs_cpp','./...'],cwd=copy)
 result['status']='passed';save();print('parallel integration checks PASS')
if __name__=='__main__':main()
