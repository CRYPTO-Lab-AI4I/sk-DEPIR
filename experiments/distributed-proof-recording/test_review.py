#!/usr/bin/env python3
from pathlib import Path
import copy
import hashlib
import runpy
import struct
import unittest
HERE=Path(__file__).resolve().parent
H=runpy.run_path(str(HERE/'run.py'))


class ReviewTest(unittest.TestCase):
    def test_independent_go_golden_and_summary_rejection(self):
        row=dict(case='golden',kind='operand-proof',packet=dict(Operation=1,Domain=1<<61,Operand=0,Offset=1<<60,Words=4,Source=0,Target=1),rank=0)
        summary=H['proof_summary']([row]);digest=summary[0,'operand-proof']['sha256']
        go=(HERE.parents[1]/'encoding/liftrs-shared/distributed_proof_log_test.go').read_text()
        self.assertIn(digest,go)
        record=dict(kind='proof-summary',proof_kind='operand-proof',rank=0,complete=True,mode='summary',codec='go-json-v1-length64le',**summary[0,'operand-proof'])
        self.assertEqual(H['check_summaries']([record],[row],'summary')['records'],1)
        for field,value in [('records',2),('sha256','0'*64),('complete',False),('mode','trace')]:
            bad=dict(record);bad[field]=value
            with self.assertRaises(AssertionError):H['check_summaries']([bad],[row],'summary')
        with self.assertRaises(AssertionError):H['check_summaries']([record,row],[row],'summary')

    def test_native_window_count_from_independent_intervals(self):
        rows=[dict(kind='pipeline',case='s4',n=32)]+[dict(kind='taylor-plan',case='s4/spectrum',rank=rank,level=0,degrees=[1],capacity=12)for rank in range(4)]
        expected=H['native_expected'](rows)
        self.assertEqual(expected,{rank:dict(windows=2,coefficients=4)for rank in range(4)})


if __name__=='__main__':unittest.main()
