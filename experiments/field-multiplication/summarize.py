#!/usr/bin/env python3
"""Validate output compatibility and regenerate comparison/next-field forecasts."""
import hashlib,json,runpy,statistics
from pathlib import Path
ROOT=Path(__file__).resolve().parents[2];HERE=Path(__file__).resolve().parent
read=lambda p:json.loads(p.read_text())
current=read(HERE/'scaling-results.json')
old=read(ROOT/'experiments/cyclic-scaling/results.json')
control=[json.loads(line) for line in (HERE/'baseline-s10-events.jsonl').read_text().splitlines()]
control_phases={e['phase']:e['elapsed_ns']/1e9 for e in control if e['event']=='phase'}
control_check=next(e for e in control if e['event']=='verified')
lines=['# Fixed-constant cyclic encoding comparison','','Encoding is the median of three calls; memory is whole-case peak RSS. Baselines for s=11/12 are historical runs of the same driver; s=10 is a fresh control.','','| s / t | Setup before → after | Encoding before → after | Encoding speedup | New first use | New peak GiB |','|---|---:|---:|---:|---:|---:|']
for case in current['cases']:
 s=case['s'];before=None
 if s==10:
  before={'preparation_s':control_phases['preparation'],'encoding_s':statistics.median(v for k,v in control_phases.items() if k.startswith('encode_'))}
  check=control_check
 elif s>10:
  prior=next(c for c in old['cases'] if c['s']==s);before=prior['measurement'];check=prior['verification']
 else:
  # This earlier s=9 driver loads artifacts and uses separate processes. Show
  # it in the report as context rather than a controlled before/after speedup.
  historical=read(ROOT/'experiments/fast-polynomial/results.json')
  encoding=next(c for c in historical['scenarios'] if c['s']==9 and c['method']=='cyclic-fft')['backends']['new']['encoding']
  check={'canonical_sha256':encoding['output_sha256'],'packed_sha256':encoding['packed_sha256']}
 if check:
  for key in ['canonical_sha256','packed_sha256']:
   assert check[key]==case['verification'][key],(s,key)
 first=(case['discovery_s']+case['preparation_s']+case['encoding_s'])/60
 setup=f'{before["preparation_s"]:.2f} → {case["preparation_s"]:.2f} s' if before else f'{case["preparation_s"]:.2f} s'
 enc=f'{before["encoding_s"]:.2f} → {case["encoding_s"]:.2f} s' if before else f'{case["encoding_s"]:.3f} s'
 speed=f'{before["encoding_s"]/case["encoding_s"]:.2f}×' if before else '—'
 lines.append(f'| {s} / {case["t"]} | {setup} | {enc} | {speed} | {first:.2f} min | {case["peak_bytes"]/2**30:.3f} |')
runpy.run_path(str(ROOT/'experiments/source_evidence.py'))['verify_sources'](ROOT,current['source_sha256'],current['revision'])
for case in current['cases']:assert hashlib.sha256((HERE/case['events']).read_bytes()).hexdigest()==case['events_sha256']
assert len(current['cases'])==4,'full scaling run incomplete'
model=runpy.run_path(str(ROOT/'experiments/cyclic-scaling/run.py'))['model']
entry=next(e for e in read(ROOT/'pirconfig/configs_catalogue.json')['configurations'] if e['code']['s']==13)
anchor=current['cases'][-1]
policy=read(ROOT/'experiments/cyclic-scaling/policy.json')
forecast={k:v*policy['safety_factor'] for k,v in model({'s':13,'n':entry['capacity']['n'],'k':entry['capacity']['k']},anchor).items()}
first=sum(forecast[k] for k in ['discovery_s','preparation_s','encoding_s'])
admitted=first<policy['seconds'] and forecast['peak_bytes']<=policy['memory_gib']*2**30
result={'s':13,'t':entry['code']['t'],'capacity_kind':'catalogue_estimate','k':entry['capacity']['k'],'n':entry['capacity']['n'],'anchor_s':12,'safety_factor':policy['safety_factor'],'policy':policy,'forecast':forecast,'first_use_seconds':first,'admitted':admitted,'policy_source':'../cyclic-scaling/policy.json','model_source':'../cyclic-scaling/run.py'}
result['input_sha256']={str(p.relative_to(ROOT)):hashlib.sha256(p.read_bytes()).hexdigest() for p in [HERE/'scaling-results.json',ROOT/'pirconfig/configs_catalogue.json',ROOT/'experiments/cyclic-scaling/run.py',ROOT/'experiments/cyclic-scaling/policy.json']}
(HERE/'s13-forecast.json').write_text(json.dumps(result,indent=2)+'\n')
(HERE/'table.md').write_text('\n'.join(lines)+'\n')
print('\n'.join(lines));print('s13 forecast:',first/60,'min;',forecast['peak_bytes']/2**30,'GiB; admitted:',admitted)
