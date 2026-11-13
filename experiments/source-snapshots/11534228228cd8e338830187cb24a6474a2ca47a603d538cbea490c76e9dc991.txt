#!/usr/bin/env python3
"""Reconcile transform component evidence; no full encoding projection."""
import json
import runpy
import sys
from pathlib import Path
ROOT=Path(__file__).resolve().parents[2]
HERE=Path(__file__).resolve().parent
sys.path.insert(0,str(ROOT/'experiments'))
from source_evidence import verify_sources

def main():
    if len(sys.argv)!=2 or Path(sys.argv[1]).name!=sys.argv[1]:raise SystemExit('Usage: analyze.py RUN_NAME')
    out=HERE/'runs'/sys.argv[1];r=json.loads((out/'results.json').read_text())
    if r['status']!='passed':print(json.dumps({'status':r['status'],'error':r.get('error')},indent=2));return
    h=runpy.run_path(str(HERE/'run.py'));cache=Path(r['cache'])
    sources=verify_sources(ROOT,r['source_sha256'],r['revision'])
    for name,digest in r['evidence_sha256'].items():assert h['sha'](out/name)==digest,name
    for name,digest in r['cache_sha256'].items():assert h['sha'](cache/name)==digest,name
    for key in ('packed_checkpoint','matrix_reference'):assert h['sha'](Path(r[key]['path']))==r[key]['sha256']
    assert all(c['exit_code']==0 and c['guard'] is None for c in r['commands'])
    matrix=[json.loads(line) for line in (out/'matrix-regression.log').read_text().splitlines()];assert len(matrix)==80 and all(row.get('match',True) for row in matrix)
    for label,small in [('backed',False),('memcheck',True)]:
        rows=h['verify'](json.loads((out/label/'backed-checks.json').read_text()),small);assert rows==r[label]
        for row in rows:
            stem='s%d-d%d-'%(row['s'],row['dimension'])
            for field,filename in [('input_sha256','input.bin'),('forward_sha256','forward.bin')]:assert h['sha'](cache/label/(stem+filename))==row[field]
    assert 'ERROR SUMMARY: 0 errors' in (out/'memcheck.log').read_text()
    summary={'status':'passed','job_id':r['job_id'],'sources_verified':len(sources),'peak_sampled_rss_bytes':max(c['peak_rss_bytes'] for c in r['commands']),'cases':r['backed'],'limitations':'Progressive component through 2^22 values, forward/inverse, fixed order and profiled. Full s16 metadata only; no end-to-end backed encoder or s17 feasibility/runtime claim.'}
    (out/'analysis.json').write_text(json.dumps(summary,indent=2)+'\n');print(json.dumps(summary,indent=2))

if __name__=='__main__':main()
