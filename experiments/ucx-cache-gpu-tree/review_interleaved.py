#!/usr/bin/env python3
"""Review78249 with an explicit, byte-preserving raw log and derived repair view."""
from pathlib import Path
import hashlib
import json
import re
import runpy
HERE=Path(__file__).resolve().parent
HELP=runpy.run_path(str(HERE.parent/'mpi-buffer-comparison/run.py'))


def recover(text):
    """Only one known OpenMPI warning between adjacent same-rank JSON fragments."""
    lines=text.splitlines();repairs=[]
    for j,line in enumerate(lines):
        match=re.fullmatch(r'(\d+: )(\{.*)(\[[^\]]+:\d+\] \d+ more process has sent help message help-mpi-btl-openib\.txt / no device params found)',line)
        if not match:continue
        prefix,head,warning=match.groups()
        assert j+1<len(lines)and lines[j+1].startswith(prefix),'ambiguous rank continuation'
        tail=lines[j+1][len(prefix):];event=json.loads(head+tail)
        assert event['phase']=='local-vector-transpose'and event['event']=='completed'
        assert event['completed']==event['total']==65536 and event['unit']=='elements'
        repairs.append(dict(first_line=j+1,second_line=j+2,head=head,tail=tail,warning=warning,event=event))
        lines[j]=prefix+head+tail;lines[j+1]=prefix+warning
    assert len(repairs)==1,'expected exactly the one observed interleave'
    return '\n'.join(lines)+'\n',repairs


def main():
    out=HERE/'runs/tree-01';raw=json.loads((out/'results.json').read_text());assert raw['job_id']=='78249'
    reconcile=runpy.run_path(str(HERE.parent/'s17-arithmetic/analyze.py'))['reconcile']
    result=reconcile(out,require_success=False,command_fields=('label','seconds','exit_code'))
    assert raw['status']=='failed'and raw['error'].startswith('Unterminated string')
    state=json.loads((out/'comparison-progress.json').read_text());cache=Path(raw['cache'])
    arms=[HELP['review_arm'](out,cache,r['name'],r['mode'],raw['resources']['diagnostic_transport'])for r in state['arms']]
    assert arms==state['arms']
    original=out/'direct-b-normal.log';derived=out/'direct-b-normal.recovered.log'
    fixed,repairs=recover(original.read_text());derived.write_text(fixed)
    rows=HELP['TREE']['BASE']['parse'](derived)
    HELP['TREE']['review'](rows,cache/'direct-b-normal',repeats=12);HELP['PROOF']['review_proofs'](rows)
    HELP['review_hosts'](rows,'direct',transport=raw['resources']['diagnostic_transport'])
    assert next(c for c in raw['commands']if c['label']=='direct-b-normal')['exit_code']==0
    for case in range(3):assert hashlib.sha256((cache/'direct-b-normal/rank-0'/f'case-{case}-final.bin').read_bytes()).hexdigest()==arms[0]['final_sha256'][str(case)]
    result.update(scope='Partial GPU comparison; explicit derived log recovery, raw producer failure preserved.',completed_arms=arms,
        direct_b_normal=dict(status='passed after explicit log recovery',cases=HELP['summarize_cases'](rows),measurement_rows=sum(r.get('kind')=='measurement'for r in rows),proof_events=len(HELP['PROOF']['review_proofs'](rows)),peak_rss_bytes_per_rank=max(r.get('peak_rss_bytes',0)for r in rows),memcheck='not run',profile_summary='not run'),
        recovery=dict(original_file=original.name,original_sha256=hashlib.sha256(original.read_bytes()).hexdigest(),derived_file=derived.name,derived_sha256=hashlib.sha256(derived.read_bytes()).hexdigest(),repairs=repairs),
        comparisons=[],staged_b='not run',workflow_error=raw['error'])
    (out/'analysis.json').write_text(json.dumps(result,indent=2)+'\n');print(json.dumps(result,indent=2))


if __name__=='__main__':main()
