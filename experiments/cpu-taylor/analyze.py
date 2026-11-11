#!/usr/bin/env python3
"""Verify old/new Taylor comparison and summarize complete-encoding effects."""
import json,runpy,statistics,sys
from pathlib import Path
ROOT=Path(__file__).resolve().parents[2]
HERE=Path(__file__).resolve().parent
sys.path.insert(0,str(ROOT/'experiments'))
from source_evidence import verify_sources

def main():
    if len(sys.argv)!=2 or Path(sys.argv[1]).name!=sys.argv[1]:raise SystemExit('Usage: analyze.py RUN_NAME')
    out=HERE/'runs'/sys.argv[1];r=json.loads((out/'results.json').read_text())
    if r['status']!='passed':print(json.dumps({'status':r['status'],'error':r.get('error')},indent=2));return
    sources=verify_sources(ROOT,r['source_sha256'],r['revision'])
    h=runpy.run_path(str(HERE/'run.py'));sha=h['sha'];cache=Path(r['cache'])
    for path,digest in r['evidence_sha256'].items():assert sha(out/path)==digest,path
    for path,digest in r['cache_sha256'].items():assert sha(cache/path)==digest,path
    assert '72 Taylor reference/forward/inverse checks passed' in (out/'asan.log').read_text()
    assert all(c['exit_code']==0 and c['guard'] is None for c in r['commands'] if c['label']!='tsan')
    rows=[]
    for workers in (16,32):
        row={'workers':workers}
        for variant in ('baseline','candidate'):
            enc=[h['verify_case'](json.loads((out/f'{variant}-w{workers}-r{i}/phases.json').read_text())) for i in range(2)]
            row[variant]={'seconds':statistics.median(p['elapsed_ns']/1e9 for p in enc),'samples_seconds':[p['elapsed_ns']/1e9 for p in enc],
                          'stage_seconds':{key:statistics.median(p['native_stages'][key]/1e9 for p in enc) for key in ('taylor_ns','columns_transposes_ns','rows_ns','buffer_ns','pointwise_ns','assembly_ns')}}
        row['encoding_speedup']=row['baseline']['seconds']/row['candidate']['seconds'];row['taylor_speedup']=row['baseline']['stage_seconds']['taylor_ns']/row['candidate']['stage_seconds']['taylor_ns'];rows.append(row)
    result={'status':'passed','job_id':r['job_id'],'sources_verified':len(sources),'baseline_revision':r['baseline_revision'],'tsan':r['tsan'],'workers':rows,'peak_sampled_rss_bytes':max(c['peak_rss_bytes'] for c in r['commands']),
            'limits':'Two interleaved profiled samples per worker/variant; full s14 canonical hashes. No s16 performance/pressure or native instruction attribution.'}
    (out/'analysis.json').write_text(json.dumps(result,indent=2)+'\n');print(json.dumps(result,indent=2))

if __name__=='__main__':main()
