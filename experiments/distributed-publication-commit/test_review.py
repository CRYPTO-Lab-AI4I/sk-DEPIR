#!/usr/bin/env python3
"""Reject invalid committed manifests/hashes and implicit publication selection."""
from pathlib import Path
import copy
import hashlib
import json
import runpy
import tempfile
import unittest
H=runpy.run_path(str(Path(__file__).with_name('run.py')))
class ReviewTest(unittest.TestCase):
    def test_publication_manifest(self):
        with tempfile.TemporaryDirectory()as tmp:
            p=Path(tmp);files=[]
            for name in ('s8-code.bin','s8-generator.bin','s8-archive.raw'):
                data=name.encode();(p/name).write_bytes(data);files.append(dict(name=name,bytes=len(data),sha256=hashlib.sha256(data).hexdigest()))
            record=dict(version=1,identity_verified=True,samples=2,support_format='uint64-exponents-le',header=dict(K=7381,N=65536),files=files)
            (p/'s8-publication.json').write_text(json.dumps(record))
            rows=[dict(kind='workflow-publication',rank=0,record=record)]
            self.assertEqual(H['WORKFLOW']['publication_review'](rows,p),record)
            (p/'s8-code.bin').write_bytes(b'bad')
            with self.assertRaises(AssertionError):H['WORKFLOW']['publication_review'](rows,p)
    def test_actual_suite_selects_commit(self):
        class Captured(Exception):pass
        original=H['WORKFLOW']['suite'];captured={}
        def capture(c):captured.update(c['env']);raise Captured()
        H['WORKFLOW']['suite']=capture
        try:
            with self.assertRaises(Captured):H['suite'](dict(env={},run=lambda *args:None,go='go'))
        finally:H['WORKFLOW']['suite']=original
        self.assertEqual(captured['PIR_WORKFLOW_PUBLICATION'],'commit')
if __name__=='__main__':unittest.main()
