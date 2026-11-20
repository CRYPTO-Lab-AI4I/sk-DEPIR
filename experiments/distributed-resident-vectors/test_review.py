#!/usr/bin/env python3
from pathlib import Path
import runpy
import tempfile
import unittest
HERE=Path(__file__).resolve().parent
H=runpy.run_path(str(HERE/'run.py'))
Fixture=runpy.run_path(str(HERE.parent/'distributed-gpu-vectors/test_review.py'))['ReviewTest']

class ReviewTests(unittest.TestCase):
    def test_review_and_preservation(self):
        with tempfile.TemporaryDirectory()as tmp:
            directory=Path(tmp);rows=Fixture().fixture(directory)
            for row in rows:
                if row.get('kind')=='measurement':row.update(resident=True,third_preserved=True,reused=True,range_rejection_preserved=True)
            self.assertEqual(len(H['review'](rows,directory)['measurements']),24)
            row=next(r for r in rows if r.get('kind')=='measurement')
            for key in ('resident','third_preserved','reused','range_rejection_preserved'):
                row[key]=False
                with self.subTest(key=key),self.assertRaises(AssertionError):H['review'](rows,directory)
                row[key]=True
    def test_selects_driver_and_baseline(self):
        old=H['VECTORS']['suite'];calls=[]
        H['VECTORS']['suite']=lambda c,**kwargs:calls.append(kwargs)
        try:
            H['suite'](dict(run=lambda *args:None));self.assertEqual(calls[0]['test_name'],'TestDistributedResidentVectors');self.assertTrue(calls[0]['baseline']);self.assertIs(calls[0]['review_fn'],H['review'])
        finally:H['VECTORS']['suite']=old

if __name__=='__main__':unittest.main()
