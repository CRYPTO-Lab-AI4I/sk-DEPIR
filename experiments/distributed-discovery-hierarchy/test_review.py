#!/usr/bin/env python3
from pathlib import Path
import copy
import runpy
import unittest
H=runpy.run_path(str(Path(__file__).with_name('run.py')))


class ReviewTest(unittest.TestCase):
    def test_four_rank_counts_and_random_stream(self):
        rows=[];counts=[25,24,24,24]
        for rank in range(4):
            rows.append(dict(kind='discovery-result',rank=rank,s=4,k=97,samples=2,grid_words=64,first=rank*64,local_good=counts[rank],raw_count=counts[rank],root_count=counts[rank],entropy_sha256='ab'*32 if rank==0 else'',owned_device_bytes=1,peak_rss_bytes=1))
            for sample,found in ((1,[20]*4),(2,counts)):
                rows.append(dict(kind='discovery-sample',rank=rank,s=4,sample=sample,k=97,global_count=sum(found),local_count=found[rank],good_counts=counts,operation_start=1+(sample-1)*12))
        rows.append(dict(kind='discovery-reference',rank=0,s=4,samples=2,matches=True,entropy_sha256='ab'*32,reference_entropy_sha256='ab'*32))
        self.assertEqual(H['DISCOVERY']['summary'](rows,4,ranks=4)[1:],(97,2))
        for field,value in (('global_count',97),('local_count',65)):
            bad=copy.deepcopy(rows);next(r for r in bad if r['kind']=='discovery-sample')[field]=value
            with self.assertRaises(AssertionError):H['DISCOVERY']['summary'](bad,4,ranks=4)
        with self.assertRaises(AssertionError):H['DISCOVERY']['summary'](rows[:-1],4,ranks=4)

    def test_one_composed_arm_same_budget(self):
        old=H['VECTORS']['suite'];selected=[];H['VECTORS']['suite']=lambda c,**kw:selected.append(kw)
        try:
            H['suite'](dict(run=lambda *a:None,env={}))
            self.assertEqual(selected[0]['ranks'],4)
            self.assertFalse(selected[0]['full_fft'])
            self.assertEqual([(a['test'],a['memcheck'],a['profile'])for a in selected[0]['arms']],[('TestGenuineDiscoveryHierarchy',False,True)])
        finally:H['VECTORS']['suite']=old


if __name__=='__main__':unittest.main()
