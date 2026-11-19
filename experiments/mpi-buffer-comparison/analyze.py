#!/usr/bin/env python3
"""Recheck all successful comparison arms without changing producer outcomes."""
from pathlib import Path
import argparse
import hashlib
import json
import runpy
HERE=Path(__file__).resolve().parent
HELP=runpy.run_path(str(HERE/'run.py'))


def main():
    parser=argparse.ArgumentParser(description=__doc__);parser.add_argument('name');args=parser.parse_args()
    out=HERE/'runs'/args.name;raw=json.loads((out/'results.json').read_text())
    reconcile=runpy.run_path(str(HERE.parent/'s17-arithmetic/analyze.py'))['reconcile']
    result=reconcile(out,require_success=raw['status']=='passed',command_fields=('label','seconds','exit_code'))
    cache=Path(raw['cache']);progress=out/'comparison-progress.json'
    state=json.loads(progress.read_text())if progress.exists()else dict(status='preparation failed',arms=[],build_sha256={})
    for name,digest in state['build_sha256'].items():assert hashlib.sha256((cache/name).read_bytes()).hexdigest()==digest
    arms=[]
    for recorded in state['arms']:
        arm=HELP['review_arm'](out,cache,recorded['name'],recorded['mode']);assert arm==recorded;arms.append(arm)
    if raw['status']=='passed':
        assert state['status']=='passed'and state==raw['buffer_comparison']
        ratios=HELP['compare'](arms);assert ratios==state['comparisons']
    else:ratios=[]
    result.update(scope='Bounded same-build diagnostic crossover; original direct-buffer fault remains unexplained.',
        completed_arms=arms,comparisons=ratios,workflow_error=raw.get('error'),current_arm=state.get('current_arm'))
    (out/'analysis.json').write_text(json.dumps(result,indent=2)+'\n');print(json.dumps(result,indent=2))


if __name__=='__main__':main()
