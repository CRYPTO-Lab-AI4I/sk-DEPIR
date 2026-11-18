#!/usr/bin/env python3
"""Review completed staged GPU exchange without executing GPU/MPI work."""
import argparse
import json
from pathlib import Path
import runpy
import statistics
HERE=Path(__file__).resolve().parent


def summarize(rows):
    result=[]
    for size in (1<<20,8<<20,32<<20):
        samples=[r for r in rows if r.get('kind')=='measurement' and r['rank']==0 and r['bytes']==size and r['sample']>=0]
        assert len(samples)==3
        names=('host_seconds','producer_seconds','d2h_seconds','exchange_seconds','h2d_seconds','consumer_seconds','pipeline_seconds')
        medians={name:statistics.median(r[name]for r in samples)for name in names}
        result.append(dict(input_MiB_per_rank=size/(1<<20),remote_MiB_per_node=size/2/(1<<20),median_seconds=medians,
            remote_payload_GiBs_per_node=size/2/(1<<30)/medians['exchange_seconds'],
            effective_pipeline_payload_GiBs_per_node=size/2/(1<<30)/medians['pipeline_seconds']))
    return result


def main():
    parser=argparse.ArgumentParser(description=__doc__);parser.add_argument('name');args=parser.parse_args()
    out=HERE/'runs'/args.name;record=json.loads((out/'results.json').read_text())
    result=runpy.run_path(str(HERE.parent/'s17-arithmetic/analyze.py'))['reconcile'](out,command_fields=('label','seconds','exit_code'))
    helper=runpy.run_path(str(HERE/'run.py'))
    rows=helper['parse'](out/'normal.log');helper['verify'](rows)
    sanitized=helper['parse'](out/'memcheck.log');helper['verify'](sanitized,small=True)
    assert rows==record['measurements'] and sanitized==record['sanitized']
    text=(out/'memcheck.log').read_text()
    assert text.count('ERROR SUMMARY: 0 errors')==2 and text.count('LEAK SUMMARY: 0 bytes')==2
    result.update(summary=summarize(rows),command_RSS='not sampled; rank RSS is recorded in measurements, not an aggregate process-group peak',tests=['canonical Go field fixture','full polynomial-basis CPU producer oracle',
        'receiving GPU ownership/inverse plus full CPU output','full reverse ownership exchange',
        'two distinct nodes and paired timing coverage','two-rank CUDA memcheck and retained native/fixture profiles'],
        scope='two-node host-staged GPU communication component; no GPU-direct, distributed FFT/root pipeline, four-node or fullsize runtime')
    (out/'analysis.json').write_text(json.dumps(result,indent=2)+'\n');print(json.dumps(result,indent=2))


if __name__=='__main__':main()
