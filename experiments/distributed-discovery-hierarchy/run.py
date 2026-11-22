#!/usr/bin/env python3
"""Genuine partitioned discovery feeding the shared four-worker root hierarchy."""
from pathlib import Path
import runpy
import sys
HERE=Path(__file__).resolve().parent
TREE=runpy.run_path(str(HERE.parent/'distributed-root-hierarchy/run.py'))
DISCOVERY=runpy.run_path(str(HERE.parent/'distributed-discovery-pilot/run.py'))
VECTORS=TREE['VECTORS']


def review(rows,directory):
    distributed=[r for r in rows if r.get('discovery')is True]
    core=[r for r in rows if r.get('discovery')is not True]
    discovered=DISCOVERY['review_discovery'](distributed,directory,[4,6],ranks=4)
    cases=[]
    for s in (4,6):
        label=f's{s}';n=1<<(2*s);q=1<<s
        raw,indices=DISCOVERY['PILOT']['read_words'](directory/f'{label}-native-support.bin',8)
        info=discovered['discovery'][label];k=info[0]['k']
        assert k=={4:97,6:1910}[s]and info[0]['samples']==1
        assert len(indices)==k and indices==sorted(set(indices))and all(i<n-1 for i in indices)
        assert all(i*q%(n-1)in set(indices)for i in indices)
        counts=[]
        for rank in range(4):
            _,owned=DISCOVERY['PILOT']['read_words'](directory/f'{label}-rank-{rank}-support.bin',8)
            assert owned==[i for i in indices if min(i,i*q%(n-1))%4==rank]
            assert len(owned)==info[rank]['root_count'];counts.append(len(owned)+1)
        cases.append((s,k,label,counts))
    result=TREE['review'](core,directory,cases=cases)
    result.update(discovered)
    return result


def suite(c):
    c['run']('discovery-hierarchy-review-tests',['python3','-B','experiments/distributed-discovery-hierarchy/test_review.py'],30)
    c['env'].update(CGO_ENABLED='1',CC='mpicc',CXX='mpicxx')
    c['run']('discovery-shared-review-tests',['python3','-B','experiments/distributed-discovery-pilot/test_review.py'],30)
    c['run']('hierarchy-shared-review-tests',['python3','-B','experiments/distributed-root-hierarchy/test_review.py'],30)
    arms=[dict(name='normal',test='TestGenuineDiscoveryHierarchy',memcheck=False,profile=True,env={},review=review)]
    return VECTORS['suite'](c,full_fft=False,arms=arms,ranks=4,scope='genuine s4/s6 four-worker partitioned discovery feeding concurrent subgroup root hierarchy;not Newton/restart/fulls18')


if __name__=='__main__':
    if '--check'in sys.argv:print('2 HPC nodes/2H200/4CPU/8GiB requested RAM;4workers,64MiBowned/worker=128MiB/card,1GiBRSS/worker,128MiBcache/2GiBscratch/32MiBfile,720s/15min;genuine s4/s6 discovery+hierarchical roots;no CUDA change/default/fullsize/resource increase;no submission')
    else:VECTORS['BASE']['main'](here=HERE,extra_suite=suite)
