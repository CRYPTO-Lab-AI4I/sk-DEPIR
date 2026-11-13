#!/usr/bin/env python3
"""Reconcile completed native64 checks; never inspect scheduler or other jobs."""
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
    result=VERIFY(out)
    counts={}
    for label in ('native','sanitizer'):
        rows=[json.loads(line) for line in (out/(label+'.log')).read_text().splitlines()]
        assert len(rows)==120 and all(row['match'] for row in rows)
        assert {row['s'] for row in rows}=={8,16,17,18,20}
        assert {row['n'] for row in rows}=={8,128,1024}
        for s in (8,16,17,18,20):
            for n in (8,128,1024):
                for tables in (0,1):assert sum(row['s']==s and row['n']==n and row['tables']==tables for row in rows)==4
        counts[label]=len(rows)
    assert 'ok ' in (out/'native-regressions.log').read_text()
    assert '[no test files]' in (out/'binding-regressions.log').read_text()
    result.update(native_cases=counts['native'],sanitizer_cases=counts['sanitizer'],
                  binding_scope='build only; package has no tests',
                  tests=['TestWideOrbitArithmetic','TestTableFreeFields','TestWideNativeFixtures',
                         'native forward/inverse, independent polynomial oracle, scalar/tables,32/64 agreement, canaries, invalid shape/offset',
                         'ASan/UBSan','legacy native scalar','shared liftrs_cpp Go suite'],
                  scope='bounded recursive kernels only; wide C ABI/matrix/cache/preprocessing/full s17 encoder pending')
    (out/'analysis.json').write_text(json.dumps(result,indent=2)+'\n')
    print(json.dumps({k:result[k] for k in ('status','job_id','source_files_verified','evidence_files_verified','cache_files_verified','native_cases','sanitizer_cases')}))


if __name__=='__main__':main()
