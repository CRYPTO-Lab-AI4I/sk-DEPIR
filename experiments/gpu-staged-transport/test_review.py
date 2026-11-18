#!/usr/bin/env python3
from pathlib import Path
import runpy
import tempfile
import unittest
from unittest import mock
HELP=runpy.run_path(str(Path(__file__).with_name('run.py')))


def fixture():
    rows=[]
    for rank in (0,1):
        rows.extend([dict(kind='host',rank=rank,host='node'+str(rank)),dict(rank=rank,phase='workflow',event='completed')])
        for size in HELP['SIZES']:
            for sample in (-1,0,1,2):
                row=dict(kind='measurement',rank=rank,bytes=size,sample=sample,matches=True,owned_device_bytes=64<<20,pinned_host_bytes=2*size,peak_rss_bytes=512<<20)
                row.update({name:0.01 for name in ('host_seconds','producer_seconds','d2h_seconds','exchange_seconds','h2d_seconds','consumer_seconds','pipeline_seconds')})
                rows.append(row)
    return rows


class Review(unittest.TestCase):
    def test_complete_and_remote_byte_count(self):
        rows=fixture();HELP['verify'](rows)
        summary=runpy.run_path(str(Path(__file__).with_name('analyze.py')))['summarize'](rows)
        self.assertEqual(summary[-1]['remote_MiB_per_node'],16)
        self.assertEqual(summary[-1]['remote_payload_GiBs_per_node'],1.5625)
    def test_reject_missing_rank_coverage_bad_caps_or_timing(self):
        rows=fixture();rows.pop()
        with self.assertRaises(AssertionError):HELP['verify'](rows)
        rows=fixture();rows[2]['owned_device_bytes']=129<<20
        with self.assertRaises(AssertionError):HELP['verify'](rows)
        rows=fixture();rows[2]['exchange_seconds']=0.02
        with self.assertRaises(AssertionError):HELP['verify'](rows)
        rows=fixture();rows[14]['host']='node0'
        with self.assertRaises(AssertionError):HELP['verify'](rows)
    def test_snapshot_is_git_blob_not_ignored_worktree(self):
        names='code.go\ndocs/no.md\nexperiments/x/runs/evidence.json\n'
        blob=b'package fixture\n';response=b'a'*40+b' blob '+str(len(blob)).encode()+b'\n'+blob+b'\n'
        with tempfile.TemporaryDirectory()as temporary, mock.patch('subprocess.check_output',side_effect=[names,response]):
            root=Path(temporary);hashes=HELP['snapshot']('revision',root)
            self.assertEqual(set(hashes),{'code.go'})
            self.assertEqual((root/'code.go').read_bytes(),blob)
            self.assertFalse((root/'docs/no.md').exists())


if __name__=='__main__':unittest.main()
