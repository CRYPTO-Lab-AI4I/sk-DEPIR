#!/usr/bin/env python3
"""Verify partial-prefix measurements without claiming a complete s16 result."""
import json,runpy,sys
from pathlib import Path
ROOT=Path(__file__).resolve().parents[2];HERE=Path(__file__).resolve().parent
sys.path.insert(0,str(ROOT/'experiments'))
from source_evidence import verify_sources

def main():
    if len(sys.argv)!=2 or Path(sys.argv[1]).name!=sys.argv[1]:raise SystemExit('Usage: analyze.py RUN_NAME')
    out=HERE/'runs'/sys.argv[1];record=json.loads((out/'results.json').read_text());assert record['status']=='passed',record.get('error')
    sha=runpy.run_path(str(ROOT/'experiments/hpc-cached-s15/run.py'))['sha'];sources=verify_sources(ROOT,record['source_sha256'],record['revision'])
    for name,digest in record['evidence_sha256'].items():assert sha(out/name)==digest,name
    for name,digest in record['cache_sha256'].items():assert sha(Path(record['cache'])/name)==digest,name
    assert sha(Path(record['library']))==record['library_sha256']
    census=json.loads((out/'census.json').read_text());assert sum(census['batch_degrees'])==census['k']==918672421
    assert census['n']==1<<32 and census['prefix_positions']==1<<20
    assert census['accepted_factors']*2-census['fixed_factors']==census['k']
    assert len(census['fixed_factor_ordinals'])==census['fixed_factors']
    stages=json.loads((out/'stages.json').read_text());phases={s['phase']:s for s in stages};assert len(phases)==13
    partials={};reference=None
    for backend in ('cpu','resident','resident-4096'):
        partial=json.loads((out/(backend+'-partial.json')).read_text());assert partial['prefix_positions']==1<<20
        if reference is None:reference=(partial['coefficients'],partial['sha256'])
        assert (partial['coefficients'],partial['sha256'])==reference
        for interval in range(2):
            item=phases[f'{backend}-interval-{interval}'];assert item['start_position']==interval*(1<<19) and item['end_position']==(interval+1)*(1<<19)
        partials[backend]={'interval_seconds':[phases[f'{backend}-interval-{i}']['seconds'] for i in range(2)],
                           'finish_seconds':phases[backend+'-finish']['seconds'],'partial':partial}
    inventory=runpy.run_path(str(HERE/'inventory.py'))['inventory'](census)
    (out/'inventory.json').write_text(json.dumps(inventory,indent=2)+'\n')
    result={'status':'passed','sources_verified':len(sources),'job_id':record['job_id'],'partials':partials,
            'estimation_status':inventory['estimation_status'],'limits':'One actual prefix in two consecutive intervals per backend; not a complete code, randomized sample or full-run forecast.'}
    (out/'analysis.json').write_text(json.dumps(result,indent=2)+'\n');print(json.dumps(result,indent=2))

if __name__=='__main__':main()
