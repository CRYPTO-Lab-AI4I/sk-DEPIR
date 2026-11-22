#!/usr/bin/env python3
from pathlib import Path
import copy
import runpy
import unittest
HERE=Path(__file__).resolve().parent
H=runpy.run_path(str(HERE/'run.py'))


class ReviewTest(unittest.TestCase):
    def test_prefix_clone_and_upload_accounting(self):
        # N256,K97,target159 has8 updates. Rank1 owns[64,128):only the
        # lasttwo root prefixes and lasttwo square prefixes overlap that rank.
        r=dict(n=256,k=97,first=64,words=64,newton_input_mode='prefix',newton_steps=8,newton_root_clone_words_per_step=0,newton_input_upload_words=195)
        result=dict(measurements=[r]);self.assertIs(H['input_claims'](result,'prefix'),result)
        for field,value in [('newton_root_clone_words_per_step',64),('newton_input_upload_words',1024),('newton_steps',7),('newton_input_mode','dense')]:
            bad=copy.deepcopy(result);bad['measurements'][0][field]=value
            with self.assertRaises(AssertionError):H['input_claims'](bad,'prefix')
        r.update(newton_input_mode='dense',newton_root_clone_words_per_step=64,newton_input_upload_words=1024)
        self.assertIs(H['input_claims'](result,'dense'),result)

    def test_matched_control_keeps_other_memory_modes_constant(self):
        old=H['SCALE']['suite'];captured=[]
        H['SCALE']['suite']=lambda c,**kw:captured.append((dict(c['env']),kw))or{}
        try:
            H['suite'](dict(run=lambda *a:None,env={},resources={},go='/go'))
            env,kw=captured[0];self.assertEqual(env['PIR_PIPELINE_NEWTON_INPUT'],'prefix')
            self.assertEqual(env['PIR_PIPELINE_DISCOVERY_RESULT'],'stream');self.assertEqual(env['PIR_PIPELINE_SPECTRUM'],'stream')
            self.assertIs(kw['review_fn'],H['review'])
            self.assertEqual([a['name']for a in kw['extra_arms']],['dense','wide'])
            self.assertEqual(kw['extra_arms'][0]['env']['PIR_PIPELINE_NEWTON_INPUT'],'dense')
        finally:H['SCALE']['suite']=old


if __name__=='__main__':unittest.main()
