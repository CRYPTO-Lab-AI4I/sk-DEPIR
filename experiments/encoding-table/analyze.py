#!/usr/bin/env python3
"""Reconcile table-gap results after the user's completion ping; no scheduler calls."""
import argparse
import json
from pathlib import Path
import runpy

HERE = Path(__file__).resolve().parent
RUN = runpy.run_path(str(HERE / 'run.py'))
SHA = RUN['HELP']['sha']


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('batch')
    args = parser.parse_args()
    base = HERE / 'runs' / args.batch
    submitted = json.loads((base / 'submissions.json').read_text())
    rows = []
    digests = {}
    for case, item in submitted['cases'].items():
        out = base / case
        record = json.loads((out / 'results.json').read_text())
        assert record['case'] == case and record['job_id'] == str(item['job_id'])
        cache = Path(record['cache'])
        for rel, digest in record['evidence_sha256'].items(): assert SHA(out / rel) == digest, (case, rel)
        for rel, digest in record['cache_sha256'].items(): assert SHA(cache / rel) == digest, (case, rel)
        row = {'case': case, 'job_id': record['job_id'], 'status': record['status'], 'machine_class': 'hpc', 's': record['s']}
        if record['status'] == 'passed':
            phases = [r for name in record['phase_files'] for r in json.loads((out / name).read_text())]
            row['encoding'] = [r for r in phases if r['phase'].startswith(('encode-', 'encode-output-'))]
            row['setup'] = [r for r in phases if r['phase'] in ('load-online', 'online-setup')]
            row['memory'] = next(c for c in record['commands'] if c['label'] == 'workflow')
            digest = next(r['canonical_sha256'] for r in phases if 'systematic_checked' in r)
            if record['s'] in digests: assert digests[record['s']] == digest, (case, 'cross-mode hash')
            digests[record['s']] = digest
        else:
            row['error'] = record.get('error')
        rows.append(row)
    result = {'status': 'passed' if all(r['status'] == 'passed' for r in rows) else 'incomplete', 'rows': rows}
    (base / 'analysis.json').write_text(json.dumps(result, indent=2) + '\n')
    print(json.dumps({'status': result['status'], 'cases': len(rows)}))


if __name__ == '__main__':
    main()
