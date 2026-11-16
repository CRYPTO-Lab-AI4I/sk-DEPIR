#!/usr/bin/env python3
"""Review only after the completion ping: seal checks, paired times and ranges."""
import argparse,hashlib,json,runpy,statistics,subprocess,tempfile
from pathlib import Path
HERE=Path(__file__).resolve().parent;ROOT=HERE.parents[1]
VERIFY=runpy.run_path(str(HERE.parent/'s17-arithmetic/analyze.py'))['reconcile']
PHASES=('load-resident','open-file','native-plan','workspace','warmup','encoding')
TRANSFORMS={'prefix-p-forward','prefix-message-forward','prefix-inverse','full-forward','full-inverse'}

def rows(path):
 result=[]
 for line in Path(path).read_text().splitlines():
  if line.startswith('{'):result.append(json.loads(line))
 return result

def sample(path,profiled,wanted):
 events=rows(path);result={}
 for phase in PHASES:
  completed=[r for r in events if r.get('phase')==phase and r.get('event')=='completed'and'seconds'in r]
  assert len(completed)==1,(path,phase)
  row=completed[0];assert row['profiled']==profiled and row['seconds']>=0
  result[phase]=row['seconds']
 verified=[r for r in events if r.get('phase')=='verify-word'and r.get('event')=='completed']
 assert len(verified)==2 and all(r['sha256']==wanted for r in verified),path
 native=[r for r in events if r.get('work')=='encoding-operations'and r.get('event')=='completed']
 assert TRANSFORMS<={r['phase']for r in native},(path,'missing transform/pass identifiers')
 assert sum(r['completed']==r['total']==10 for r in native)==2,(path,'incomplete native encoding')
 assert Path(path).read_text().rstrip().endswith('PASS'),path
 return result

def verify_git(revision,identities):
 names=list(identities)
 data=subprocess.check_output(['git','cat-file','--batch'],cwd=ROOT,input=''.join(f'{revision}:{p}\n'for p in names).encode())
 pos=0
 for name in names:
  end=data.index(b'\n',pos);header=data[pos:end].split();assert header[1]==b'blob',name
  size=int(header[2]);pos=end+1
  assert hashlib.sha256(data[pos:pos+size]).hexdigest()==identities[name],name
  pos+=size+1

def verify_baseline(record):
 pristine=record['baseline_pristine_sha256'];modified=record['baseline_modified_sha256']
 verify_git(record['baseline_revision'],pristine)
 changes=set(record['baseline_patch_changed'])
 assert changes=={'encoding/liftrs-cpp/c_api.cpp','encoding/liftrs-cpp/matrix.hpp','encoding/liftrs-cpp/progress.hpp'}
 for name,digest in modified.items():
  if name not in changes:assert digest==pristine[name],name
 patch=HERE/'baseline_progress.py'
 assert hashlib.sha256(patch.read_bytes()).hexdigest()==record['source_sha256']['experiments/cpu-encoding-regression/baseline_progress.py']
 instrument=runpy.run_path(str(patch))['instrument']
 with tempfile.TemporaryDirectory()as directory:
  base=Path(directory)/'old';current=Path(directory)/'new'
  for name in changes:
   tree,revision=(current,record['revision'])if name.endswith('progress.hpp')else(base,record['baseline_revision'])
   p=tree/name;p.parent.mkdir(parents=True,exist_ok=True)
   p.write_bytes(subprocess.check_output(['git','show',revision+':'+name],cwd=ROOT))
  instrument(base,current)
  for name in changes:assert hashlib.sha256((base/name).read_bytes()).hexdigest()==modified[name],name
 assert record['probe_sha256']==record['source_sha256']['experiments/cpu-encoding-regression/probe.go.txt']
 return len(pristine)

def main():
 parser=argparse.ArgumentParser(description=__doc__);parser.add_argument('name');args=parser.parse_args()
 out=HERE/'runs'/args.name;record=json.loads((out/'results.json').read_text())
 result=VERIFY(out);count=verify_baseline(record)
 fixture=json.loads((Path(record['cache'])/'fixture/input.json').read_text())
 for version in('old','new'):sample(out/('small-'+version+'.log'),False,fixture['word_sha256'])
 samples={version:[]for version in('old','new')}
 for repeat in range(4):
  for version in samples:samples[version].append(sample(out/f'sample-{repeat}-{version}.log',False,record['input']['word_sha256']))
 comparison={}
 for phase in PHASES:
  values={v:[r[phase]for r in samples[v]]for v in samples}
  medians={v:statistics.median(values[v])for v in samples}
  ratio=medians['new']/medians['old']
  comparison[phase]={'old_median_seconds':medians['old'],'new_median_seconds':medians['new'],'new_over_old':ratio,'needs_review':ratio>1.2,'samples_seconds':values,'ranges_seconds':{v:[min(values[v]),max(values[v])]for v in samples},'paired_ratios':[n/o for o,n in zip(values['old'],values['new'])]}
 for version in samples:sample(out/('profile-'+version+'.log'),True,record['input']['word_sha256'])
 result.update(baseline_files_verified=count,comparison=comparison,scope='s14 matched resident CPU encoding/setup with logging enabled; no pristine unlogged equivalence, no s17 result',tests=['small independent-Go fixture','historical full s14 canonical word hash','baseline progress-only patch replay','paired order and profile exclusion'])
 (out/'analysis.json').write_text(json.dumps(result,indent=2)+'\n')
 print(json.dumps({k:result[k]for k in('status','job_id','source_files_verified','evidence_files_verified','cache_files_verified','baseline_files_verified')},indent=2))
 for phase,c in comparison.items():print(phase,': old',round(c['old_median_seconds'],6),'s; new',round(c['new_median_seconds'],6),'s; ratio',round(c['new_over_old'],3))
if __name__=='__main__':main()
