#!/usr/bin/env python3
"""Verify analytical arithmetic, measured evidence and exact source provenance."""
import argparse
import hashlib
import json
import subprocess
import sys
from pathlib import Path

ROOT=Path(__file__).resolve().parents[2]
HERE=Path(__file__).resolve().parent
sys.path.insert(0,str(ROOT/'experiments'))
from source_evidence import verify_sources
p=argparse.ArgumentParser(description=__doc__)
p.add_argument('--go',default='go');a=p.parse_args()
checks={}
for run in ['replicas','replicas-final']:
    base=HERE/run;record=json.loads((base/'results.json').read_text())
    assert record['status']=='passed'
    checks[run]={'sources_resolved':verify_sources(ROOT,record['source_sha256'],record['revision'])}
    assert hashlib.sha256((ROOT/record['binary']['path']).read_bytes()).hexdigest()==record['binary']['sha256']
    for name,expected in record['evidence_sha256'].items():
        assert hashlib.sha256((base/name).read_bytes()).hexdigest()==expected,name
    assert max(c['process_peak_rss_bytes'] for c in record['cases'])<1<<30
subprocess.run([sys.executable,str(HERE/'model.py')],cwd=ROOT,check=True,capture_output=True,text=True)
model=json.loads((HERE/'results.json').read_text())
checks['model']={'sources_resolved':verify_sources(ROOT,model['source_sha256'],model['revision'])}
point=next(r for r in model['rows'] if r['s']==15)
assert point['membership_bitset_bytes']==128<<20
assert 21<point['identity_array_scenario_bytes']/2**30<22
assert 18<point['online_complete_array_payload_bytes']/2**30<19
command=[a.go,'vet','./experiments/preparation-memory']
result=subprocess.run(command,cwd=ROOT,capture_output=True,text=True)
checks['go_vet']={'command':command,'exit_code':result.returncode,'stdout':result.stdout,'stderr':result.stderr}
assert result.returncode==0
(HERE/'verification.json').write_text(json.dumps({'status':'passed','checks':checks,'not_run':'No library encoding tests repeated: production code is unchanged; no s15 run.'},indent=2)+'\n')
print('Preparation-memory audit verification PASS')
