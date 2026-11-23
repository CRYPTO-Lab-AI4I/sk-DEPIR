#!/usr/bin/env python3
"""Verify real arm settings reset dense/prefix and restart state before launch."""
from pathlib import Path
import runpy
import unittest
H=runpy.run_path(str(Path(__file__).with_name('run.py')))
class ReviewTest(unittest.TestCase):
    def test_actual_arm_transitions(self):
        class Captured(Exception):pass
        captured={};original=H['SCALE']['suite']
        def capture(c,**kwargs):captured.update(env=dict(c['env']),**kwargs);raise Captured()
        H['SCALE']['suite']=capture
        try:
            with self.assertRaises(Captured):H['suite'](dict(env={},run=lambda *args:None,go='go'))
        finally:H['SCALE']['suite']=original
        self.assertEqual([a['name']for a in captured['extra_arms']],['dense','wide'])
        env=dict(captured['env'],PIR_PIPELINE_RESUME_FROM='/cold')
        for arm in captured['extra_arms']:
            env.update(arm['env'])
            self.assertEqual(env['PIR_PIPELINE_RESUME_FROM'],'')
            self.assertEqual(env['PIR_PIPELINE_NEWTON_INPUT'],'dense'if arm['name']=='dense'else'prefix')
            self.assertEqual(env['PIR_PIPELINE_DIAGNOSTIC_WRITE'],'bounded')
if __name__=='__main__':unittest.main()
