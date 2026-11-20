#!/usr/bin/env python3
"""Audit bounded transpose shards and packets, keeping failed stages failed."""
from pathlib import Path
import argparse
import json
import runpy
HERE=Path(__file__).resolve().parent
HELP=runpy.run_path(str(HERE/'run.py'))


def main():
    parser=argparse.ArgumentParser(description=__doc__);parser.add_argument('name');args=parser.parse_args()
    out=HERE/'runs'/args.name;raw=json.loads((out/'results.json').read_text());cache=Path(raw['cache'])
    reconcile=runpy.run_path(str(HERE.parent/'s17-arithmetic/analyze.py'))['reconcile']
    result=reconcile(out,require_success=raw['status']=='passed',command_fields=('label','seconds','exit_code'))
    state=json.loads((out/'transpose-progress.json').read_text())if(out/'transpose-progress.json').exists()else dict(arms=[])
    checked=[]
    for arm in state['arms']:
        mode=arm['mode'];value=HELP['review'](HELP['BASE']['parse_rank_logs'](out/(mode+'.log')),cache/mode,mode,raw['resources']['diagnostic_transport'])
        assert json.loads(json.dumps(value))==arm;checked.append(value)
    if raw['status']=='passed':assert len(checked)==2 and state['status']=='passed'and state==raw['distributed_transpose']
    result.update(completed_arms=checked,current_arm=state.get('current_arm'),workflow_error=raw.get('error'),scope='Host packing/transport transpose reference, not a distributed FFT or fulls18 preprocessing.')
    (out/'analysis.json').write_text(json.dumps(result,indent=2)+'\n');print(json.dumps(result,indent=2))


if __name__=='__main__':main()
