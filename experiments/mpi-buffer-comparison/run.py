#!/usr/bin/env python3
"""Same-build, same-allocation staged/direct crossover with strict tree checks."""
from collections import Counter
from pathlib import Path
import hashlib
import json
import runpy
import statistics
import sys
HERE=Path(__file__).resolve().parent
TREE=runpy.run_path(str(HERE.parent/'root-merge-tree/run.py'))
PROOF=runpy.run_path(str(HERE.parent/'root-transfer-diagnostics/analyze.py'))
ARMS=(('staged-a','staged'),('direct-a','direct'),('direct-b','direct'),('staged-b','staged'))


def review_hosts(rows,mode,small=False,transport=None):
    hosts=[r for r in rows if r.get('kind')=='host']
    expected=(8192 if small else 524288)if mode=='staged'else 0
    assert mode in ('direct','staged')and len(hosts)==4 and {r['rank']for r in hosts}==set(range(4))
    assert all(r['buffer_mode']==mode and r['staging_bytes']==expected for r in hosts)
    if transport is not None:
        assert all(all(r['mpi_transport'].get(key)==value for key,value in transport.items())for r in hosts)
    return {str(r['rank']):r['host']for r in hosts}


def summarize_cases(rows):
    measurements=[r for r in rows if r.get('kind')=='measurement'and r['rank']==0 and r['sample']not in(-1,2)]
    return [dict(case=case,s=group[0]['s'],support_count=group[0]['support_count'],timed_samples=len(group),
        median_seconds={field:statistics.median(r[field]for r in group)for field in TREE['FIELDS']})
        for case in sorted({r['case']for r in measurements})
        for group in [[r for r in measurements if r['case']==case]]]


def review_arm(out,cache,name,mode,transport=None):
    result=dict(name=name,mode=mode,status='passed')
    for label,small in (('normal',False),('memcheck',True)):
        rows=TREE['BASE']['parse'](out/f'{name}-{label}.log')
        directory=cache/(name+('-sanitized'if small else '-normal'))
        TREE['review'](rows,directory,small=small,repeats=12 if not small else 2)
        proofs=PROOF['review_proofs'](rows);hosts=review_hosts(rows,mode,small,transport)
        result[label]=dict(proof_events=len(proofs),proof_phases=dict(Counter(r['phase']for r in proofs)),
            measurement_rows=sum(r.get('kind')=='measurement'for r in rows),hosts=hosts,
            peak_logged_rss_bytes_per_rank=max(r.get('peak_rss_bytes',0)for r in rows),
            peak_owned_device_bytes_per_rank=max(r.get('owned_device_bytes',0)for r in rows))
        if small:
            text=(out/f'{name}-{label}.log').read_text()
            assert text.count('ERROR SUMMARY: 0 errors')==text.count('LEAK SUMMARY: 0 bytes')==4
        else:
            result['cases']=summarize_cases(rows)
            result['final_sha256']={str(case):hashlib.sha256((directory/'rank-0'/f'case-{case}-final.bin').read_bytes()).hexdigest()for case in range(3)}
            result['setup_seconds_by_case']={str(case):max(r['setup_seconds']for r in rows if r.get('kind')=='fixture'and r['case']==case)for case in range(3)}
            for rank in range(4):assert (out/f'{name}-normal-{rank}.pprof').stat().st_size>0
    assert result['normal']['hosts']==result['memcheck']['hosts']
    return result


def compare(arms):
    assert [(r['name'],r['mode'])for r in arms]==list(ARMS)
    assert all(r['status']=='passed'and r['final_sha256']==arms[0]['final_sha256']and
        r['normal']['hosts']==arms[0]['normal']['hosts']for r in arms)
    comparisons=[]
    for case in range(3):
        times=[next(c for c in r['cases']if c['case']==case)['median_seconds']['seconds']for r in arms]
        assert all(t>0 for t in times)
        comparisons.append(dict(case=case,staged_over_direct_pair_a=times[0]/times[1],staged_over_direct_pair_b=times[3]/times[2],
            note='Ratios of arm medians; not independent allocation repeats or an isolated memcpy measurement.'))
    return comparisons


def suite(context):
    env=context['env'];cache=context['cache'];out=context['out']
    env.update(PIR_ROOT_TREE_DIAGNOSTICS='1',PIR_ROOT_TREE_REPEATS='12')
    context['resources'].update(mpi_buffer_comparison_order=[name for name,_ in ARMS],max_staging_bytes_per_rank=524288)
    prepared=TREE['prepare'](context)
    hashes={p.name:hashlib.sha256(p.read_bytes()).hexdigest()for p in prepared[:2]}
    state=dict(status='running',build_sha256=hashes,arms=[])
    def save():(out/'comparison-progress.json').write_text(json.dumps(state,indent=2)+'\n')
    save()
    for name,mode in ARMS:
        state['current_arm']=name;save()
        env['PIR_ROOT_TREE_MPI_BUFFER']=mode
        try:
            TREE['execute'](context,prepared,prefix=name+'-')
            state['arms'].append(review_arm(out,cache,name,mode,context['resources'].get('diagnostic_transport')));save()
        except BaseException as error:
            state.update(status='failed',error=str(error));save();raise
        assert hashes=={p.name:hashlib.sha256(p.read_bytes()).hexdigest()for p in prepared[:2]}
    state.update(status='passed',comparisons=compare(state['arms']));save()
    return dict(buffer_comparison=state)


if __name__=='__main__':
    if '--check'in sys.argv:print('staged/direct/direct/staged; one build; 2 HPC nodes / 2 H200 / 4 logical ranks / 4 CPUs / 8 GiB host; 15 min scheduler; unchanged 128 MiB retained cap; stop on failure')
    else:TREE['BASE']['main'](here=HERE,extra_suite=suite)
