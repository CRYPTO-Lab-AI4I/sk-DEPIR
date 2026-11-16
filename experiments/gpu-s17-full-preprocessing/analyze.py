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

def write_summary(out,record,result):
 cache=Path(record['code_cache']);phases=result['phase_seconds']
 worker=next(c for c in record['commands']if c['label']=='preprocessing')
 canonical=sum(f['bytes']for m in record['stage_manifests'].values()for f in m['files'].values())
 recovery=sum(f.stat().st_size for p in cache.glob('recovery-*')for f in p.rglob('*')if f.is_file())
 old_log=out.parent/'first-01'/'preprocessing.log'
 discovery=phase_durations(review_events(old_log)).get('discovery',[None])[0]if old_log.exists()else None
 saved=phases.get('recovery-roots',[])
 summary=dict(job_id=record['job_id'],status=record['status'],s=17,m=2,t=5,k=record['plan']['exact_k'],n=1<<34,
  worker_seconds=worker['seconds'],controller_seconds=record['elapsed_seconds'],
  discovery_seconds_in_predecessor=discovery,discovery_samples=record['stage_manifests']['support']['support']['parameters']['samples'],
  construction_identity_spectrum_seconds=phases['preparation-and-identities'][0],root_product_seconds=phases['root-product'][0],
  reciprocal_seconds=phases['reciprocal'][0],identity_seconds=phases['identity'][0],spectrum_seconds=phases['generator-spectrum'][0],
  final_readback_seconds=phases['verify-final-cache'][0],peak_sampled_rss_bytes=worker['peak_rss_bytes'],peak_sampled_rss_gib=worker['peak_rss_bytes']/(1<<30),
  canonical_payload_bytes=canonical,canonical_payload_gib=canonical/(1<<30),
  retained_recovery_bytes_including_manifests=recovery,retained_recovery_gib=recovery/(1<<30),
  checkpoint_writes=len(saved),checkpoint_seconds=sum(saved),checkpoint_wall_fraction=sum(saved)/worker['seconds'],
  gpu_device_peak='not sampled; three product arrays model384GiB aggregate before metadata',
  gpu_final_observed_owned_bytes=next(r['completed']for r in result['review_events']if r['phase']=='device-owned-bytes'),
  large_payload_rehash_in_review=False,
  notes=['Resumed preparation reused certified support from75819; not one uninterrupted fresh benchmark.',
   'Full P*G identity and final complete-cache readback passed in worker.',
   'Full large recovery snapshots were written, not used for a restart in this run.',
   'Raw progress log3.2GiB is retained outside Git; overhead unquantified and included.'])
 (out/'summary.json').write_text(json.dumps(summary,indent=2)+'\n')
 return summary

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
 write_summary(out,record,result)
 (out/'analysis.json').write_text(json.dumps(result,indent=2)+'\n');print(json.dumps(result,indent=2))
if __name__=='__main__':main()
