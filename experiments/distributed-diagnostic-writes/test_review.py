#!/usr/bin/env python3
"""Reject incomplete or inconsistent diagnostic write progress."""
from pathlib import Path
import json
import runpy
import tempfile
import unittest
H=runpy.run_path(str(Path(__file__).with_name('run.py')))
class ReviewTest(unittest.TestCase):
    def test_diagnostic_progress(self):
        with tempfile.TemporaryDirectory()as tmp:
            p=Path(tmp);rows=[]
            for rank in range(4):
                name=f'rank-{rank}.bin';(p/name).write_bytes(b'1234')
                row=dict(work='diagnostic-write',mode='bounded',unit='bytes',phase=name,total=4)
                rows.append([dict(row,completed=0),dict(row,completed=4)])
            def save():
                for rank,items in enumerate(rows):(p/f'cold-{rank}.stderr.log').write_text(''.join(json.dumps(r)+'\n'for r in items))
            save();self.assertEqual(H['diagnostic_review'](p,'cold',p)['bytes'],16)
            rows[0][-1]['completed']=3;save()
            with self.assertRaises(AssertionError):H['diagnostic_review'](p,'cold',p)
            rows[0][-1]['completed']=4;rows[0][-1]['total']=5;save()
            with self.assertRaises(AssertionError):H['diagnostic_review'](p,'cold',p)
            rows[0][-1]['total']=4;rows[0][-1]['event']='failed';save()
            with self.assertRaises(AssertionError):H['diagnostic_review'](p,'cold',p)
if __name__=='__main__':unittest.main()
