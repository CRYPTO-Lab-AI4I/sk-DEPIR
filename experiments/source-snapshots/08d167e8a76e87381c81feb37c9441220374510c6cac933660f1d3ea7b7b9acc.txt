#!/usr/bin/env python3
"""Verify matched outputs, retained sources/constants and stage measurements."""
import hashlib
import json
from pathlib import Path
import runpy
import statistics
import sys

HERE=Path(__file__).resolve().parent
ROOT=HERE.parents[1]
sys.path.insert(0,str(ROOT/'experiments'))
from source_evidence import verify_sources


def sha(path):
    h=hashlib.sha256()
    with path.open('rb') as f:
        for block in iter(lambda:f.read(1<<20),b''): h.update(block)
    return h.hexdigest()


def main():
    if len(sys.argv)!=2 or Path(sys.argv[1]).name!=sys.argv[1]: raise SystemExit('Usage: analyze.py RUN_NAME')
    out=HERE/'runs'/sys.argv[1]
    result=json.loads((out/'results.json').read_text())
    assert result['status']=='passed'
    for name,digest in result['evidence_sha256'].items(): assert sha(out/name)==digest,name
    cache=Path(result['cache'])
    verify=runpy.run_path(str(ROOT/'experiments/hpc-s16/run.py'))['verify_phases']
    summary={}; all_runs=[]; cache_hashes={}
    for variant in ('baseline','optimized'):
        sources=verify_sources(ROOT,result[variant+'_sources'],result['revision'])
        assert sha(cache/(variant+'.test'))==result[variant+'_binary_sha256']
        leaf=json.loads((out/(variant+'-leaves')/'leaves.json').read_text())
        runs=[]
        for repeat in range(3):
            phases=verify(out/(variant+'-'+str(repeat)),10)
            runs.append(phases); all_runs.append(phases)
            constants=cache/(variant+'-'+str(repeat))
            for name in ('basis-metadata.json','online-metadata.json','word-metadata.json'):
                meta=json.loads((constants/name).read_text())
                # Metadata keys differ by phase; phase records also carry hashes.
                if 'path' in meta and 'sha256' in meta:
                    assert sha(constants/meta['path'])==meta['sha256']
            for p in constants.iterdir():
                if p.is_file(): cache_hashes[str(p.relative_to(cache))]=sha(p)
        summary[variant]={'sources_verified':len(sources),'leaf':leaf,'runs':runs,
            'preparation_median_seconds':statistics.median(r['phase_seconds']['preparation'] for r in runs),
            'discovery_to_readback_median_seconds':statistics.median(sum(r['phase_seconds'].values()) for r in runs)}
    reference=all_runs[0]
    for r in all_runs:
        for key in ('k','n','canonical_sha256','packed_sha256'): assert r[key]==reference[key],key
    assert summary['baseline']['leaf']['sha256']==summary['optimized']['leaf']['sha256']
    # Public preprocessing constants must match as well as the final codeword.
    for name in ('support.bin','online.bin','g.bin'):
        assert len({cache_hashes[v+'-'+str(i)+'/'+name] for v in summary for i in range(3)})==1,name
    analysis={'status':'passed','variants':summary,'cache_sha256':cache_hashes,
        'preparation_speedup':summary['baseline']['preparation_median_seconds']/summary['optimized']['preparation_median_seconds'],
        'leaf_speedup':summary['baseline']['leaf']['seconds']/summary['optimized']['leaf']['seconds'],
        'limits':'Profiled s16 leaf kernel and complete s10 single-node runs, not full s16 feasibility. Variant order is baseline then optimized; no unprofiled or cold-I/O claim.'}
    (out/'analysis.json').write_text(json.dumps(analysis,indent=2)+'\n')
    print(json.dumps(analysis,indent=2))


if __name__=='__main__':main()
