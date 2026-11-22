#!/usr/bin/env python3
"""Check explicit s18 coverage and reuse of admitted launch/resource patterns."""
from pathlib import Path
import runpy
import tempfile
import unittest
HERE=Path(__file__).resolve().parent
H=runpy.run_path(str(HERE/'run.py'))


class ReviewTest(unittest.TestCase):
    def test_s18_full_fft_cases_and_missing_packet(self):
        fixture=runpy.run_path(str(HERE.parent/'distributed-gpu-fft/test_review.py'))['ReviewTest']
        for small in (False,True):
            with tempfile.TemporaryDirectory()as tmp:
                d=Path(tmp);cases=[(18,v)for v in ((5,)if small else(5,12))]
                rows=fixture().fixture(d,cases)
                for r in rows:
                    if r.get('kind')=='measurement':r.update(resident=True,third_preserved=True,reused=True,range_rejection_preserved=True,local_taylor=True)
                spec=dict(kind='fft',small=small)
                self.assertEqual(len(H['review'](rows,d,spec)['measurements']),4*len(cases))
                i=next(i for i,r in enumerate(rows)if r.get('kind')=='tile-proof')
                with self.assertRaises(AssertionError):H['review'](rows[:i]+rows[i+1:],d,spec)
                with self.assertRaises(AssertionError):H['review'](rows,d,dict(spec,small=not small))

    def test_only_new_field_and_small_sanitizers(self):
        old=H['VECTORS']['suite'];selected=[];H['VECTORS']['suite']=lambda c,**kw:selected.append(kw)
        try:
            H['suite'](dict(run=lambda *a:None,go='go'))
            arms=selected[0]['arms'];self.assertEqual(len(arms),4)
            self.assertTrue(selected[0]['full_fft'])
            self.assertEqual([a['memcheck']for a in arms],[False,False,True,True])
            self.assertEqual([a['env']['PIR_S18_FIELD_SMALL']for a in arms],['0','0','1','1'])
            self.assertEqual({a['test']for a in arms},{'TestS18GPUFFT','TestS18GPUProducts'})
        finally:H['VECTORS']['suite']=old


if __name__=='__main__':unittest.main()
