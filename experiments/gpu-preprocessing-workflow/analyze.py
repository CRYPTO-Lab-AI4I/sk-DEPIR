#!/usr/bin/env python3
"""Reconcile completed persistent workflow validation, only after completion ping."""
import argparse,json,runpy
from pathlib import Path
HERE=Path(__file__).resolve().parent

def main():
 p=argparse.ArgumentParser(description=__doc__);p.add_argument('name');args=p.parse_args();out=HERE/'runs'/args.name
 record=json.loads((out/'results.json').read_text())
 result=runpy.run_path(str(HERE.parent/'s17-arithmetic/analyze.py'))['reconcile'](out)
 if record.get('peer_probe'):
  probe=[json.loads(s)for s in(out/'peer-probe.log').read_text().splitlines()if s.startswith('{')]
  assert all(any(r['phase']==phase and r['event']=='completed'for r in probe)for phase in('runtime-peer-copies','peer-transpose-probe'))
 assert len(record['cases'])==4
 assert {(c['layout'],c['gpus'])for c in record['cases']}=={(layout,n)for layout in('peer','local')for n in(1,4)}
 assert all(c['canonical_sha256']==record['independent_go_reference']['sha256']for c in record['cases'])
 for c in record['cases']:
  label=c['layout']+'-'+str(c['gpus'])
  def events(suffix):return [json.loads(s)for s in(out/(label+suffix+'.log')).read_text().splitlines()if s.startswith('{')]
  fresh=events('-fresh');cached=events('-cached');root=events('-root-restart');plain=events('-cuda-free-readback')
  for rows in(fresh,cached,root,plain):assert any(r.get('phase')=='workflow'and r.get('event')=='completed'for r in rows)
  assert any(r.get('phase')=='discovery'and r.get('event')=='completed'for r in fresh)
  assert not any(r.get('phase')in('setup','discovery')for r in cached+plain)
  assert not any(r.get('phase')=='discovery'for r in root)
  assert any(r.get('phase')=='read-root-root.bin'for r in root)
  assert any(r.get('unit')=='bytes'and r.get('completed',0)>0 for r in fresh)
  directory=Path(c['cache']);assert len(list(directory.glob('cpu-*.pprof')))>=4
  for stage in('support','root','code'):
   manifest=json.loads((directory/stage/'manifest.json').read_text())
   assert manifest['producer']['revision']==record['revision']
   assert manifest['producer']['gpu_plugin_sha256'] and manifest['producer']['executable_sha256']
 assert 'ERROR SUMMARY: 0 errors'in(out/'memcheck.log').read_text()
 assert 'LEAK SUMMARY: 0 bytes'in(out/'memcheck.log').read_text()
 result.update(tests=['atomic stage publication and failure retention','independent Go canonical bytes','fresh and completed-cache/root restarts','CUDA-free final-cache readback','corruption/source-binding rejection','race/ASan/vet','fresh executable CUDA memcheck'],cases=record['cases'],scope='bounded complete persistent GPU preprocessing; not full s17 admission or latency')
 (out/'analysis.json').write_text(json.dumps(result,indent=2)+'\n');print(json.dumps(result,indent=2))
if __name__=='__main__':main()
