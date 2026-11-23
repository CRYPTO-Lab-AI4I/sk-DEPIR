#!/usr/bin/env python3
"""Reject incorrect real-count root admission and inherited restart configuration."""
from pathlib import Path
import copy
import hashlib
from unittest.mock import patch
import runpy
import unittest
H=runpy.run_path(str(Path(__file__).with_name('run.py')))
class ReviewTest(unittest.TestCase):
    def test_producing_manifest_needs_no_retained_source_directory(self):
        data=b'{"cases":[]}'
        key='experiments/distributed-owned-roots/root-references.json'
        raw=dict(revision='producer',source_sha256={key:hashlib.sha256(data).hexdigest()})
        with patch.object(H['subprocess'],'check_output',return_value=data)as read:
            self.assertEqual(H['producing_manifest'](raw),dict(cases=[]))
            self.assertEqual(read.call_args.args[0],['git','show','producer:'+key])
            raw['source_sha256'][key]='00'*32
            with self.assertRaises(AssertionError):H['producing_manifest'](raw)

    def test_actual_rounded_stage_admission(self):
        case=dict(label='s8',s=8,counts=[3442,1507,1541,895]);rows=[]
        for rank in range(4):rows.append(dict(kind='root-admission',case='s8',rank=rank,leaf_words=case['counts'][rank],largest_shard_words=4096 if rank<2 else 2048,max_local_words=8<<20,base_word_bytes=4,max_domain=65536,stages=2))
        self.assertEqual(len(H['admission_review'](rows,[case])),4)
        for field,value in (('largest_shard_words',2047),('leaf_words',0),('max_local_words',1024),('stages',1)):
            bad=copy.deepcopy(rows);bad[0][field]=value
            with self.assertRaises(AssertionError):H['admission_review'](bad,[case])
    def test_actual_owned_wide_arm_clears_restart(self):
        class Captured(Exception):pass
        captured={};old_suite=H['SCALE']['suite'];old_copy=H['OWN']['copy_references']
        def capture(c,**kwargs):captured.update(env=dict(c['env']),**kwargs);raise Captured()
        H['SCALE']['suite']=capture;H['OWN']['copy_references']=lambda *args:0
        try:
            context=dict(env={},run=lambda *args:None,go='go',source=Path(__file__).resolve().parents[2],cache=Path('/tmp/root-runtime-config'),resources={},save=lambda:None)
            with self.assertRaises(Captured):H['suite'](context)
        finally:H['SCALE']['suite']=old_suite;H['OWN']['copy_references']=old_copy
        self.assertEqual([a['name']for a in captured['extra_arms']],['owned-wide'])
        arm=captured['extra_arms'][0]
        self.assertEqual(arm['test'],'TestOwnedWideRoots')
        env=dict(captured['env'],PIR_PIPELINE_RESUME_FROM='/cold');env.update(arm['env'])
        self.assertEqual(env['PIR_PIPELINE_RESUME_FROM'],'')
        self.assertEqual(env['PIR_PIPELINE_DIAGNOSTIC_WRITE'],'bounded')

if __name__=='__main__':unittest.main()
