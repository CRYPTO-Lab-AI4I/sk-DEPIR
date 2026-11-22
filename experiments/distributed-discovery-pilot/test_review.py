#!/usr/bin/env python3
from pathlib import Path
import copy
import runpy
import unittest
H=runpy.run_path(str(Path(__file__).with_name('run.py')))


class ReviewTest(unittest.TestCase):
    def fixture(self):
        rows=[]
        for rank in (0,1):
            rows.append(dict(kind='discovery-result',rank=rank,s=4,k=97,samples=2,grid_words=128,first=rank*128,local_good=[60,37][rank],raw_count=[50,47][rank],root_count=[49,48][rank],entropy_sha256='ab'*32 if rank==0 else'',owned_device_bytes=1,peak_rss_bytes=1))
            for sample,count in ((1,[40,40]),(2,[50,47])):rows.append(dict(kind='discovery-sample',rank=rank,s=4,sample=sample,k=97,global_count=sum(count),local_count=count[rank],good_counts=[60,37],operation_start=1+(sample-1)*12))
        rows.append(dict(kind='discovery-reference',rank=0,s=4,samples=2,matches=True,entropy_sha256='ab'*32,reference_entropy_sha256='ab'*32))
        return rows

    def test_random_stream_and_global_stop(self):
        rows=self.fixture();self.assertEqual(H['summary'](rows,4)[1:],(97,2))
        bad=copy.deepcopy(rows);bad[-1]['reference_entropy_sha256']='cd'*32
        with self.assertRaises(AssertionError):H['summary'](bad,4)
        bad=copy.deepcopy(rows);next(r for r in bad if r.get('kind')=='discovery-sample')['global_count']=97
        with self.assertRaises(AssertionError):H['summary'](bad,4)

    def test_same_two_integrated_arms(self):
        old=H['VECTORS']['suite'];captured=[];H['VECTORS']['suite']=lambda c,**kw:captured.append(kw)
        try:
            H['suite'](dict(run=lambda *a:None,env={},go='go'))
            self.assertFalse(captured[0]['full_fft'])
            self.assertEqual([(a['name'],a['memcheck'])for a in captured[0]['arms']],[('normal',False),('memcheck',True)])
            self.assertTrue(all(a['env']['PIR_PIPELINE_DISCOVERY']=='distributed'for a in captured[0]['arms']))
        finally:H['VECTORS']['suite']=old


if __name__=='__main__':unittest.main()
