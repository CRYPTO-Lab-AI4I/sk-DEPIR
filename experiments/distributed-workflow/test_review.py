#!/usr/bin/env python3
"""Verify assembled arm has a distinct test and clears inherited restart policy."""
from pathlib import Path
import runpy
import copy
import unittest
H=runpy.run_path(str(Path(__file__).with_name('run.py')))
class ReviewTest(unittest.TestCase):
    def test_workflow_metrics_reject_partial_identity_or_work(self):
        rows=[]
        for rank,uploaded in enumerate((103238,32767,16384,9003)):
            rows.append(dict(kind='workflow-result',rank=rank,case='s8',n=65536,k=7381,samples=2,identity_verified=True,reference_input=False,root_counts=[3442,1507,1541,895],refinement=dict(Steps=16,RootCloneWordsPerStep=0,UploadedWords=uploaded)))
        # Use an independently fixed total;distribution is separately tied to
        # exact baseline proof and output comparison in the full reviewer.
        rows[-1]['refinement']['UploadedWords']+=161392-sum(r['refinement']['UploadedWords']for r in rows)
        self.assertEqual(len(H['workflow_claims'](rows)),4)
        for field,value in (('identity_verified',False),('reference_input',True),('samples',1)):
            bad=copy.deepcopy(rows);bad[0][field]=value
            with self.assertRaises(AssertionError):H['workflow_claims'](bad)
        bad=copy.deepcopy(rows);bad[0]['refinement']['UploadedWords']+=1
        with self.assertRaises(AssertionError):H['workflow_claims'](bad)

    def test_actual_assembled_arm(self):
        class Captured(Exception):pass
        captured={};original=H['SCALE']['suite']
        def capture(c,**kwargs):captured.update(env=dict(c['env']),**kwargs);raise Captured()
        H['SCALE']['suite']=capture
        try:
            with self.assertRaises(Captured):H['suite'](dict(env={},run=lambda *args:None,go='go'))
        finally:H['SCALE']['suite']=original
        self.assertEqual([a['name']for a in captured['extra_arms']],['assembled'])
        arm=captured['extra_arms'][0]
        self.assertEqual(arm['test'],'TestAssembledRankWorkflow')
        env=dict(captured['env'],PIR_PIPELINE_RESUME_FROM='/cold');env.update(arm['env'])
        self.assertEqual(env['PIR_PIPELINE_RESUME_FROM'],'')
        self.assertEqual(env['PIR_PIPELINE_DIAGNOSTIC_WRITE'],'bounded')
if __name__=='__main__':unittest.main()
