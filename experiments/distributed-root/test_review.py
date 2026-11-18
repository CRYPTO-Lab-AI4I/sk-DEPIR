#!/usr/bin/env python3
"""Small reviewer regressions; no MPI, GPU or scheduler needed."""
import copy
import hashlib
from pathlib import Path
import runpy
import struct
import tempfile
import unittest
HELP=runpy.run_path(str(Path(__file__).with_name('run.py')))


def fixture(directory):
    rows=[]
    # A small complete fixture lets tests exercise both-rank coverage and every
    # retained byte without reproducing root arithmetic in the Python reviewer.
    reference=struct.pack('<III',1,2,3);digest=hashlib.sha256(reference).hexdigest()
    for rank in (0,1):
        dest=directory/f'rank-{rank}';dest.mkdir()
        local=struct.pack('<II',1,rank+2);peer=struct.pack('<II',1,3-rank)
        for suffix,data in [('reference',reference),('merged',reference),('local',local),('peer',peer)]:
            (dest/f'case-0-{suffix}.bin').write_bytes(data)
        rows.extend([dict(kind='host',rank=rank,host=f'node{rank}',visible_devices='0'),
            dict(kind='fixture',rank=rank,case=0,s=8,genuine_support=True,support_count=2,sha256=digest),
            dict(kind='completed',rank=rank,cases=1)])
        for sample in (-1,0,1):
            row=dict(kind='measurement',rank=rank,case=0,s=8,sample=sample,support_count=2,local_count=1,peer_count=1,
                remote_bytes=8,owned_device_bytes=1024,peak_rss_bytes=2048,sha256=digest,matches=True)
            row.update({f:1.0 if f!='seconds' else 6.0 for f in HELP['FIELDS']})
            rows.append(row)
    return rows


class Review(unittest.TestCase):
    def test_complete_coefficient_evidence(self):
        with tempfile.TemporaryDirectory()as path:
            root=Path(path);rows=fixture(root)
            self.assertEqual(len(HELP['review'](rows,root,small=True)),6)
    def test_reject_missing_duplicate_or_wrong_owner(self):
        with tempfile.TemporaryDirectory()as path:
            root=Path(path);rows=fixture(root)
            for changed in (rows[:-1],rows+[rows[-1]],copy.deepcopy(rows)):
                if len(changed)==len(rows):changed[6]['host']='node0'
                with self.assertRaises(AssertionError):HELP['review'](changed,root,small=True)
    def test_reject_bad_resource_count_timing_or_match(self):
        with tempfile.TemporaryDirectory()as path:
            root=Path(path);rows=fixture(root)
            for change in (dict(owned_device_bytes=(128<<20)+1),dict(peak_rss_bytes=(2<<30)+1),dict(remote_bytes=4),
                dict(exchange_seconds=8),dict(matches=False),dict(local_count=2),dict(sha256='bad')):
                changed=copy.deepcopy(rows);changed[3].update(change)
                with self.assertRaises(AssertionError):HELP['review'](changed,root,small=True)
    def test_reject_corrupt_peer_or_merged_file(self):
        with tempfile.TemporaryDirectory()as path:
            root=Path(path);rows=fixture(root)
            for name in ('peer','merged','reference'):
                file=root/f'rank-0/case-0-{name}.bin';data=file.read_bytes()
                file.write_bytes(b'\x00'+data[1:])
                with self.assertRaises(AssertionError):HELP['review'](rows,root,small=True)
                file.write_bytes(data)


if __name__=='__main__':unittest.main()
