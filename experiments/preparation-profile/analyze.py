#!/usr/bin/env python3
"""Verify bounded profile evidence without treating kernels as a whole run."""
import json
from pathlib import Path
import runpy
import sys

HERE=Path(__file__).resolve().parent
ROOT=HERE.parents[1]
sys.path.insert(0,str(ROOT/'experiments'))
from source_evidence import verify_sources


def main():
    if len(sys.argv)!=2 or Path(sys.argv[1]).name!=sys.argv[1]: raise SystemExit('Usage: analyze.py RUN_NAME')
    out=HERE/'runs'/sys.argv[1]
    record=json.loads((out/'results.json').read_text())
    if record['status']!='passed':
        print(json.dumps({'status':record['status'],'error':record.get('error'),'commands':record['commands'][-1:]},indent=2));return
    sha=runpy.run_path(str(ROOT/'experiments/hpc-cached-s15/run.py'))['sha']
    sources=verify_sources(ROOT,record['source_sha256'],record['revision'])
    assert sha(Path(record['binary']))==record['binary_sha256']
    for name,digest in record['evidence_sha256'].items(): assert sha(out/name)==digest,name
    stages=json.loads((out/'stages.json').read_text());by={s['phase']:s for s in stages}
    assert len(by)==len(stages)==12
    assert by['support-read']['sampled']==65536
    assert by['support-read']['sha256']==record['support_reference']['sha256']
    assert by['bounded-setup']['maximum_transform_points']==1<<22
    assert by['factor-verification']['roots_checked']==by['root-generation']['factors']
    assert by['factor-verification']['product_evaluation_points']==3
    for log in (18,20,22):
        conv=by['convolution-'+str(log)];check=by['verify-convolution-'+str(log)]
        assert conv['transform_points']==1<<log
        assert conv['output_coefficients']==(1<<log)-1
        assert conv['native_transform_calls']==3
        assert 0<conv['native_transform_seconds']<=conv['seconds']+0.001
        assert check['base_range_checked']==(1<<log)-1
        assert check['evaluation_points']==3 and check['direct_prefix_checked']==32
    worker=next(c for c in record['commands'] if c['label']=='profile')
    assert worker['exit_code']==0 and worker['guard'] is None
    analysis={'status':'passed','job_id':record['job_id'],'sources_verified':len(sources),
        'worker_peak_rss_bytes':worker['peak_rss_bytes'],'stages':stages,
        'limits':'Representative s16 kernels at different sizes, not full-stage proportions, full s16 feasibility, CPU scaling or GPU benefit.'}
    (out/'analysis.json').write_text(json.dumps(analysis,indent=2)+'\n')
    print(json.dumps(analysis,indent=2))


if __name__=='__main__':main()
