#!/usr/bin/env python3
"""Exercise actual resident control options before GPU allocation work."""
from pathlib import Path
import runpy
import unittest
H=runpy.run_path(str(Path(__file__).with_name('run.py')))
class ReviewTest(unittest.TestCase):
    def test_actual_resident_control_clears_restart(self):
        class Captured(Exception):pass
        captured={};original=H['SCALE']['suite']
        def capture(c,**kwargs):captured.update(env=dict(c['env']),**kwargs);raise Captured()
        H['SCALE']['suite']=capture
        try:
            with self.assertRaises(Captured):H['suite'](dict(env={},run=lambda *args:None,go='go'))
        finally:H['SCALE']['suite']=original
        self.assertEqual(captured['env']['PIR_PIPELINE_DISCOVERY_RESULT'],'stream')
        self.assertEqual([a['name']for a in captured['extra_arms']],['resident-discovery'])
        env=dict(captured['env'],PIR_PIPELINE_RESUME_FROM='/cold');env.update(captured['extra_arms'][0]['env'])
        self.assertEqual(env['PIR_PIPELINE_RESUME_FROM'],'')
        self.assertEqual(env['PIR_PIPELINE_DISCOVERY_RESULT'],'resident')
        self.assertEqual(env['PIR_PIPELINE_SPECTRUM'],'stream')
        self.assertEqual(env['PIR_PIPELINE_NEWTON_INPUT'],'prefix')
if __name__=='__main__':unittest.main()
