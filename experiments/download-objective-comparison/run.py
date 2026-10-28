#!/usr/bin/env python3
"""Compare saved total-objective and download-objective costs, without searching.

Run after the four cost-grid scripts. Inputs are their linked archives/current
results and option tables. No database encoding or timing is performed.
"""
import hashlib
import json
from pathlib import Path
import subprocess
from datetime import datetime, timezone

ROOT = Path(__file__).resolve().parents[2]
HERE = Path(__file__).resolve().parent

def costs(case):
    if 'result' in case:
        return case['result']['alternatives'][-1]['evaluation']['costs']
    return case['evaluation']['costs']

def frontier(case):
    return {(a['evaluation']['costs']['storage_bits'],
             a['evaluation']['costs']['download_bits'])
            for a in case['result']['alternatives']}

def table_rows(path):
    return [line for line in path.read_text().splitlines()
            if line.startswith('| ')][1:]

def main():
    hashes = {}
    def record(path):
        hashes[str(path.relative_to(ROOT))] = hashlib.sha256(path.read_bytes()).hexdigest()
    summaries = []
    for name in ['greedy-costs', 'application-costs', 'tradeoff-costs', 'dp-costs']:
        directory = ROOT / 'experiments' / name
        before, after = directory/'total-objective-results.json', directory/'results.json'
        record(before); record(after)
        old = json.loads(before.read_text())['scenarios']
        new = json.loads(after.read_text())['scenarios']
        lookup = {(c['application_bytes'],c['record_bytes']):c for c in old}
        scenarios = []
        for case in new:
            previous = lookup[(case['application_bytes'],case['record_bytes'])]
            a,b = costs(previous),costs(case)
            row = {'application_bytes':case['application_bytes'], 'record_bytes':case['record_bytes'],
                   'old_storage_bits':a['storage_bits'], 'new_storage_bits':b['storage_bits'],
                   'old_download_bits':a['download_bits'], 'new_download_bits':b['download_bits'],
                   'endpoint_costs_changed':a['storage_bits']!=b['storage_bits'] or a['download_bits']!=b['download_bits']}
            if 'result' in case:
                fa,fb = frontier(previous),frontier(case)
                row.update(frontier_changed=fa!=fb, points_added=len(fb-fa), points_removed=len(fa-fb))
            scenarios.append(row)
        assert len(scenarios)==21
        summary={'experiment':name, 'scenarios':scenarios,
                 'changed_endpoint_costs':sum(c['endpoint_costs_changed'] for c in scenarios)}
        if 'result' in new[0]:
            before,after=directory/'total-objective-options.md',directory/'options.md'
            record(before);record(after)
            oldrows,newrows=table_rows(before),table_rows(after)
            assert len(oldrows)==len(newrows)==21
            changed=[{'old':a,'new':b} for a,b in zip(oldrows,newrows) if a!=b]
            summary.update(changed_display_rows=len(changed),display_changes=changed,
                           changed_frontiers=sum(c['frontier_changed'] for c in scenarios),
                           points_added=sum(c['points_added'] for c in scenarios),
                           points_removed=sum(c['points_removed'] for c in scenarios))
        summaries.append(summary)
    record(HERE/'run.py')
    result={'timestamp_utc':datetime.now(timezone.utc).isoformat(),
            'base_revision':subprocess.check_output(['git','rev-parse','HEAD'],cwd=ROOT,text=True).strip(),
            'method':'Exact saved-cost comparison; fixed historical catalogue snapshots; no new encoding measurements.',
            'input_sha256':hashes,'comparisons':summaries}
    (HERE/'results.json').write_text(json.dumps(result,indent=2)+'\n')
    lines=['# Download objective comparison','','All four grids contain 21 workloads. Endpoint = minimum-download frontier endpoint for searches, or selected greedy plan.','',
           '| Experiment | Changed endpoint costs | Changed displayed rows | Changed intermediate frontiers |',
           '|---|---:|---:|---:|']
    for x in summaries:
        lines.append(f"| {x['experiment']} | {x['changed_endpoint_costs']}/21 | {x.get('changed_display_rows','—')} | {x.get('changed_frontiers','—')} |")
    lines+=['','Display changes:']
    for x in summaries:
        for change in x.get('display_changes',[]):
            lines+=['',f"{x['experiment']}:", '| Version | Database | Record | Reference N/K · D/B | Lower-storage option | Minimum download option |', '|---|---|---|---:|---:|---:|', '| Before '+change['old'], '| After '+change['new']]
    (HERE/'table.md').write_text('\n'.join(lines)+'\n')
    print('\n'.join(lines))

if __name__=='__main__':
    main()
