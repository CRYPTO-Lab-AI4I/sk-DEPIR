#!/usr/bin/env python3
"""Recheck all successful comparison arms without changing producer outcomes."""
from pathlib import Path
import argparse
import hashlib
import json
import runpy
HERE=Path(__file__).resolve().parent
HELP=runpy.run_path(str(HERE/'run.py'))


def review_captures(rows,cache,arm):
    """Bind both durable endpoint payloads to their phase/sequence and proofs."""
    captures=[]
    for row in rows:
        if row.get('kind')!='transport-failure':continue
        assert HELP['PROOF']['root_mismatch'](row)
        path=cache/(arm+'-normal')/f"rank-{row['rank']}"/row['file']
        data=path.read_bytes();assert len(data)==row['word_bytes']*row['words']
        digest=hashlib.sha256(data).hexdigest()
        captures.append(dict(row,file=str(path),snapshot_sha256=digest,capture_matches_logged_sha=digest==row['sha256']))
    pairs=[]
    for sender in captures:
        if sender['direction']!='send':continue
        peers=[r for r in captures if r['direction']=='receive'and r['rank']==sender['peer']and r['peer']==sender['rank']and
            all(r[key]==sender[key]for key in ('case','sample','phase','sequence','word_bytes','words'))]
        assert len(peers)==1
        receiver=peers[0]
        assert sender['own_proof']==receiver['remote_proof']and sender['remote_proof']==receiver['own_proof']
        comparison=HELP['PROOF']['compare_snapshots'](sender['file'],receiver['file'],sender['word_bytes'])
        a=Path(sender['file']).read_bytes();b=Path(receiver['file']).read_bytes();width=sender['word_bytes']
        changes=[j for j in range(0,len(a),width)if a[j:j+width]!=b[j:j+width]]
        pairs.append(dict(sender=sender['rank'],receiver=receiver['rank'],same_host=sender['host']==receiver['host'],
            sender_host=sender['host'],receiver_host=receiver['host'],comparison=comparison,changed_words=len(changes),
            changed_words_received_as_zero=sum(not any(b[j:j+width])for j in changes)))
    assert len(captures)==2*len(pairs)
    return dict(captures=captures,paired_snapshots=pairs)


def main():
    parser=argparse.ArgumentParser(description=__doc__);parser.add_argument('name');parser.add_argument('--experiment',type=Path,default=HERE);args=parser.parse_args()
    out=args.experiment/'runs'/args.name;raw=json.loads((out/'results.json').read_text())
    reconcile=runpy.run_path(str(HERE.parent/'s17-arithmetic/analyze.py'))['reconcile']
    result=reconcile(out,require_success=raw['status']=='passed',command_fields=('label','seconds','exit_code'))
    cache=Path(raw['cache']);progress=out/'comparison-progress.json'
    state=json.loads(progress.read_text())if progress.exists()else dict(status='preparation failed',arms=[],build_sha256={})
    for name,digest in state['build_sha256'].items():assert hashlib.sha256((cache/name).read_bytes()).hexdigest()==digest
    arms=[]
    for recorded in state['arms']:
        arm=HELP['review_arm'](out,cache,recorded['name'],recorded['mode'],raw['resources'].get('diagnostic_transport'));assert arm==recorded;arms.append(arm)
    if raw['status']=='passed':
        assert state['status']=='passed'and state==raw['buffer_comparison']
        ratios=HELP['compare'](arms);assert ratios==state['comparisons']
    else:ratios=[]
    failure={}
    current=state.get('current_arm')
    if current and (out/f'{current}-normal.log').exists():
        failure=review_captures(HELP['TREE']['BASE']['parse_rank_logs'](out/f'{current}-normal.log'),cache,current)
    worker_settings={}
    native_selection={}
    transport=raw['resources'].get('diagnostic_transport')
    if transport:
        for name,mode in HELP['ARMS']:
            for phase in ('normal','memcheck'):
                log=out/f'{name}-{phase}.log'
                if not log.exists():continue
                hosts=[r for r in HELP['TREE']['BASE']['parse_rank_logs'](log)if r.get('kind')=='host']
                assert all(all(r['mpi_transport'].get(key)==value for key,value in transport.items())for r in hosts)
                worker_settings[name+'-'+phase]=len(hosts)
                if transport.get('OMPI_MCA_pml')=='ucx'and transport.get('OMPI_MCA_pml_base_verbose')=='20':
                    checked=0
                    for rank in range(4):
                        stderr=out/f'{name}-{phase}-{rank}.stderr.log'
                        if not stderr.exists():continue  # Historical merged logs remain separate evidence.
                        assert any(line.rstrip().endswith('select: component ucx selected')for line in stderr.read_text().splitlines())
                        checked+=1
                    native_selection[name+'-'+phase]=checked
    result.update(native_ucx_selection_checked=native_selection,worker_settings_checked=worker_settings,failure_evidence=failure,scope='Bounded same-build diagnostic crossover; original direct-buffer fault remains unexplained.',
        completed_arms=arms,comparisons=ratios,workflow_error=raw.get('error'),current_arm=state.get('current_arm'))
    (out/'analysis.json').write_text(json.dumps(result,indent=2)+'\n');print(json.dumps(result,indent=2))


if __name__=='__main__':main()
