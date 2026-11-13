#!/usr/bin/env python3
"""Reconcile private buffers, file I/O and progressive CPU disk transforms."""
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
    assert all(c['exit_code']==0 and c['guard'] is None for c in r['commands']) and r['policy']['memory_mode']=='disk'
    for label,small in [('disk',False),('asan',True)]:
        rows=h['verify'](json.loads((out/label/'disk-checks.json').read_text()),small);assert rows==r[label]
        for row in rows:
            path=Path(r['reference']['path'])/('s%d-d%d-forward.bin'%(row['s'],row['dimension']));assert sha(path)==row['forward_sha256']
    summary={'status':'passed','job_id':r['job_id'],'sources_verified':len(sources),'cases':r['disk'],'peak_sampled_rss_bytes':max(c['peak_rss_bytes'] for c in r['commands']),'job_cgroup':r['job_cgroup'],'retained_file_cache_bytes':r['retained_file_cache_bytes'],'limits':'Private RAM tiles only: filesystem cache/validation arrays extra. Shared-filesystem component through2^22; cache use is not isolated; no physical-disk/full encoding/s16/s17 runtime claim. Full online integration still pending.'}
    (out/'analysis.json').write_text(json.dumps(summary,indent=2)+'\n');print(json.dumps(summary,indent=2))
if __name__=='__main__':main()
