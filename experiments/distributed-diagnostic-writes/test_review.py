#!/usr/bin/env python3
"""Reject incomplete or inconsistent diagnostic write progress."""
from pathlib import Path
import json
import runpy
import tempfile
import unittest
H=runpy.run_path(str(Path(__file__).with_name('run.py')))
class ReviewTest(unittest.TestCase):
    def test_actual_suite_restores_bounded_mode_after_bulk(self):
        # Exercise the actual suite configuration before GPU/build work, then
        # reproduce the maintained launcher's successive environment updates.
        class Captured(Exception):pass
        captured={}
        original=H['SCALE']['suite']
        def capture(c,**kwargs):
            captured.update(base=dict(c['env']),arms=kwargs['extra_arms'])
            raise Captured()
        H['SCALE']['suite']=capture
        try:
            with self.assertRaises(Captured):H['suite'](dict(env={},run=lambda *args:None,go='go'))
        finally:H['SCALE']['suite']=original
        env=captured['base'];self.assertEqual(env['PIR_PIPELINE_DIAGNOSTIC_WRITE'],'bounded')
        for arm in captured['arms']:
            env.update(arm['env'])
            self.assertEqual(env['PIR_PIPELINE_DIAGNOSTIC_WRITE'],'full'if arm['name']=='bulk'else'bounded')

    def test_targeted_retry_runs_only_bounded_wide(self):
        target=runpy.run_path(str(Path(__file__).with_name('run-wide.py')))
        class Captured(Exception):pass
        captured={};original=target['H']['VECTORS']['suite']
        def capture(c,**kwargs):
            captured.update(env=c['env'],**kwargs);raise Captured()
        target['H']['VECTORS']['suite']=capture
        try:
            with self.assertRaises(Captured):target['suite'](dict(env={'PIR_PIPELINE_DIAGNOSTIC_WRITE':'full'},run=lambda *args:None))
        finally:target['H']['VECTORS']['suite']=original
        self.assertEqual(captured['ranks'],4)
        self.assertEqual([a['name']for a in captured['arms']],['wide'])
        self.assertEqual(captured['env']['PIR_PIPELINE_DIAGNOSTIC_WRITE'],'bounded')
        self.assertEqual(captured['arms'][0]['env']['PIR_PIPELINE_DIAGNOSTIC_WRITE'],'bounded')

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
