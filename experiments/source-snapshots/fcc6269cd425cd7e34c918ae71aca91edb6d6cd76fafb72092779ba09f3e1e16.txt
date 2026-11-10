#!/usr/bin/env python3
"""Verify actual preparation evidence without rerunning expensive arithmetic."""
import json,runpy,sys
from pathlib import Path
ROOT=Path(__file__).resolve().parents[2]
HERE=Path(__file__).resolve().parent
sys.path.insert(0,str(ROOT/'experiments'))
from source_evidence import verify_sources

def main():
    if len(sys.argv)!=2 or Path(sys.argv[1]).name!=sys.argv[1]:raise SystemExit('Usage: analyze.py RUN_NAME')
    out=HERE/'runs'/sys.argv[1];r=json.loads((out/'results.json').read_text())
    if r['status']!='passed':print(json.dumps({'status':r['status'],'error':r.get('error'),'last_command':r['commands'][-1] if r['commands'] else None},indent=2));return
    sha=runpy.run_path(str(ROOT/'experiments/hpc-cached-s15/run.py'))['sha']
    sources=verify_sources(ROOT,r['source_sha256'],r['revision'])
    for path,digest in r['evidence_sha256'].items():assert sha(out/path)==digest,path
    cache=Path(r['cache'])
    for path,digest in r['cache_sha256'].items():assert sha(cache/path)==digest,path
    assert all(c['exit_code']==0 and not c['guard'] for c in r['commands'])
    workers=[out/'worker/first',out/'worker/replay'] if r['validation'] else [out/'worker']
    phases=[]
    for worker in workers:
        p=json.loads((worker/'phases.json').read_text());by_name={row['phase']:row for row in p}
        if 'preparation' in by_name:assert by_name['preparation']['identity_verified']
        summary=json.loads((worker/'gpu-summary.json').read_text());assert summary['owned_after_close']==0 and summary['product_calls']>0 and summary['calls']>0
        phases.append({'worker':str(worker.relative_to(out)),'seconds':{row['phase']:row['elapsed_ns']/1e9 for row in p},'gpu':summary})
    if r['validation']:
        for name in ('g.bin','online.bin'):assert r['cache_sha256']['first/'+name]==r['cache_sha256']['replay/'+name]
        verify=runpy.run_path(str(ROOT/'experiments/hpc-s16/run.py'))['verify_phases'];verify(out/'worker/first',10)
        checkpoints=json.loads((out/'worker/replay/checkpoints.json').read_text());assert any(c['stage']=='reuse-normalized-root' for c in checkpoints)
    else:assert r['support_sha256']=='021446aa0d2764f4020e989ee54f3a99c63c7d9a407f5babbcb333919a2854ad'
    result={'status':'passed','job_id':r['job_id'],'sources_verified':len(sources),'validation':r['validation'],'phases':phases,'limits':'Recorded identity/checkpoint/cache evidence; full actual preparation only if validation is false. No database encoding in actual mode.'}
    (out/'analysis.json').write_text(json.dumps(result,indent=2)+'\n');print(json.dumps(result,indent=2))

if __name__=='__main__':main()
