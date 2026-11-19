#!/usr/bin/env python3
"""Verify replay stages and inspect mismatch files without masking failed status."""
from pathlib import Path
import argparse
import hashlib
import json
import runpy
HERE=Path(__file__).resolve().parent
HELP=runpy.run_path(str(HERE/'run.py'))


def main():
    parser=argparse.ArgumentParser(description=__doc__);parser.add_argument('name');args=parser.parse_args()
    out=HERE/'runs'/args.name;raw=json.loads((out/'results.json').read_text());cache=Path(raw['cache'])
    reconcile=runpy.run_path(str(HERE.parent/'s17-arithmetic/analyze.py'))['reconcile']
    result=reconcile(out,require_success=raw['status']=='passed',command_fields=('label','seconds','exit_code'))
    state=json.loads((out/'host-progress.json').read_text())if(out/'host-progress.json').exists()else dict(arms=[])
    arms=[];payload=(cache/'replay.bin').read_bytes()
    for recorded in state['arms']:
        name=next(name for name,mode,reclaim in HELP['ARMS']if mode==recorded['mode']and reclaim==recorded['reclaim'])
        checked=HELP['review'](HELP['BASE']['parse'](out/(name+'.log')),payload,recorded['mode'],recorded['reclaim'])
        # JSON converts integer rank keys to strings.
        assert json.loads(json.dumps(checked))==recorded;arms.append(checked)
    if raw['status']=='passed':assert len(arms)==4 and state['status']=='passed'
    captures=[];current=state.get('current_arm')
    if current:
        for row in HELP['BASE']['parse'](out/(current+'.log')):
            if row.get('kind')=='proof'and not row['matches']:
                path=cache/current/f"rank-{row['rank']}"/row['file'];data=path.read_bytes()
                captures.append(dict(row,file=str(path),snapshot_sha256=hashlib.sha256(data).hexdigest(),bytes=len(data)))
    result.update(scope='Host-only MPI replay; stress reclamation is diagnostic, not a production mode or proof of cause.',completed_arms=arms,captures=captures,current_arm=current,workflow_error=raw.get('error'))
    (out/'analysis.json').write_text(json.dumps(result,indent=2)+'\n');print(json.dumps(result,indent=2))


if __name__=='__main__':main()
