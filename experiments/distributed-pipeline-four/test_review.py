#!/usr/bin/env python3
from pathlib import Path
import json
import runpy
import tempfile
import unittest
HERE=Path(__file__).resolve().parent
H=runpy.run_path(str(HERE/'run.py'))


class ReviewTest(unittest.TestCase):
    def test_four_rank_commit_and_resume_suffix(self):
        fixture=runpy.run_path(str(HERE.parent/'distributed-preprocessing-pilot/test_review.py'))['ReviewTest']
        with tempfile.TemporaryDirectory()as tmp:
            d=Path(tmp);s,k=fixture().fixture(d,ranks=4)
            result=H['PILOT']['artifacts'](d,s,k,ranks=4)
            self.assertEqual(result['checkpoint_prefixes'][-1],248)
            commit=d/'s4-newton-8-commit.json';record=json.loads(commit.read_text());record['ranks']=2;commit.write_text(json.dumps(record))
            with self.assertRaises(AssertionError):H['PILOT']['artifacts'](d,s,k,ranks=4)
            record['ranks']=4;commit.write_text(json.dumps(record))
            data=d/'s4-rank-3-newton-8.bin';raw=bytearray(data.read_bytes());raw[-1]^=1;data.write_bytes(raw)
            with self.assertRaises(AssertionError):H['PILOT']['artifacts'](d,s,k,ranks=4)

    def test_same_two_launches_four_workers(self):
        old=H['RESTART']['VECTORS']['suite'];selected=[];H['RESTART']['VECTORS']['suite']=lambda c,**kw:selected.append(kw)
        try:
            H['RESTART']['suite'](dict(run=lambda *a:None,go='go',env={},cache=Path('/cache'),out=Path('/out')),reference_from=Path('/references'),review_fn=H['review'],ranks=4)
            self.assertEqual(selected[0]['ranks'],4)
            self.assertEqual([a['name']for a in selected[0]['arms']],['cold','resume'])
            self.assertTrue(all(a['env']['PIR_PIPELINE_RANKS']=='4'and a['profile']and not a['memcheck']for a in selected[0]['arms']))
            self.assertEqual(selected[0]['arms'][1]['env']['PIR_PIPELINE_RESUME_FROM'],'/cache/cold')
        finally:H['RESTART']['VECTORS']['suite']=old


if __name__=='__main__':unittest.main()
