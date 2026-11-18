#!/usr/bin/env python3
import json
import hashlib
from pathlib import Path
import runpy
import tempfile
import unittest
VERIFY=runpy.run_path(str(Path(__file__).with_name('run.py')))['review_output']


class WorkerReview(unittest.TestCase):
    def test_complete_and_reject_unbalanced_missing_or_corrupt(self):
        with tempfile.TemporaryDirectory() as temporary:
            root=Path(temporary);data=b'complete-reference';digest=hashlib.sha256(data).hexdigest()
            (root/'case-0-reference.bin').write_bytes(data)
            rows=[]
            for workers in (0,1,2,4):
                for sample in (-1,0,1):
                    name=str(workers)+'.bin' if sample==1 else ''
                    if name:(root/name).write_bytes(data)
                    rows.append(dict(kind='measurement',case=0,independent_workers=workers,sample=sample,
                        matches=True,sha256=digest,seconds=1,owned_device_bytes=1024,support_count=8,
                        partition_sizes=[8//workers]*workers if workers else [],file=name))
            rows.append(dict(kind='completed',cases=1));log=root/'log'
            def write():log.write_text(''.join(json.dumps(row)+'\n' for row in rows))
            write();self.assertEqual(len(VERIFY(log,root,1)),12)
            rows[6]['partition_sizes']=[7,1];write()
            with self.assertRaises(AssertionError):VERIFY(log,root,1)
            rows[6]['partition_sizes']=[4,4];rows.pop();write()
            with self.assertRaises(AssertionError):VERIFY(log,root,1)
            rows.append(dict(kind='completed',cases=1));write();(root/'4.bin').write_bytes(b'corrupt')
            with self.assertRaises(AssertionError):VERIFY(log,root,1)


if __name__=='__main__':unittest.main()
