#!/usr/bin/env python3
"""Verify primary CPU/CUDA encoding validation and reusable setup boundaries."""
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
    sources=verify_sources(ROOT,r['source_sha256'],r['revision']);sha=runpy.run_path(str(ROOT/'experiments/hpc-cached-s15/run.py'))['sha'];cache=Path(r['cache'])
    for path,digest in r['evidence_sha256'].items():assert sha(out/path)==digest,path
    for path,digest in r['cache_sha256'].items():assert sha(cache/path)==digest,path
    assert all(c['exit_code']==0 and c['guard'] is None for c in r['commands'])
    cases=json.loads((out/'online-checks/online-checks.json').read_text());assert [c['s'] for c in cases]==[2,3,4,6,8]
    assert all(c['matches']==4 and c['closed'] and c['invalid_messages'] and c['owned_bytes']<=512*2**20 for c in cases)
    assert 'ERROR SUMMARY: 0 errors' in (out/'memcheck.log').read_text()
    matrix=[json.loads(line) for line in (out/'matrix-regression.log').read_text().splitlines()];assert len(matrix)==80 and all(row.get('match',True) for row in matrix)
    baseline=r['matrix_reference'];assert sha(Path(baseline['path']))==baseline['sha256']
    old=r['old_online_library'];assert sha(Path(old['path']))==old['sha256']
    workflow={};canonical=None;packed=None
    for backend in ('cpu','cuda'):
        phases=json.loads((out/'workflow'/backend/'phases.json').read_text());by={p['phase']:p for p in phases}
        for i in (0,1):
            check=by[f'verify-{i}'];assert check['systematic_checked']>0 and check['curves_checked']==16 and check['repeat_match']
            if canonical is None:canonical=check['canonical_sha256']
            assert check['canonical_sha256']==canonical
        assert by['verify-packed']['positions_checked']==1027 and by['verify-packed']['readback_match']
        meta=json.loads((cache/'workflow'/backend/'word-metadata.json').read_text())
        if packed is None:packed=meta['sha256']
        assert meta['sha256']==packed and meta['canonical_sha256']==canonical
        enc=[by[f'encode-{i}'] for i in (0,1)]
        workflow[backend]={'setup_seconds':by['online-setup']['elapsed_ns']/1e9,'encode_seconds':[p['elapsed_ns']/1e9 for p in enc]}
        if backend=='cpu':
            for p in enc:
                stats=p['native_stages'];assert stats['calls']==5 and stats['boundary_calls']==1
                assert sum(stats[k] for k in ('taylor_ns','columns_transposes_ns','rows_ns','buffer_ns','pointwise_ns','assembly_ns'))<=p['native_after']['steady_ns']-p['native_before']['steady_ns']+1_000_000
        if backend=='cuda':
            for p in enc:
                stats=p['cuda_stages'];assert stats['owned_device_bytes']<=512*2**20 and stats['native_total_ns']<=stats['elapsed_ns']+1_000_000
    result={'status':'passed','job_id':r['job_id'],'sources_verified':len(sources),'small_s_cases':cases,'workflow_s10':workflow,
            'limits':'Independent small Go word comparisons, old-library rejection, CPU-only/CUDA-only/combined builds and complete s10 workflow checks. Not yet full s16 CUDA encoding, production security or constant-time arithmetic.'}
    (out/'analysis.json').write_text(json.dumps(result,indent=2)+'\n');print(json.dumps(result,indent=2))

if __name__=='__main__':main()
