#!/usr/bin/env python3
from pathlib import Path
import copy
import runpy
import unittest
HERE=Path(__file__).resolve().parent
H=runpy.run_path(str(HERE/'run.py'))


class ReviewTest(unittest.TestCase):
    def fixture(self,mode):
        result=dict(measurements=[dict(discovery_result_mode=mode)for _ in range(4)])
        common=dict(grid_words=16384,result_mode=mode,result_buffer_words=4096 if mode=='stream'else 16384,result_resident_host_words=0 if mode=='stream'else 16384)
        rows=[dict(common,kind='discovery-result',samples=2,result_downloaded_words=32768)for _ in range(4)]
        rows.extend(dict(common,kind='discovery-sample',result_downloaded_words=16384)for _ in range(8))
        return rows,result

    def test_exact_download_and_buffer_bounds(self):
        for mode in ('stream','resident'):
            rows,result=self.fixture(mode)
            self.assertIs(H['discovery_claims'](rows,result,mode),result)
        rows,result=self.fixture('stream')
        for field,value in [('result_buffer_words',16384),('result_resident_host_words',16384),('result_downloaded_words',65536),('result_mode','resident')]:
            bad=copy.deepcopy(rows);bad[0][field]=value
            with self.assertRaises(AssertionError):H['discovery_claims'](bad,result,'stream')
        with self.assertRaises(AssertionError):H['discovery_claims'](rows,result,'stream',True)
        self.assertIs(H['discovery_claims']([],result,'stream',True),result)

    def test_matched_arm_holds_spectrum_mode_constant(self):
        old=H['SCALE']['suite'];captured=[]
        H['SCALE']['suite']=lambda c,**kw:captured.append((dict(c['env']),kw))or{}
        try:
            H['suite'](dict(run=lambda *a:None,env={},resources={},go='/go'))
            env,kw=captured[0];self.assertEqual(env['PIR_PIPELINE_DISCOVERY_RESULT'],'stream');self.assertEqual(env['PIR_PIPELINE_SPECTRUM'],'stream')
            self.assertIs(kw['review_fn'],H['review'])
            arm=kw['extra_arms'][0];self.assertEqual(arm['name'],'resident-discovery')
            self.assertEqual(arm['env']['PIR_PIPELINE_DISCOVERY_RESULT'],'resident');self.assertEqual(arm['env']['PIR_PIPELINE_SPECTRUM'],'stream')
            self.assertIs(arm['review'],H['resident_review'])
        finally:H['SCALE']['suite']=old


if __name__=='__main__':unittest.main()
