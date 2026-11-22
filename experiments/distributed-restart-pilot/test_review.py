#!/usr/bin/env python3
from pathlib import Path
import runpy
import tempfile
import unittest
HERE=Path(__file__).resolve().parent
H=runpy.run_path(str(HERE/'run.py'))


class ReviewTest(unittest.TestCase):
    def test_resume_needs_only_completed_prefix_and_suffix(self):
        fixture=runpy.run_path(str(HERE.parent/'distributed-preprocessing-pilot/test_review.py'))['ReviewTest']
        with tempfile.TemporaryDirectory()as tmp:
            d=Path(tmp);s,k=fixture().fixture(d)
            for size in (2,4):
                (d/f's4-newton-{size}-commit.json').unlink()
                for rank in (0,1):(d/f's4-rank-{rank}-newton-{size}.bin').unlink()
            self.assertEqual(H['PILOT']['artifacts'](d,s,k,8)['checkpoint_prefixes'][0],8)
            with self.assertRaises(FileNotFoundError):H['PILOT']['artifacts'](d,s,k)

    def test_two_fresh_launches_and_source(self):
        old=H['VECTORS']['suite'];selected=[];H['VECTORS']['suite']=lambda c,**kw:selected.append(kw)
        try:
            H['suite'](dict(run=lambda *a:None,go='go',env={},cache=Path('/cache'),out=Path('/out')))
            arms=selected[0]['arms'];self.assertEqual([r['name']for r in arms],['cold','resume']);self.assertTrue(all(r['memcheck']is False for r in arms))
            self.assertEqual(arms[0]['env']['PIR_PIPELINE_RESUME_FROM'],'');self.assertEqual(arms[1]['env']['PIR_PIPELINE_RESUME_FROM'],'/cache/cold')
            self.assertTrue(all(r['env']['PIR_PIPELINE_PUBLICATION']=='stream'for r in arms))
        finally:H['VECTORS']['suite']=old


if __name__=='__main__':unittest.main()
