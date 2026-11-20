#!/usr/bin/env python3
"""Audit diagnostic outcomes; a clean retry does not repair unexplained corruption."""
import argparse
import hashlib
import json
from pathlib import Path
import runpy
import statistics
from collections import Counter
HERE=Path(__file__).resolve().parent


def compare_snapshots(left,right,word_bytes=4):
    """Bounded diagnostic files: report the first byte/word mismatch, not a cause."""
    assert word_bytes in (4,8)
    a=Path(left).read_bytes();b=Path(right).read_bytes()
    first=next((j for j,(x,y)in enumerate(zip(a,b))if x!=y),None)
    if first is None and len(a)!=len(b):first=min(len(a),len(b))
    result=dict(equal=a==b,left_bytes=len(a),right_bytes=len(b),first_differing_byte=first)
    if first is not None:
        word=first//word_bytes;start=word*word_bytes
        result.update(first_differing_word=word,left_word=int.from_bytes(a[start:start+word_bytes],'little')if len(a)>=start+word_bytes else None,
            right_word=int.from_bytes(b[start:start+word_bytes],'little')if len(b)>=start+word_bytes else None)
    return result


def root_mismatch(row):
    own=bytes.fromhex(row['own_proof']);remote=bytes.fromhex(row['remote_proof'])
    assert len(own)==len(remote)==64
    mismatch=own[:32]!=own[32:]or remote[:32]!=remote[32:]or own[:32]!=remote[:32]
    assert row['matches']==(not mismatch)
    return mismatch


def review_proofs(rows):
    """Every received support shard and root transfer must have paired proofs."""
    proofs=[r for r in rows if r.get('kind')=='wire-proof']
    for row in proofs:
        assert not root_mismatch(row)
        opposite='receive'if row['direction']=='send'else'send'
        peers=[r for r in proofs if r['rank']==row['peer']and r['peer']==row['rank']and
            r['direction']==opposite and r['case']==row['case']and r['sample']==row['sample']and
            r.get('phase','root')==row.get('phase','root')and
            r.get('sequence',r.get('stride',0))==row.get('sequence',row.get('stride',0))]
        assert len(peers)==1
        peer=peers[0]
        assert peer['own_proof']==row['remote_proof']and peer['remote_proof']==row['own_proof']and peer['words']==row['words']
    assert sum(r.get('phase','root')=='root'for r in proofs)==sum(r.get('kind')=='transfer'for r in rows)
    # Legacy runs had only root proofs. New runs require full support coverage.
    if any(r.get('phase')=='support'for r in proofs):
        for row in rows:
            if row.get('kind')!='measurement'or row['rank']==0:continue
            received=[r for r in proofs if r.get('phase')=='support'and r['direction']=='receive'and
                r['rank']==row['rank']and r['case']==row['case']and r['sample']==row['sample']]
            assert sum(r['words']for r in received)==row['shard_count']
            assert sorted(r['sequence']for r in received)==list(range(len(received)))
    return proofs


def completed_stage(raw,label):
    """A later reporting failure does not erase an explicitly completed stage."""
    commands=[c for c in raw['commands']if c['label']==label]
    assert len(commands)<=1
    return bool(commands and commands[0]['exit_code']==0 and not commands[0].get('guard'))


