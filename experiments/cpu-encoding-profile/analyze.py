#!/usr/bin/env python3
"""Verify saved profile evidence and summarize native stage/worker scaling."""
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
    helpers=runpy.run_path(str(HERE/'run.py'));sha=helpers['sha'];cache=Path(r['cache'])
    for name,digest in r['evidence_sha256'].items():assert sha(out/name)==digest,name
    for name,digest in r['cache_sha256'].items():assert sha(cache/name)==digest,name
    enc=helpers['verify_phases'](json.loads((out/'profile/phases.json').read_text()))
    assert all(c['exit_code']==0 and c['guard'] is None for c in r['commands'])
    rows=[]
    for workers in (8,16,32):
        selected=[p for p in enc if p['workers']==workers]
        rows.append({'workers':workers,'encoding_median_seconds':statistics.median(p['elapsed_ns']/1e9 for p in selected),'sample_seconds':[p['elapsed_ns']/1e9 for p in selected],
                     'stage_median_seconds':{key:statistics.median(p['native_stages'][key]/1e9 for p in selected) for key in ('taylor_ns','columns_transposes_ns','rows_ns','buffer_ns','pointwise_ns','assembly_ns')}})
    result={'status':'passed','job_id':r['job_id'],'sources_verified':len(sources),'workers':rows,'peak_sampled_rss_bytes':max(c['peak_rss_bytes'] for c in r['commands']),
            'limits':'Two sequential profiled s14 samples per worker count, fixed order. Full canonical hash checks; no s16 memory-pressure model or CPU instruction samples. No new discovery/preparation.'}
    (out/'analysis.json').write_text(json.dumps(result,indent=2)+'\n');print(json.dumps(result,indent=2))

if __name__=='__main__':main()
