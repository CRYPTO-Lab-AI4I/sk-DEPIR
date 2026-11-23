#!/usr/bin/env python3
"""Reconcile the completed archive scan without rereading its 29 GiB input."""
import hashlib
import json
from pathlib import Path
import re
import subprocess

HERE = Path(__file__).resolve().parent
ROOT = HERE.parents[1]
receipt = json.loads((HERE/'submission.json').read_text())
output = ROOT/receipt['results']
metadata = json.loads((output/'metadata.json').read_text())
result = json.loads((output/'owners.json').read_text())
assert receipt['job_id'] == metadata['job_id'] == '82200'
assert metadata['status'] == 'passed'
assert metadata['revision'] == receipt['submission_revision']
source = subprocess.check_output(['git', 'show', metadata['revision']+':experiments/distributed-fullsize-plan/scan_support.cpp'], cwd=ROOT)
assert hashlib.sha256(source).hexdigest() == metadata['source_sha256']
assert result['sha256'] == metadata['input_sha256'] == '1d5251ecf13148683a28bff04b7e3f1c5f71cd8e3094eb083e147cb28c2e2419'
assert result['s'] == 17 and result['count'] == 3876566203
assert result['bytes'] == metadata['input_bytes'] == 8*result['count']
progress = [json.loads(line) for line in (output/'progress.log').read_text().splitlines()]
assert progress[0]['event'] == 'started' and progress[-1]['event'] == 'completed'
assert progress[0]['completed'] == 0 and progress[-1]['completed'] == result['count']
assert all(a['completed'] < b['completed'] and a['seconds'] < b['seconds'] for a,b in zip(progress,progress[1:]))
assert all(row['total'] == result['count'] for row in progress)
text = (output/'resources.log').read_text()
assert re.search(r'Exit status: 0\s*$', text)
peak = int(re.search(r'Maximum resident set size \(kbytes\): (\d+)', text)[1])
assert peak == max(row['peak_rss_kib'] for row in progress) and peak*1024 < 1<<30
histograms = {}
previous = None
for ranks in (64,32,16,8,4,2):
    counts = result['owners'][str(ranks)]
    assert len(counts) == ranks and all(type(v) is int and v > 0 for v in counts)
    assert sum(counts) == result['count']
    if previous is not None:
        assert counts == [previous[r]+previous[r+ranks] for r in range(ranks)]
    previous = counts
    histograms[str(ranks)] = dict(minimum=min(counts), maximum=max(counts),
        maximum_over_average=max(counts)*ranks/result['count'],
        maximum_index_GiB=8*max(counts)/(1<<30))
analysis = dict(status='passed', job_id='82200', controller_seconds=metadata['controller_seconds'],
                scan_seconds=result['seconds'], peak_rss_MiB=peak/1024,
                input_GiB=result['bytes']/(1<<30), histograms=histograms,
                limitation='Saved evidence/source reconciliation;input SHA256 was checked by the job,not reread by this analyzer')
(output/'analysis.json').write_text(json.dumps(analysis,indent=2)+'\n')
print(json.dumps(analysis,indent=2))