def main():
    p=argparse.ArgumentParser(description=__doc__);p.add_argument('name');p.add_argument('--experiment',type=Path,default=HERE);a=p.parse_args()
    out=a.experiment/'runs'/a.name;raw=json.loads((out/'results.json').read_text());tree=runpy.run_path(str(HERE.parent/'root-merge-tree/run.py'))
    reconcile=runpy.run_path(str(HERE.parent/'s17-arithmetic/analyze.py'))['reconcile']
    result=reconcile(out,require_success=raw['status']=='passed',command_fields=('label','seconds','exit_code'))
    rows=tree['BASE']['parse_rank_logs'](out/'normal.log')if(out/'normal.log').exists()else[]
    captures=[r for r in rows if r.get('kind')=='transport-failure']
    inspected=[]
    for r in captures:
        path=Path(raw['cache'])/'normal'/f"rank-{r['rank']}"/r['file']
        data=path.read_bytes();assert len(data)==r.get('word_bytes',4)*r['words']
        inspected.append(dict(rank=r['rank'],peer=r['peer'],case=r['case'],sample=r['sample'],direction=r['direction'],file=str(path),
            word_bytes=r.get('word_bytes',4),phase=r.get('phase','root'),sequence=r.get('sequence',r.get('stride',0)),snapshot_sha256=hashlib.sha256(data).hexdigest(),own_proof=r['own_proof'],remote_proof=r['remote_proof']))
    proofs=[r for r in rows if r.get('kind')=='wire-proof']
    for r in proofs:
        assert len(bytes.fromhex(r['own_proof']))==len(bytes.fromhex(r['remote_proof']))==64
    normal_ok=completed_stage(raw,'normal');memcheck_ok=completed_stage(raw,'memcheck')
    if raw['status']=='passed':assert normal_ok and memcheck_ok
    tests=['source/evidence/cache seals','diagnostic snapshot extents']
    if normal_ok:
        if 'measurements'in raw:assert rows==raw['measurements']
        review_proofs(rows);tree['review'](rows,Path(raw['cache'])/'normal',repeats=12)
        tests+=['normal CPU equality, ownership, snapshots and paired proofs']
    small=[]
    if memcheck_ok:
        assert normal_ok
        small=tree['BASE']['parse_rank_logs'](out/'memcheck.log')
        if 'sanitized'in raw:assert small==raw['sanitized']
        tree['review'](small,Path(raw['cache'])/'sanitized',small=True);review_proofs(small)
        text=tree['BASE']['rank_output_text'](out/'memcheck.log');assert text.count('ERROR SUMMARY: 0 errors')==4 and text.count('LEAK SUMMARY: 0 bytes')==4
        tests+=['four-rank memcheck, zero leaks/errors, snapshots and paired proofs']
    if normal_ok and memcheck_ok:
        outcome='bounded diagnostic validation passed; original fault not reproduced and remains unexplained'
    else:
        outcome='captured transfer disagreement'if captures else'failed elsewhere; inspect raw errors'
    timings=[]
    if normal_ok:
        for case in sorted({r['case']for r in rows if r.get('kind')=='measurement'}):
            samples=[r for r in rows if r.get('kind')=='measurement'and r['case']==case and r['rank']==0 and r['sample']not in(-1,2)]
            timings.append(dict(case=case,s=samples[0]['s'],support_count=samples[0]['support_count'],samples=len(samples),
                median_seconds={key:statistics.median(r[key]for r in samples)for key in tree['FIELDS']}))
    for phase,data,expected in [('normal',rows,524288),('memcheck',small,8192)]:
        if raw.get('resources',{}).get('mpi_buffer_mode')=='staged'and data and (normal_ok if phase=='normal'else memcheck_ok):
            hosts=[r for r in data if r.get('kind')=='host']
            assert len(hosts)==4 and all(r['buffer_mode']=='staged'and r['staging_bytes']==expected for r in hosts)
    peak_rss=max((r.get('peak_rss_bytes',0)for r in rows),default=0)
    peak_device=max((r.get('owned_device_bytes',0)for r in rows),default=0)
    pairs=[]
    for sender in inspected:
        if sender['direction']!='send':continue
        peers=[r for r in inspected if r['direction']=='receive'and r['rank']==sender['peer']and r['peer']==sender['rank']and r['case']==sender['case']and r['sample']==sender['sample']and r['phase']==sender['phase']and r['sequence']==sender['sequence']]
        pairs.append(dict(sender=sender['rank'],receiver=sender['peer'],paired=len(peers)==1,
            comparison=compare_snapshots(sender['file'],peers[0]['file'],sender['word_bytes'])if len(peers)==1 else None))
    result.update(timed_case_medians=timings,normal_proof_phases=dict(Counter(r.get('phase','root')for r in proofs)),normal_peak_logged_rss_bytes_per_rank=peak_rss,normal_peak_owned_device_bytes_per_rank=peak_device,outcome=outcome,captures=inspected,paired_snapshots=pairs,wire_proof_events=len(proofs),tests=tests,
        scope='diagnostic test, not a production fix or clearance of77704; timings include new fingerprints/control messages',gate_status='original unexplained corruption remains open',normal_validation='passed'if normal_ok else'unvalidated',memcheck_validation='passed'if memcheck_ok else'unvalidated',memcheck_proof_events=sum(r.get('kind')=='wire-proof'for r in small),workflow_error=raw.get('error'),successful_measurement_rows=sum(r.get('kind')=='measurement'for r in rows),all_logged_proofs_match=all(not root_mismatch(r)for r in proofs))
    (out/'analysis.json').write_text(json.dumps(result,indent=2)+'\n');print(json.dumps(result,indent=2))


if __name__=='__main__':main()
