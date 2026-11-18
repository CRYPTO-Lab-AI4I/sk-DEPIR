#!/usr/bin/env python3
import runpy
from pathlib import Path
import unittest
VERIFY=runpy.run_path(str(Path(__file__).with_name('run.py')))['verify_rows']


def fixture():
    rows=[]
    for rank in (0,1):
        rows.extend([dict(rank=rank,host='node'+str(rank)),dict(rank=rank,phase='workflow',event='completed')])
        for size in (1<<20,8<<20,32<<20):
            for done in (2,3,4):
                rows.append(dict(rank=rank,phase='timed-redistribution',buffer_bytes_each=size,
                    completed=done,exchange_seconds=0.01,peak_rss_KiB=65536))
    return rows


class Admission(unittest.TestCase):
    def test_complete(self):VERIFY(fixture())
    def test_remote_payload_is_half_input_not_duplicate_rank_sum(self):
        summary=runpy.run_path(str(Path(__file__).with_name('analyze.py')))['summarize'](fixture())
        largest=summary['samples'][-1]
        self.assertEqual(largest['remote_MiB_per_rank'],16)
        self.assertEqual(len(largest['samples_seconds']),3)
        self.assertEqual(largest['remote_payload_GiBs_per_node'],1.5625)
        self.assertEqual(largest['bidirectional_aggregate_payload_GiBs'],3.125)

    def test_reject_same_node_missing_case_and_rss(self):
        rows=fixture();rows[2]['exchange_seconds']=0
        with self.assertRaises(AssertionError):VERIFY(rows)
        with self.assertRaises(AssertionError):VERIFY(fixture()[:-1])
        rows=fixture();rows[11]['host']='node0'
        with self.assertRaises(AssertionError):VERIFY(rows)
        rows=fixture();rows[2]['peak_rss_KiB']=2<<20
        with self.assertRaises(AssertionError):VERIFY(rows)


if __name__=='__main__':unittest.main()
