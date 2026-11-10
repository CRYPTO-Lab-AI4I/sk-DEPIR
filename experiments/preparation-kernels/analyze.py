#!/usr/bin/env python3
"""Check retained trial evidence and summarize matched component measurements."""
import json,runpy,statistics,sys
from pathlib import Path
ROOT=Path(__file__).resolve().parents[2]
sys.path.insert(0,str(ROOT/'experiments'))
from source_evidence import verify_sources
HELP=runpy.run_path(str(ROOT/'experiments/hpc-cached-s15/run.py'))

def analyze(mode,name):
    if mode not in ('cpu','gpu') or Path(name).name!=name:raise ValueError('invalid mode/name')
    out=ROOT/'experiments'/('preparation-kernels' if mode=='cpu' else 'gpu-transform-tiles')/'runs'/name
    record=json.loads((out/'results.json').read_text());assert record['status']=='passed',record.get('error')
    sources=verify_sources(ROOT,record['source_sha256'],record['revision'])
    for path,digest in record['evidence_sha256'].items():assert HELP['sha'](out/path)==digest,path
    cache=Path(record['cache'])
    for path,digest in record['cache_sha256'].items():assert HELP['sha'](cache/path)==digest,path
    result={'status':'passed','sources_verified':len(sources),'job_id':record['job_id'],'mode':mode}
    if mode=='cpu':
        rows=json.loads((out/'cpu-go.json').read_text())['rows'];cpp=[json.loads(x) for x in (out/'cpp-kernels.log').read_text().splitlines()]
        assert len(rows)==18 and len(cpp)==18
        kernels=[]
        for n in (32,128,512,2048,8192,32768):
            entries=[r for r in rows if r['N']==n];assert len(entries)==3
            vals={r['Method']:statistics.median(r['Seconds']) for r in entries}
            assert all(len(r['Seconds'])==3 and min(r['Seconds'])>0 for r in entries)
            native=[r for r in cpp if r['n']==n];assert len(native)==3 and all(r['match'] for r in native)
            vals['native-cpp']=statistics.median(r['seconds'] for r in native)
            kernels.append({'n':n,'seconds':vals})
        verify=runpy.run_path(str(ROOT/'experiments/hpc-s16/run.py'))['verify_phases']
        full={};reference=None
        for cutoff in (32768,1024,128):
            runs=[verify(out/f'full-{cutoff}-{i}',10) for i in range(2)]
            for r in runs:
                if reference is None:reference=r
                for key in ('k','n','canonical_sha256','packed_sha256'):assert r[key]==reference[key],key
            full[str(cutoff)]={'preparation_median_seconds':statistics.median(r['phase_seconds']['preparation'] for r in runs),'runs':runs}
        for path in ('support.bin','online.bin','g.bin'):
            assert len({record['cache_sha256'][f'full-{c}-{i}/'+path] for c in (32768,1024,128) for i in range(2)})==1,path
        result.update(kernels=kernels,full=full,fixture_sha256=record['cache_sha256']['fixtures.bin'])
    else:
        rows=[json.loads(x) for x in (out/'gpu-tiles.log').read_text().splitlines()]
        setup=rows[0];tiles=rows[1:];assert setup['kind']=='setup' and len(tiles)==27
        summary=[]
        for n in (1<<18,1<<20,1<<22):
            for tile in (16,64,256):
                rs=[r for r in tiles if r['n']==n and r['tile']==tile];assert len(rs)==3
                assert all(r['forward_match'] and r['inverse_match'] and r['owned_device_bytes']<=512*2**20 for r in rs)
                for r in rs:
                    assert r['forward_end_to_end_seconds']>=r['gpu_forward_seconds']
                med={k:statistics.median(r[k] for r in rs) for k in rs[0] if k.endswith('_seconds')}
                med['native_cpu_over_transfer_inclusive_gpu']=med['cpu_forward_seconds']/med['forward_end_to_end_seconds']
                if 'cpu_lookup_forward_seconds' in med:med['lookup_cpu_over_transfer_inclusive_gpu']=med['cpu_lookup_forward_seconds']/med['forward_end_to_end_seconds']
                summary.append({'n':n,'tile':tile,'medians':med})
        ref=json.loads((ROOT/'experiments/preparation-kernels/runs'/record['cpu_reference']/'results.json').read_text())
        assert ref['cache_sha256']['fixtures.bin']==record['fixture_sha256']
        assert HELP['sha'](Path(ref['cache'])/'fixtures.bin')==record['fixture_sha256']
        result.update(setup=setup,tiles=summary,fixture_sha256=record['fixture_sha256'])
    result['limits']='Profiled CPU kernels/two complete s10 runs per cutoff, or warmed GPU transform tiles; not a complete GPU backend, full s16 feasibility or an optimal universal cutoff.'
    (out/'analysis.json').write_text(json.dumps(result,indent=2)+'\n');return result

if __name__=='__main__':
    if len(sys.argv)==4 and sys.argv[1]=='compare':
        old,new=(analyze('cpu',name) for name in sys.argv[2:])
        assert old['fixture_sha256']==new['fixture_sha256']
        for cutoff in ('32768','1024','128'):
            for key in ('k','n','canonical_sha256','packed_sha256'):
                assert old['full'][cutoff]['runs'][0][key]==new['full'][cutoff]['runs'][0][key]
        comparison={'status':'passed','before':sys.argv[2],'after':sys.argv[3],
            'default_preparation_seconds':[r['full']['32768']['preparation_median_seconds'] for r in (old,new)],
            'limits':'Sequential same-node profiled runs, not randomized or load/frequency-controlled; all fixtures and full codewords agree.'}
        path=ROOT/'experiments/preparation-kernels/runs'/sys.argv[3]/'comparison.json'
        path.write_text(json.dumps(comparison,indent=2)+'\n');print(json.dumps(comparison,indent=2));sys.exit(0)
    if len(sys.argv)!=3:raise SystemExit('Usage: analyze.py cpu|gpu RUN_NAME')
    print(json.dumps(analyze(*sys.argv[1:]),indent=2))
