#!/usr/bin/env python3
"""Reconcile complete or failed matrix validation evidence without scheduler calls."""
import argparse
import json
from pathlib import Path
import runpy

HERE=Path(__file__).resolve().parent
VERIFY=runpy.run_path(str(HERE.parent/'s17-arithmetic/analyze.py'))['reconcile']


def main():
    parser=argparse.ArgumentParser(description=__doc__)
    parser.add_argument('name')
    args=parser.parse_args()
    out=HERE/'runs'/args.name
    record=json.loads((out/'results.json').read_text())
    result=VERIFY(out,require_success=False)
    commands={row['label']:row for row in record['commands']}
    checks={}
    required=['binding-checks','lifetime-race','shared-regressions','matrix','matrix-sanitizer','legacy-matrix']
    for label in required:
        command=commands.get(label)
        checks[label]='not_run' if command is None else 'passed' if command['exit_code']==0 and not command['guard'] else 'failed'
    if checks['binding-checks']=='passed':
        log=(out/'binding-checks.log').read_text()
        for test in ('TestWidePlan','TestWideMatrix','TestWideMatrixParallelTaylor','TestWideCloseConcurrent','ExampleNewWideMatrix'):
            assert '--- PASS: '+test+' ' in log,test
    for label in ('native','sanitizer'):
        if label in commands and commands[label]['exit_code']==0:
            rows=[json.loads(line) for line in (out/(label+'.log')).read_text().splitlines()]
            assert len(rows)==120 and all(row['match'] for row in rows)
    for label in ('matrix','matrix-sanitizer'):
        if checks[label]=='passed':
            data=json.loads((out/(label+'.log')).read_text())
            assert data['matrix_cases']==112 and data['match']
    if checks['legacy-matrix']=='passed':
        assert 'checks passed' in (out/'legacy-matrix.log').read_text()
    complete=all(v=='passed' for v in checks.values())
    assert (record['status']=='passed')==complete
    result.update(checks=checks,error=record.get('error'),
                  tests=['wide Go/C plan and matrix oracles','parallel Taylor','Go lifetime race',
                         'shared CPU regressions','direct C matrix oracle/ASan/UBSan','legacy matrix sanitizer'],
                  scope='bounded validation only; failed or unrun stages keep the interface gate open')
    (out/'analysis.json').write_text(json.dumps(result,indent=2)+'\n')
    print(json.dumps({k:result[k] for k in ('status','job_id','source_files_verified','evidence_files_verified','cache_files_verified','checks')}))


if __name__=='__main__':main()
