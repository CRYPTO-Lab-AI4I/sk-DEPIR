#!/usr/bin/env python3
"""Verify exact matrix/preparation evidence and summarize bounded measurements."""
import json,runpy,statistics,sys
from pathlib import Path
ROOT=Path(__file__).resolve().parents[2]
HERE=Path(__file__).resolve().parent
sys.path.insert(0,str(ROOT/'experiments'))
from source_evidence import verify_sources

def main():
    if len(sys.argv)!=2 or Path(sys.argv[1]).name!=sys.argv[1]:raise SystemExit('Usage: analyze.py RUN_NAME')
    out=HERE/'runs'/sys.argv[1];record=json.loads((out/'results.json').read_text());assert record['status']=='passed',record.get('error')
    sha=runpy.run_path(str(ROOT/'experiments/hpc-cached-s15/run.py'))['sha']
    sources=verify_sources(ROOT,record['source_sha256'],record['revision'])
    for path,digest in record['evidence_sha256'].items():assert sha(out/path)==digest,path
    cache=Path(record['cache'])
    for path,digest in record['cache_sha256'].items():assert sha(cache/path)==digest,path
    fixtures=json.loads((out/'fixtures.json').read_text());assert len(fixtures)==6
    assert all(f['horner_checked'] for f in fixtures if f['n']<=256)
    rows=[json.loads(line) for line in (out/'matrix.log').read_text().splitlines()];assert len(rows)==80
    assert all(r.get('match',True) for r in rows)
    assert rows[-1]['kind']=='negative-checks' and rows[-1]['owned_device_bytes']<=512*2**20
    matrices=[]
    for n in (16,64,256,1<<18,1<<20,1<<22):
        gpu=[r for r in rows if r['kind']=='gpu' and r['n']==n];assert len(gpu)==3
        for row in gpu:
            for key in ('forward','inverse'):
                values=row[key];assert len(values)==8 and min(values)>=0 and sum(values[:7])<=values[7]+0.002
        result={'n':n,'gpu_forward_seconds':statistics.median(r['forward'][7] for r in gpu),'gpu_inverse_seconds':statistics.median(r['inverse'][7] for r in gpu)}
        for workers in (4,16):
            cpu=[r for r in rows if r['kind']=='cpu' and r['n']==n and r['workers']==workers];assert len(cpu)==3
            result[f'cpu_{workers}_forward_seconds']=statistics.median(r['forward_seconds'] for r in cpu)
        matrices.append(result)
    convolutions=[]
    for n in (1<<18,1<<20,1<<22):
        gpu=[r for r in rows if r['kind']=='gpu-convolution' and r['n']==n];assert len(gpu)==3
        assert any(r['kind']=='zero-convolution' and r['n']==n and r['match'] for r in rows)
        result={'n':n,'gpu_seconds':statistics.median(r['stats'][7] for r in gpu)}
        for workers in (4,16):
            cpu=[r for r in rows if r['kind']=='cpu-convolution' and r['n']==n and r['workers']==workers];assert len(cpu)==1
            result[f'cpu_{workers}_seconds']=cpu[0]['seconds']
        convolutions.append(result)
    verify=runpy.run_path(str(ROOT/'experiments/hpc-s16/run.py'))['verify_phases'];pipelines={};canonical=None
    for backend in ('cpu','gpu'):
        runs=[verify(out/f'{backend}-{repeat}',10) for repeat in range(2)];cached=verify(out/(backend+'-cached'),10)
        for run in runs+[cached]:
            digest=(run['canonical_sha256'],run['packed_sha256'])
            if canonical is None:canonical=digest
            assert digest==canonical
        if backend=='gpu':
            for repeat in range(2):
                summary=json.loads((out/f'gpu-{repeat}/gpu-summary.json').read_text())
                assert summary['calls']>0 and summary['owned_device_bytes_after_close']==0
        pipelines[backend]={'runs':runs,'cached':cached,'preparation_median_seconds':statistics.median(r['phase_seconds']['preparation'] for r in runs),'phase_sum_median_seconds':statistics.median(sum(r['phase_seconds'].values()) for r in runs)}
    for name in ('support.bin','online.bin','g.bin'):
        assert len({record['cache_sha256'][f'{b}-{i}/'+name] for b in ('cpu','gpu') for i in range(2)})==1,name
    if (out/'memcheck.log').exists():assert 'ERROR SUMMARY: 0 errors' in (out/'memcheck.log').read_text()
    result={'status':'passed','job_id':record['job_id'],'sources_verified':len(sources),'setup_seconds':rows[0]['seconds'],
            'owned_device_bytes_after_all_plans':rows[-1]['owned_device_bytes'],'matrices':matrices,'convolutions':convolutions,'pipelines':pipelines,
            'limits':'Instrumented sequential samples on one data-center GPU; GPU-assisted s10 preprocessing, CPU online encoding. Not full s16, full GPU encoder, constant-time or production validation.'}
    (out/'analysis.json').write_text(json.dumps(result,indent=2)+'\n');print(json.dumps(result,indent=2))

if __name__=='__main__':main()
