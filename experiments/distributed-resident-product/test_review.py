#!/usr/bin/env python3
from pathlib import Path
import runpy
import tempfile
import unittest
HERE=Path(__file__).resolve().parent
H=runpy.run_path(str(HERE/'run.py'))
Fixture=runpy.run_path(str(HERE.parent/'distributed-gpu-product/test_review.py'))['ReviewTest']

class ReviewTests(unittest.TestCase):
    def test_resident_claims_and_reuse(self):
        with tempfile.TemporaryDirectory()as tmp:
            directory=Path(tmp);rows=Fixture().fixture(directory)
            for r in rows:
                if r.get('kind')=='measurement':r.update(resident=True,right_spectrum_preserved=True,reused=True)
                if r.get('kind')=='pointwise-check':r.update(words=min(1<<(r['dimension']-1),r['words']),checked_words=1<<(r['dimension']-1),resident=True,alias_rejected=True,layout_rejected=True,invalid_rejected=True,third_preserved=True)
            self.assertEqual(len(H['review'](rows,directory)['measurements']),24)
            for kind,keys in (('measurement',('resident','right_spectrum_preserved','reused')),('pointwise-check',('resident','alias_rejected','layout_rejected','invalid_rejected','third_preserved'))):
                row=next(r for r in rows if r.get('kind')==kind)
                for key in keys:
                    row[key]=False
                    with self.subTest(key=key),self.assertRaises(AssertionError):H['review'](rows,directory)
                    row[key]=True
    def test_selects_product_and_reference(self):
        old=H['VECTORS']['suite'];calls=[];H['VECTORS']['suite']=lambda c,**kwargs:calls.append(kwargs)
        try:
            H['suite'](dict(run=lambda *args:None));self.assertEqual(calls[0]['test_name'],'TestDistributedResidentProducts');self.assertEqual(calls[0]['reference_test'],'TestDistributedGPUProducts');self.assertIs(calls[0]['reference_review'],H['PRODUCT']['review']);self.assertTrue(calls[0]['full_fft'])
        finally:H['VECTORS']['suite']=old

if __name__=='__main__':unittest.main()
