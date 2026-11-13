#!/usr/bin/env python3
"""Reconcile CPU direct packed output and matched complete file pipelines."""
import json,runpy,sys
from pathlib import Path
ROOT=Path(__file__).resolve().parents[2]
HERE=Path(__file__).resolve().parent
sys.path.insert(0,str(ROOT/'experiments'))
from source_evidence import verify_sources

def main():
    if len(sys.argv)!=2 or Path(sys.argv[1]).name!=sys.argv[1]:raise SystemExit('Usage: analyze.py RUN_NAME')
    out=HERE/'runs'/sys.argv[1];r=json.loads((out/'results.json').read_text())
    if r['status']!='passed':print(json.dumps({'status':r['status'],'error':r.get('error')},indent=2));return
    h=runpy.run_path(str(HERE/'run.py'));sha=h['sha'];cache=Path(r['cache']);sources=verify_sources(ROOT,r['source_sha256'],r['revision'])
    for key in ('evidence_sha256','cache_sha256'):
        for name,digest in r[key].items():assert sha((out if key=='evidence_sha256' else cache)/name)==digest,name
    assert sha(Path(r['component_checkpoint']['path']))==r['component_checkpoint']['sha256']
    for path,digest in r['input_fixtures'].items():assert sha(Path(path))==digest,path
    assert sha(Path(r['workflow_reference']['path'])/'online.bin')==r['workflow_reference']['online_sha256']
    assert all(c['exit_code']==0 and c['guard'] is None for c in r['commands']) and r['policy']['output_mode']=='packed'
    for label in ('small','asan'):assert h['verify_small'](json.loads((out/label/'packed-checks.json').read_text()))==r[label]
    rows=h['verify_matched'](json.loads((out/'matched/matched.json').read_text()));assert rows==r['matched']
    for row in rows:assert sha(cache/'matched'/('%s-%d.bin'%(row['mode'],row['repeat'])))==row['packed_sha256']
    summary={'status':'passed','job_id':r['job_id'],'sources_verified':len(sources),'peak_sampled_rss_bytes':max(c['peak_rss_bytes'] for c in r['commands']),'matched':rows,'limits':'Explicit direct CPU packed sink reuses full two-array transform workspace. File cache and caller-retained words remain extra; no lower transform peak RAM or fulls16/s17 latency claim. Fixed-order/profiled pipeline samples.'}
    (out/'analysis.json').write_text(json.dumps(summary,indent=2)+'\n');print(json.dumps(summary,indent=2))
if __name__=='__main__':main()
