#!/usr/bin/env python3
from pathlib import Path
import copy
import runpy
import unittest
HERE=Path(__file__).resolve().parent
H=runpy.run_path(str(HERE/'run.py'))


class ReviewTest(unittest.TestCase):
    def test_source_residency_and_single_download_claims(self):
        result=dict(measurements=[dict(spectrum_mode='stream',spectrum_resident_host_words=0,spectrum_downloaded_words=16384,words=16384,spectrum_host_words=4096)])
        self.assertIs(H['source_claims'](result,narrow=True),result)
        for field,value in [('spectrum_mode','resident'),('spectrum_resident_host_words',16384),('spectrum_downloaded_words',32768),('spectrum_host_words',16384)]:
            bad=copy.deepcopy(result);bad['measurements'][0][field]=value
            with self.assertRaises(AssertionError):H['source_claims'](bad,narrow=True)

    def test_shared_suite_receives_source_selection_and_wide_arm(self):
        old=H['SCALE']['suite'];captured=[]
        H['SCALE']['suite']=lambda c,**kw:captured.append((dict(c['env']),kw))or{}
        try:
            H['suite'](dict(run=lambda *a:None,env={},resources={}))
            env,kw=captured[0]
            self.assertEqual(env,dict(PIR_PIPELINE_SPECTRUM='stream'))
            self.assertIs(kw['review_fn'],H['review'])
            self.assertEqual(len(kw['extra_arms']),1)
            self.assertEqual(kw['extra_arms'][0]['test'],'TestWidePipelineTail')
            self.assertIs(kw['extra_arms'][0]['review'],H['wide_review'])
        finally:H['SCALE']['suite']=old


if __name__=='__main__':unittest.main()
