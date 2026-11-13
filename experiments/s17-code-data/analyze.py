#!/usr/bin/env python3
"""Reconcile completed code-data checks without scheduler/process inspection."""
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
    tests=['TestScalingCodeLegacy','TestScalingCodeHeaderWidths','TestScalingCodeValidation',
           'TestScalingCodeWideSlots','ExampleLoadScalingCode']
    log=(out/'code-data.log').read_text()
    for test in tests:assert '--- PASS: '+test+' ' in log,test
    for label in ('go-regressions','cpp-regressions'):
        assert (out/(label+'.log')).read_text().strip()=='PASS',label
    result.update(tests=tests,regressions=['shared Go-only','shared liftrs_cpp'],
                  scope='small real legacy round trips and bounded wide headers/slot codecs; no genuine full s17 cache/preprocessing/encoding')
    (out/'analysis.json').write_text(json.dumps(result,indent=2)+'\n')
    print(json.dumps({k:result[k] for k in ('status','job_id','source_files_verified','evidence_files_verified','cache_files_verified','regressions')}))


if __name__=='__main__':main()
