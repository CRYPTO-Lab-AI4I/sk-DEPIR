#!/usr/bin/env python3
"""Reconcile completed transform-component evidence; no job or database run."""
import json,runpy,statistics,sys
from pathlib import Path
ROOT=Path(__file__).resolve().parents[2]
HERE=Path(__file__).resolve().parent
sys.path.insert(0,str(ROOT/'experiments'))
from source_evidence import verify_sources

def main():
    if len(sys.argv)!=2 or Path(sys.argv[1]).name!=sys.argv[1]:raise SystemExit('Usage: analyze.py RUN_NAME')
    out=HERE/'runs'/sys.argv[1];r=json.loads((out/'results.json').read_text())
    if r['status']!='passed':print(json.dumps({'status':r['status'],'error':r.get('error')}));return
    source=verify_sources(ROOT,r['source_sha256'],r['revision']);h=runpy.run_path(str(HERE/'run.py'));sha=h['HELP']['sha'];cache=Path(r['cache'])
    for name,digest in r['cache_sha256'].items():assert sha(cache/name)==digest,name
    for name,digest in r['evidence_sha256'].items():assert sha(out/name)==digest,name
    assert sha(Path(r['old_library']['path']))==r['old_library']['sha256']
    assert all(c['exit_code']==0 and c['guard'] is None for c in r['commands'])
    assert 'ERROR SUMMARY: 0 errors' in (out/'memcheck.log').read_text()
    layout=r['policy']['layout'];rows=h['verify_transforms'](out/'transforms',layout)
    h['verify_transforms'](out/'memcheck',layout,True)
    by={(row['dimension'],row['layout']):row for row in rows};summary=[]
    for row in rows:
        item={k:row[k] for k in ('dimension','n','layout','owned_bytes','working_array_bytes','setup_seconds')}
        item['forward_median_seconds']=statistics.median(row['forward_seconds']);item['inverse_median_seconds']=statistics.median(row['inverse_seconds'])
        if layout=='compare' and row['layout']=='direct':
            baseline=by[row['dimension'],'fast']
            item['forward_time_ratio_to_fast']=item['forward_median_seconds']/statistics.median(baseline['forward_seconds'])
            item['inverse_time_ratio_to_fast']=item['inverse_median_seconds']/statistics.median(baseline['inverse_seconds'])
            item['saved_device_bytes']=baseline['owned_bytes']-row['owned_bytes']
        summary.append(item)
    result={'status':'passed','job_id':r['job_id'],'sources_verified':len(source),'measurements':summary,'peak_sampled_rss_bytes':max(c['peak_rss_bytes'] for c in r['commands']),
            'limits':'Transform component in s16 arithmetic through2^25 values; no fulls16 database encoding or s17 arithmetic/capacity claim. Fixed-order fast/direct comparison, three repeats. Direct mode may trade runtime for scratch memory and remains explicitly selected. Online output assembly is a separate memory gate; CPU job73586 was not inspected.'}
    (out/'analysis.json').write_text(json.dumps(result,indent=2)+'\n');print(json.dumps(result,indent=2))

if __name__=='__main__':main()
