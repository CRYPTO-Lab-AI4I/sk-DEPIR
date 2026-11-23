#!/usr/bin/env python3
"""Verify actual four-arm wiring and resumed-only completion counters."""
from pathlib import Path
import copy
import runpy
import unittest
H=runpy.run_path(str(Path(__file__).with_name('run.py')))
class ReviewTest(unittest.TestCase):
    def test_actual_nested_suite(self):
        class Captured(Exception):pass
        captured={};original=H['SCALE']['suite']
        def capture(c,**kw):captured.update(env=dict(c['env']),arms=kw['extra_arms']);raise Captured()
        H['SCALE']['suite']=capture
        try:
            with self.assertRaises(Captured):H['suite'](dict(env={},run=lambda *args:None,go='go',cache=Path('/tmp/cache'),out=Path('/tmp/out')))
        finally:H['SCALE']['suite']=original
        env=captured['env'];self.assertEqual(env['PIR_WORKFLOW_RESUME_FROM'],'')
        self.assertEqual(env['PIR_WORKFLOW_CHECKPOINT'],'commit');self.assertEqual(env['PIR_WORKFLOW_PUBLICATION'],'commit')
        self.assertEqual([a['name']for a in captured['arms']],['assembled','assembled-resume'])
        for arm in captured['arms']:env.update(arm['env'])
        self.assertEqual(env['PIR_WORKFLOW_RESUME_FROM'],'/tmp/cache/assembled')
    def test_resumed_claims(self):
        rows=[dict(kind='workflow-result',rank=r,case='s8',n=65536,k=7381,samples=2,root_counts=None,identity_verified=True,reference_input=False,resumed=True,initial_prefix=8,refinement=dict(Steps=13,RootCloneWordsPerStep=0,UploadedWords=161367 if r==0 else 0))for r in range(4)]
        self.assertEqual(len(H['WORKFLOW']['workflow_claims'](rows,known=8)),4)
        for key,value in (('resumed',False),('initial_prefix',1),('identity_verified',False)):
            bad=copy.deepcopy(rows);bad[0][key]=value
            with self.assertRaises(AssertionError):H['WORKFLOW']['workflow_claims'](bad,known=8)
        bad=copy.deepcopy(rows);bad[0]['refinement']['Steps']=16
        with self.assertRaises(AssertionError):H['WORKFLOW']['workflow_claims'](bad,known=8)
if __name__=='__main__':unittest.main()
