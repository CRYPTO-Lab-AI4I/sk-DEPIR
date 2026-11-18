#!/usr/bin/env python3
"""Review completed publication evidence without executing the workload."""
import argparse
import json
from pathlib import Path
import runpy
import statistics
HERE=Path(__file__).resolve().parent


def main():
    parser=argparse.ArgumentParser(description=__doc__);parser.add_argument('name');args=parser.parse_args()
    out=HERE/'runs'/args.name;raw=json.loads((out/'results.json').read_text())
    result=runpy.run_path(str(HERE.parent/'s17-arithmetic/analyze.py'))['reconcile'](out)
    helper=runpy.run_path(str(HERE/'run.py'))
    rows=helper['SOURCE']['parse'](out/'persistence.log');assert rows==raw['measurements']
    measured=helper['review'](rows,Path(raw['cache'])/'probe')
    race=helper['SOURCE']['parse'](out/'race.log');assert race==raw['race_measurements']
    helper['review'](race,Path(raw['cache'])/'race-probe',small=True)
    summary=[]
    for buffer in (0,4<<20):
        for writers in (1,2,4):
            selected=[r for r in measured if r['buffer_bytes']==buffer and r['writers']==writers];assert len(selected)==2
            summary.append(dict(buffer_bytes=buffer,writers=writers,total_bytes=selected[0]['total_bytes'],
                median_publication_seconds=statistics.median(r['publication_seconds']for r in selected),
                median_readback_seconds=statistics.median(r['readback_seconds']for r in selected),
                median_publish_readback_seconds=statistics.median(r['publish_readback_seconds']for r in selected),
                file_write_calls_per_trial=sum(f['stats']['calls']for f in selected[0]['files'])))
    result.update(summary=summary,tests=['buffered/unbuffered byte/hash equivalence','producer-error publication/retry and corruption rejection',
        'real small CLI recovery/restart/Go arithmetic reference','race checks','fixed-load mode/call-size coverage',
        'all trial manifest/digest identities; full final payload bytes','source/evidence/cache seals'],
        scope='96MiB synthetic persistence/call aggregation gate; no full-size throughput, durable-power-failure guarantee, multi-node writers or new default')
    (out/'analysis.json').write_text(json.dumps(result,indent=2)+'\n');print(json.dumps(result,indent=2))


if __name__=='__main__':main()
