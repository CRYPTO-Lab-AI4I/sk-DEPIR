#!/usr/bin/env python3
"""Reconcile table-gap results after the user's completion ping; no scheduler calls."""
import argparse
import datetime
import hashlib
import time
import json
from pathlib import Path
import runpy

HERE = Path(__file__).resolve().parent
RUN = runpy.run_path(str(HERE / 'run.py'))
def verify_digest(path, expected):
    """Stream retained files with byte progress; these may be tens of GiB."""
    path=Path(path);size=path.stat().st_size;done=0
    began=last=time.monotonic();digest=hashlib.sha256()
    def emit(event):
        print(json.dumps({'utc':datetime.datetime.now(datetime.timezone.utc).isoformat(),
                          'phase':'verify-file','event':event,'path':str(path),
                          'completed':done,'total':size,'unit':'bytes',
                          'elapsed_seconds':time.monotonic()-began}),flush=True)
    emit('started')
    with path.open('rb') as stream:
        for block in iter(lambda:stream.read(8<<20),b''):
            digest.update(block);done+=len(block)
            if time.monotonic()-last>=30:
                emit('progress');last=time.monotonic()
    if digest.hexdigest()!=expected:
        emit('failed');raise AssertionError('checksum mismatch: '+str(path))
    emit('completed')


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('batch')
    parser.add_argument('--case', choices=RUN['CASES'])
    args = parser.parse_args()
    base = HERE / 'runs' / args.batch
    submitted = json.loads((base / 'submissions.json').read_text())
    rows = []
    digests = {}
    for case, item in submitted['cases'].items():
        if args.case and case != args.case: continue
        out = base / case
        record = json.loads((out / 'results.json').read_text())
        assert record['case'] == case and record['job_id'] == str(item['job_id'])
        cache = Path(record['cache'])
        sources = runpy.run_path(str(HERE.parent / 'source_evidence.py'))['verify_sources'](RUN['ROOT'], record['source_sha256'], record['revision'])
        for rel, digest in record['evidence_sha256'].items(): verify_digest(out / rel, digest)
        for rel, digest in record['cache_sha256'].items(): verify_digest(cache / rel, digest)
        row = {'case': case, 'job_id': record['job_id'], 'status': record['status'], 'machine_class': 'hpc', 's': record['s'], 'source_files_verified': len(sources)}
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
    assert rows, 'no selected case in submission manifest'
    result = {'status': 'passed' if all(r['status'] == 'passed' for r in rows) else 'incomplete', 'rows': rows}
    ((base / args.case if args.case else base) / 'analysis.json').write_text(json.dumps(result, indent=2) + '\n')
    print(json.dumps({'status': result['status'], 'cases': len(rows)}))


if __name__ == '__main__':
    main()
