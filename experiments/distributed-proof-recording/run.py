#!/usr/bin/env python3
"""Bounded proof recording checked against full same-allocation traces."""
from pathlib import Path
import hashlib
import json
import runpy
import struct
import sys
HERE=Path(__file__).resolve().parent
PREFIX=runpy.run_path(str(HERE.parent/'distributed-prefix-inputs/run.py'))
SCALE=PREFIX['SCALE'];VECTORS=SCALE['VECTORS']
KINDS={'tile-proof','taylor-proof','operand-proof','discovery-wire','support-proof'}
VOLATILE={'utc','rank','elapsed_seconds','peak_rss_bytes'}


def semantic(row):
    # Top-level Go maps sort keys; decoded nested packet structs preserve their
    # encoded field order. Preserve arbitrary-size integers and Go HTML escapes.
    body={k:row[k]for k in sorted(set(row)-VOLATILE)}
    data=json.dumps(body,separators=(',',':'),ensure_ascii=False).encode()
    for source,target in ((b'<',b'\\u003c'),(b'>',b'\\u003e'),(b'&',b'\\u0026'),('\u2028'.encode(),b'\\u2028'),('\u2029'.encode(),b'\\u2029')):data=data.replace(source,target)
    return data


def proof_summary(rows):
    counters={}
    for row in rows:
        if row.get('kind')not in KINDS:continue
        key=row['rank'],row['kind'];item=counters.setdefault(key,dict(records=0,semantic_bytes=0,digest=hashlib.sha256()))
        data=semantic(row);item['digest'].update(struct.pack('<Q',len(data)));item['digest'].update(data)
        item['records']+=1;item['semantic_bytes']+=len(data)
    return{key:dict(records=v['records'],semantic_bytes=v['semantic_bytes'],sha256=v['digest'].hexdigest())for key,v in counters.items()}


def check_summaries(rows,reference,mode):
    expected=proof_summary(reference);actual={}
    for row in rows:
        if row.get('kind')!='proof-summary':continue
        key=row['rank'],row['proof_kind'];assert key not in actual
        assert row['complete']is True and row['mode']==mode and row['codec']=='go-json-v1-length64le'
        actual[key]={k:row[k]for k in ('records','semantic_bytes','sha256')}
    assert actual==expected and expected
    proofs=[r for r in rows if r.get('kind')in KINDS]
    assert not proofs if mode=='summary'else len(proofs)==sum(r['records']for r in expected.values())
    return dict(records=sum(v['records']for v in expected.values()),kinds=len(expected),stored_proof_rows=len(proofs),summaries=len(actual))


