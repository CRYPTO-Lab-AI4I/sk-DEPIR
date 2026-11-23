#!/usr/bin/env python3
"""Reject omitted hashes,unexpected coordinator ownership and missing agreement."""
from pathlib import Path
import copy
import json
import hashlib
import runpy
import tempfile
import unittest
H=runpy.run_path(str(Path(__file__).with_name('run.py')))
class ReviewTest(unittest.TestCase):
    def test_support_claims(self):
        with tempfile.TemporaryDirectory()as tmp:
            p=Path(tmp);(p/'s8-support.bin').write_bytes(b'12345678')
            records={}
            for stage,prefix in (('root',98),('newton-8',8)):
                record=dict(k=97,layout='prefix',n=65536,prefix=prefix,ranks=4,s=8,sha256=['00'*32]*4,stage=stage,support_sha256=hashlib.sha256(b'12345678').hexdigest())
                records[stage]=record;(p/('s8-'+stage+'-commit.json')).write_text(json.dumps(record,sort_keys=True))
            for mode in ('coordinator','local'):
                rows=[]
                for rank in range(4):
                    passes=(1 if rank==0 else 0)if mode=='coordinator'else 2
                    rows.append(dict(kind='restart-support',rank=rank,mode=mode,verified=True,hash_passes=passes,hashed_bytes=8*passes,agreements=2 if mode=='coordinator'else 0))
                if mode=='coordinator':
                    for rank in range(4):
                        for stage,record in records.items():
                            digest=H['binding_frame'](record,p,'s8')
                            rows.append(dict(kind='restart-support-agreement',rank=rank,checkpoint=stage,case='s8',source=str(p),frames=[digest+(record['support_sha256']if peer==0 else'00'*32)for peer in range(4)]))
                self.assertEqual(H['support_claims'](rows,p,mode)['hash_passes'],1 if mode=='coordinator'else 8)
                if mode=='coordinator':
                    bad=copy.deepcopy(rows);bad[-1]['frames'][1]='00'*64
                    with self.assertRaises(AssertionError):H['support_claims'](bad,p,mode)
                for field,value in (('hashed_bytes',999),('agreements',99),('mode','bad'),('verified',False)):
                    bad=copy.deepcopy(rows);bad[1][field]=value
                    with self.assertRaises(AssertionError):H['support_claims'](bad,p,mode)
                with self.assertRaises(AssertionError):H['support_claims'](rows[:-1],p,mode)
    def test_suite_mode_transition(self):
        class Captured(Exception):pass
        original=H['SCALE']['suite'];captured={}
        def capture(c,**kwargs):captured.update(env=c['env'],arms=kwargs['extra_arms']);raise Captured()
        H['SCALE']['suite']=capture
        try:
            with self.assertRaises(Captured):H['suite'](dict(env={},cache=Path('/tmp/binding'),run=lambda *args:None,go='go'))
        finally:H['SCALE']['suite']=original
        self.assertEqual(captured['env']['PIR_PIPELINE_RESTART_SUPPORT'],'coordinator')
        self.assertEqual([a['name']for a in captured['arms']],['local-resume'])
        self.assertEqual(captured['arms'][0]['env']['PIR_PIPELINE_RESTART_SUPPORT'],'local')
        self.assertEqual(captured['arms'][0]['env']['PIR_PIPELINE_RESUME_FROM'],'/tmp/binding/cold')
    def test_metadata_codec_golden(self):
        record=dict(layout='prefix',s=4,k=97,n=256,prefix=98,ranks=2,support_sha256='00'*32,stage='root',sha256=['00'*32,'01'*32])
        self.assertEqual(H['binding_frame'](record,'/shared/checkpoint','s4'),'59d6e0931e823a7c6d5502a2114947d81df298945f2373bc7689d61af281f35a')

if __name__=='__main__':unittest.main()
