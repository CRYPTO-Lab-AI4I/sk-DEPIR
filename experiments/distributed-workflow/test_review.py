#!/usr/bin/env python3
"""Verify assembled arm has a distinct test and clears inherited restart policy."""
from pathlib import Path
import runpy
import copy
import json
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

    def test_reference_independent_discovery_uses_external_native_data(self):
        rows=[]
        for rank in range(4):
            rows.append(dict(kind='discovery-result',s=2,rank=rank,k=4,samples=1,local_good=1,raw_count=1,root_count=1,grid_words=4,first=4*rank,owned_device_bytes=1,peak_rss_bytes=1,entropy_sha256='ab'*32 if rank==0 else''))
            rows.append(dict(kind='discovery-sample',s=2,rank=rank,k=4,sample=1,local_count=1,global_count=4,good_counts=[1]*4,operation_start=1))
        reference=dict(parameters=dict(spec=dict(FieldDegree=2,Dimension=2),n=16,k=4,samples=1),entropy_sha256='ab'*32)
        review=H['OWN']['DISCOVERY']['summary']
        with self.assertRaises(AssertionError):review(rows,2,4)
        self.assertEqual(review(rows,2,4,reference)[1:],(4,1))
        bad=copy.deepcopy(reference);bad['entropy_sha256']='cd'*32
        with self.assertRaises(AssertionError):review(rows,2,4,bad)
        bad=copy.deepcopy(reference);bad['parameters']['k']=3
        with self.assertRaises(AssertionError):review(rows,2,4,bad)

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
