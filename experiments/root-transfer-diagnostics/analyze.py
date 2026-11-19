#!/usr/bin/env python3
"""Audit diagnostic outcomes; a clean retry does not repair unexplained corruption."""
import argparse
import hashlib
import json
from pathlib import Path
import runpy
HERE=Path(__file__).resolve().parent


def compare_snapshots(left,right):
    """Bounded diagnostic files: report the first byte/word mismatch, not a cause."""
    a=Path(left).read_bytes();b=Path(right).read_bytes()
    first=next((j for j,(x,y)in enumerate(zip(a,b))if x!=y),None)
    if first is None and len(a)!=len(b):first=min(len(a),len(b))
    result=dict(equal=a==b,left_bytes=len(a),right_bytes=len(b),first_differing_byte=first)
    if first is not None:
        word=first//4;start=word*4
        result.update(first_differing_word=word,left_word=int.from_bytes(a[start:start+4],'little')if len(a)>=start+4 else None,
            right_word=int.from_bytes(b[start:start+4],'little')if len(b)>=start+4 else None)
    return result


def main():
    p=argparse.ArgumentParser(description=__doc__);p.add_argument('name');a=p.parse_args()
    out=HERE/'runs'/a.name;raw=json.loads((out/'results.json').read_text());tree=runpy.run_path(str(HERE.parent/'root-merge-tree/run.py'))
    reconcile=runpy.run_path(str(HERE.parent/'s17-arithmetic/analyze.py'))['reconcile']
    result=reconcile(out,require_success=raw['status']=='passed',command_fields=('label','seconds','exit_code'))
    rows=tree['BASE']['parse'](out/'normal.log')if(out/'normal.log').exists()else[]
    captures=[r for r in rows if r.get('kind')=='transport-failure']
    inspected=[]
    for r in captures:
        path=Path(raw['cache'])/'normal'/f"rank-{r['rank']}"/r['file']
        data=path.read_bytes();assert len(data)==4*r['words']
        inspected.append(dict(rank=r['rank'],peer=r['peer'],case=r['case'],sample=r['sample'],direction=r['direction'],file=str(path),
            snapshot_sha256=hashlib.sha256(data).hexdigest(),own_proof=r['own_proof'],remote_proof=r['remote_proof']))
    proofs=[r for r in rows if r.get('kind')=='wire-proof']
    for r in proofs:
        assert len(bytes.fromhex(r['own_proof']))==len(bytes.fromhex(r['remote_proof']))==64
    if raw['status']=='passed':
        assert rows==raw['measurements'];assert len(proofs)==sum(r.get('kind')=='transfer'for r in rows) and all(r['matches']for r in proofs);tree['review'](rows,Path(raw['cache'])/'normal',repeats=12)
        small=tree['BASE']['parse'](out/'memcheck.log');assert small==raw['sanitized']
        tree['review'](small,Path(raw['cache'])/'sanitized',small=True)
        text=(out/'memcheck.log').read_text();assert text.count('ERROR SUMMARY: 0 errors')==4 and text.count('LEAK SUMMARY: 0 bytes')==4
        outcome='not reproduced with diagnostic instrumentation; original fault remains unexplained'
    else:
        outcome='captured transfer disagreement'if captures else'failed elsewhere; inspect raw errors'
    pairs=[]
    for sender in inspected:
        if sender['direction']!='send':continue
        peers=[r for r in inspected if r['direction']=='receive'and r['rank']==sender['peer']and r['peer']==sender['rank']and r['case']==sender['case']and r['sample']==sender['sample']]
        pairs.append(dict(sender=sender['rank'],receiver=sender['peer'],paired=len(peers)==1,
            comparison=compare_snapshots(sender['file'],peers[0]['file'])if len(peers)==1 else None))
    result.update(outcome=outcome,captures=inspected,paired_snapshots=pairs,wire_proof_events=len(proofs),tests=['source/evidence/cache seals','diagnostic snapshot extents'],
        scope='diagnostic test, not a production fix or clearance of77704; timings include new fingerprints/control messages',gate_status='original unexplained corruption remains open')
    (out/'analysis.json').write_text(json.dumps(result,indent=2)+'\n');print(json.dumps(result,indent=2))


if __name__=='__main__':main()