def native_expected(rows):
    stages={(r['tree_case'],r['tree_level'],r['group_first']):r['stage']['Domain']for r in rows if r.get('kind')=='root-stage'and r['event']=='started'}
    domains={r['case']:r['n']for r in rows if r.get('kind')in ('pipeline','wide-tail')}
    totals={rank:dict(windows=0,coefficients=0)for rank in range(4)}
    for row in rows:
        if row.get('kind')!='taylor-plan'or not row['degrees']:continue
        if 'tree_case'in row:
            n=stages[row['tree_case'],row['tree_level'],row['group_first']];rank=row['group_rank'];ranks=row['group_size']
        else:n=domains[row['case'].split('/')[0]];rank=row['rank'];ranks=4
        cols=1<<((n.bit_length()-1)//2);half=cols<<row['level'];span=row['capacity']//(len(row['degrees'])+1)
        first,end=rank*n//ranks,(rank+1)*n//ranks
        for block in range(0,n,2*half):
            for quotient in (False,True):
                base=block+(half if quotient else 0);begin,stop=(0,max(row['degrees']))if quotient else(min(row['degrees']),half)
                count=max(0,min(base+stop,end)-max(base+begin,first))
                totals[row['rank']]['windows']+=(count+span-1)//span
                totals[row['rank']]['coefficients']+=count
    return totals


def native_review(out,mode,rows):
    expected=native_expected(rows);actual={rank:dict(windows=0,coefficients=0)for rank in range(4)}
    names={'rank-Taylor-windows-summary':'windows','rank-Taylor-coefficients-summary':'coefficients'}
    for rank in range(4):
        for line in (out/f'{mode}-{rank}.stderr.log').read_text().splitlines():
            try:row=json.loads(line)
            except json.JSONDecodeError:continue
            if row.get('work')!='gpu-preparation':continue
            assert row.get('event')!='failed'
            if row.get('phase')in names:
                assert row['event']=='completed'and row['completed']==row['total']
                actual[rank][names[row['phase']]]+=row['completed']
            assert row.get('phase')not in ('stage-Taylor-quotient','stage-Taylor-remainder')
    assert actual==expected and any(x['windows']for x in actual.values())
    return actual


def review(rows,directory,resume=False,source=None,cold=None):
    result=PREFIX['review'](rows,directory,resume,source,cold)
    result['proof_recording']=check_summaries(rows,rows,'trace')
    return result


def summary_review(rows,directory,reference,resume=False,source=None,cold=None,wide=False):
    recording=check_summaries(rows,reference,'summary')
    expanded=[r for r in rows if r.get('kind')not in KINDS]+[r for r in reference if r.get('kind')in KINDS]
    result=PREFIX['wide_review'](expanded,directory)if wide else PREFIX['review'](expanded,directory,resume,source,cold)
    result['proof_recording']=recording
    return result


def trace_wide_review(rows,directory):
    result=PREFIX['wide_review'](rows,directory)
    result['proof_recording']=check_summaries(rows,rows,'trace')
    return result


def suite(c):
    c['run']('proof-recording-review-tests',['python3','-B','experiments/distributed-proof-recording/test_review.py'],30)
    c['env'].update(PIR_PIPELINE_PROOF_LOG='trace',LIFTRS_GPU_TAYLOR_PROGRESS='bounded',PIR_PIPELINE_NEWTON_INPUT='prefix',PIR_PIPELINE_DISCOVERY_RESULT='stream',PIR_PIPELINE_SPECTRUM='stream',CGO_ENABLED='1',CC='mpicc',CXX='mpicxx')
    for label,race in (('proof-log-cpu',[]),('proof-log-race',['-race'])):
        c['run'](label,[c['go'],'test',*race,'./encoding/liftrs-shared','-run','^TestProofLog|^TestRankStream','-count=1','-timeout=20s'],60)
    def baseline(name):return VECTORS['BASE']['parse_rank_logs'](c['out']/(name+'.log'),ranks=4)
    def summary(rows,directory,reference,resume=False):
        cold=json.loads((c['out']/'composition.json').read_text())['arms']['cold']if resume else None
        return summary_review(rows,directory,baseline(reference),resume,c['cache']/'cold'if resume else None,cold,reference=='wide')
    extra=[dict(name='summary-cold',test='TestDistributedPreprocessingPilot',memcheck=False,profile=True,env=dict(PIR_PIPELINE_RESUME_FROM='',PIR_PIPELINE_PROOF_LOG='summary'),review=lambda rows,directory:summary(rows,directory,'cold')),
        dict(name='summary-resume',test='TestDistributedPreprocessingPilot',memcheck=False,profile=True,env=dict(PIR_PIPELINE_RESUME_FROM=str(c['cache']/'cold'),PIR_PIPELINE_PROOF_LOG='summary'),review=lambda rows,directory:summary(rows,directory,'resume',True)),
        dict(name='wide',test='TestWidePipelineTail',memcheck=False,profile=True,env=dict(PIR_PIPELINE_RESUME_FROM='',PIR_PIPELINE_PROOF_LOG='trace'),review=trace_wide_review),
        dict(name='summary-wide',test='TestWidePipelineTail',memcheck=False,profile=True,env=dict(PIR_PIPELINE_RESUME_FROM='',PIR_PIPELINE_PROOF_LOG='summary'),review=lambda rows,directory:summary(rows,directory,'wide'))]
    result=SCALE['suite'](c,review_fn=review,extra_arms=extra)
    result['native_taylor_totals']={mode:native_review(c['out'],mode,baseline(mode))for mode in ('cold','resume','summary-cold','summary-resume','wide','summary-wide')}
    c['resources'].update(recording='Optional bounded proof counts/digests and native window summaries;every runtime check/RSS guard retained',scope='Same-allocation full-trace/summary genuine s8 cold/restart and bounded field18 pairs;not fulls18')
    return result


if __name__=='__main__':
    if '--check'in sys.argv:print('same2HPCnodes/2H200/4CPU/8GiBhost;4workers,64MiBowned/worker=128MiB/card,1GiBRSS/worker,128MiBcache/2GiBscratch/32MiBfile,720s/15min;6short trace/summary arms;runtime checks unchanged;no submission')
    else:VECTORS['BASE']['main'](here=HERE,extra_suite=suite)
