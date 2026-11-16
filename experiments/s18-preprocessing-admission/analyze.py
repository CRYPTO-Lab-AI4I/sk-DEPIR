#!/usr/bin/env python3
"""Review completed count evidence; never start/poll a GPU job."""
import argparse
import json
from pathlib import Path
import runpy

HERE = Path(__file__).resolve().parent
RUN = runpy.run_path(str(HERE/'run.py'))
MODEL = runpy.run_path(str(HERE/'memory.py'))['model']


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('name')
    args = parser.parse_args()
    out = HERE/'runs'/args.name
    record = json.loads((out/'results.json').read_text())
    result = runpy.run_path(str(HERE.parent/'s17-arithmetic/analyze.py'))['reconcile'](out)
    assert record['policy'] == RUN['POLICY']
    assert record['reference_source_sha256'] == record['source_sha256']['experiments/s18-preprocessing-admission/reference.go.txt']
    references = [json.loads(line) for line in (out/'reference.log').read_text().splitlines() if line.startswith('{')]
    assert {(r['s'], r['t']) for r in references} == {
        (s, t) for s in range(2, 8) for t in {1, (1 << s)//2-1}}
    for i, reference in enumerate(references):
        assert RUN['completed_dimension'](out/('reference-'+str(i)+'.log'))['k'] == reference['k']
    dimensions = record['dimensions']
    assert {row['s'] for row in dimensions} == {10, 16, 17, 18}
    for row in dimensions:
        assert row == RUN['completed_dimension'](out/('dimension-s'+str(row['s'])+'.log'))
        if row['s'] in RUN['KNOWN']:
            assert row['k'] == RUN['KNOWN'][row['s']]
    assert 'ERROR SUMMARY: 0 errors' in (out/'memcheck.log').read_text()
    assert 'LEAK SUMMARY: 0 bytes' in (out/'memcheck.log').read_text()
    k = next(row['k'] for row in dimensions if row['s'] == 18)
    memory = json.loads((out/'memory-model.json').read_text())
    assert memory == MODEL(18, k)
    result.update(dimensions=dimensions, memory_model=memory,
                  scope='exact s18 dimension only plus capacity models; no support/discovery/full preprocessing or multi-node implementation')
    (out/'analysis.json').write_text(json.dumps(result, indent=2)+'\n')
    print(json.dumps(result, indent=2))


if __name__ == '__main__':
    main()
