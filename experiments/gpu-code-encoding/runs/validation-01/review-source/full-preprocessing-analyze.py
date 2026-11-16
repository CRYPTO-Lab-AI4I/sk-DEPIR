#!/usr/bin/env python3
"""Reconcile full-job source/logs/manifests. Large independent rehash is separate."""
import argparse,json,runpy,re
from pathlib import Path
HERE=Path(__file__).resolve().parent

# Only wrapper/stage records are needed; avoid decoding millions of kernel
# events or retaining the complete progress log. Whitespace is allowed in JSON.
REVIEW_PHASES={"workflow","verify-final-cache","read-code-online.bin", "root-product",
 "normalization","reciprocal","identity","generator-spectrum","preparation-and-identities",
 "write-online.bin","write-generator.bin","recovery-roots","recovery-reciprocal",
 "device-owned-bytes","native-owned-bytes","exact-good-monomials","discovery"}
PHASE_PATTERN=re.compile(rb'"phase"\s*:\s*"(?:'+b'|'.join(re.escape(name.encode())for name in sorted(REVIEW_PHASES))+rb')"')

def review_events(path):
 rows=[]
 with Path(path).open('rb')as stream:
  for line in stream:
   if line.startswith(b'{')and PHASE_PATTERN.search(line):
    row=json.loads(line)
    if row.get('event')in('started','completed','failed','observed','write-failed'):
     rows.append(row)
 return rows

def phase_durations(rows):
 starts={};out={}
 for row in rows:
  phase=row.get('phase')
  if row.get('event')=='started':starts[phase]=row['elapsed_seconds']
  elif row.get('event')=='completed'and phase in starts:
   out.setdefault(phase,[]).append(row['elapsed_seconds']-starts.pop(phase))
 return out

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
 logs=review_events(out/'preprocessing.log')
 assert any(r.get('phase')=='workflow'and r.get('event')=='completed'for r in logs)
 assert any(r.get('phase')=='verify-final-cache'and r.get('event')=='completed'for r in logs)or any(r.get('phase')=='read-code-online.bin'and r.get('event')=='completed'for r in logs)
 assert any(r.get('phase')=='identity'and r.get('event')=='completed'for r in logs),'full polynomial identity completion absent'
 result.update(phase_seconds=phase_durations(logs),review_events=logs,scope='full CLI identities/readback plus sealed source/binary/log/manifest review; independent large-payload rehash not repeated here',stage_manifests=manifests)
 (out/'analysis.json').write_text(json.dumps(result,indent=2)+'\n');print(json.dumps(result,indent=2))
if __name__=='__main__':main()
