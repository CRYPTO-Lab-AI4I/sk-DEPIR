#!/usr/bin/env python3
import runpy
from pathlib import Path
import unittest
VERIFY=runpy.run_path(str(Path(__file__).with_name('run.py')))['validate_pair']


def records():
    return [dict(mode=mode,repeat=repeat,rows=[dict(case=case,validation=mode,reference_validation='sorted',sha256=str(case))for case in range(4)])
        for mode,repeat in [('bitmap',0),('sorted',0),('sorted',1),('bitmap',1)]]


class MatchedReview(unittest.TestCase):
    def test_complete(self):VERIFY(records())
    def test_reject_order_or_reference_change_or_digest(self):
        rows=records();rows.reverse()
        with self.assertRaises(AssertionError):VERIFY(rows)
        rows=records();rows[1]['rows'][0]['reference_validation']='bitmap'
        with self.assertRaises(AssertionError):VERIFY(rows)
        rows=records();rows[1]['rows'][0]['sha256']='different'
        with self.assertRaises(AssertionError):VERIFY(rows)


if __name__=='__main__':unittest.main()
