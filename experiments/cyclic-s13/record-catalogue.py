#!/usr/bin/env python3
"""Record standalone s=13 encoder support only after the guarded run validates."""
import json
from pathlib import Path
ROOT=Path(__file__).resolve().parents[2];HERE=Path(__file__).resolve().parent
r=json.loads((HERE/'results.json').read_text())
assert r['status']=='completed' and r['returncode']==0
c=r['code'];v=r['verification']
assert (c['s'],c['m'],c['t'],c['n'])==(13,2,5,67108864)
assert c['k']==v['k']==v['systematic_positions_checked'] and v['curve_checks']==16
assert v['packed_bytes']==109051904 and len(v['canonical_sha256'])==64
path=ROOT/'pirconfig/configs_catalogue.json';catalogue=json.loads(path.read_text())
e=next(x for x in catalogue['configurations'] if x['code']['s']==13)
evidence={'kind':'encoding','reference':'experiments/cyclic-s13/report.md','script':'experiments/cyclic-s13/run.py','results':'experiments/cyclic-s13/results.json',
 'method':f'Complete monomial enumeration and support discovery give exact K={c["k"]},N={c["n"]} at s=13,m=2,t=5. In-memory cyclic preprocessing is reused across three encodings. The final output passes all systematic symbols, sampled public coordinate mappings and 16 independently evaluated actual curves; canonical/packed hashes and sync/readback are recorded. Two-hour whole-case timeout and 12 GiB Go-memory guard. No large paper comparison, JSON cache loading or application-plan integration.'}
e['encoding_status']='supported';e['code']['polynomial']=0x201b
e['capacity'].update(status='verified',k=c['k'],rate=c['k']/c['n'])
e['capacity']['evidence']=[x for x in e['capacity']['evidence'] if x.get('results')!=evidence['results']]+[evidence]
e['encoder']={'implementation':'encoding/liftrs-shared/orbit.go: DiscoverOrbit, PrepareCyclicFFT (in-memory)',
 'preprocessing':'Code-only exact support discovery and cached cyclic preparation reused across databases. In-memory benchmark with two-hour timeout and 12 GiB Go-memory guard. Existing JSON loader caps remain unchanged; large-cache serialization/loading and application-plan integration were not measured.','evidence':[evidence]}
path.write_text(json.dumps(catalogue,indent=2)+'\n')
print('Recorded exact standalone s=13 cyclic evidence; retained earlier Monte Carlo evidence.')
