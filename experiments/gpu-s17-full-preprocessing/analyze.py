#!/usr/bin/env python3
"""Reconcile full-job source/logs/manifests. Large independent rehash is separate."""
import argparse,json,runpy
from pathlib import Path
HERE=Path(__file__).resolve().parent

def main():
 p=argparse.ArgumentParser(description=__doc__);p.add_argument('name');args=p.parse_args();out=HERE/'runs'/args.name
 record=json.loads((out/'results.json').read_text())
 result=runpy.run_path(str(HERE.parent/'s17-arithmetic/analyze.py'))['reconcile'](out)
 cache=Path(record['code_cache']);manifests=record['stage_manifests'];k=record['plan']['exact_k'];n=1<<34
 for stage,expected in manifests.items():assert json.loads((cache/stage/'manifest.json').read_text())==expected
 assert manifests['support']['support']['parameters']['k']==k
 assert manifests['support']['support']['parameters']['n']==n
 assert manifests['root']['root']['support_sha256']==manifests['support']['support']['sha256']
 code=manifests['code'];assert code['header']['K']==k and code['header']['N']==n
 assert code['support_sha256']==manifests['support']['support']['sha256'] and code['root_sha256']==manifests['root']['root']['sha256']
 for stage,m in manifests.items():
  for name,r in m['files'].items():assert (cache/stage/name).stat().st_size==r['bytes']
 logs=[json.loads(s)for s in(out/'preprocessing.log').read_text().splitlines()if s.startswith('{')]
 assert any(r.get('phase')=='workflow'and r.get('event')=='completed'for r in logs)
 assert any(r.get('phase')=='verify-final-cache'and r.get('event')=='completed'for r in logs)or any(r.get('phase')=='read-code-online.bin'and r.get('event')=='completed'for r in logs)
 result.update(scope='full CLI identities/readback plus sealed source/binary/log/manifest review; independent large-payload rehash not repeated here',stage_manifests=manifests)
 (out/'analysis.json').write_text(json.dumps(result,indent=2)+'\n');print(json.dumps(result,indent=2))
if __name__=='__main__':main()
