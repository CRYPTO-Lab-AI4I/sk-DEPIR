#!/usr/bin/env python3
"""Exercise the real suite's selection and reject corrupt/misbound commits."""
from pathlib import Path
import hashlib
import json
import os
import runpy
import tempfile
import unittest
H=runpy.run_path(str(Path(__file__).with_name('run.py')))
class ReviewTest(unittest.TestCase):
    def test_actual_suite_selects_commit(self):
        class Captured(Exception):pass
        original=H['PUBLICATION']['suite'];captured={}
        def capture(c):captured.update(c['env']);raise Captured()
        H['PUBLICATION']['suite']=capture
        try:
            with self.assertRaises(Captured):H['suite'](dict(env={},run=lambda *args:None,go='go',cache=Path('/tmp')))
        finally:H['PUBLICATION']['suite']=original
        self.assertEqual(captured['PIR_WORKFLOW_CHECKPOINT'],'commit')
    def test_independent_commits(self):
        with tempfile.TemporaryDirectory()as tmp:
            p=Path(tmp);archive=bytes(8*7381);(p/'s8-archive.raw').write_bytes(archive);os.link(p/'s8-archive.raw',p/'s8-support.bin')
            source=dict(version=1,header=dict(K=7381,N=65536),Ranks=4,Samples=2,SupportFormat='uint64-exponents-le',SupportSHA=hashlib.sha256(archive).hexdigest(),SupportBytes=len(archive))
            body=json.dumps(source).encode();(p/'s8-checkpoint-source.json').write_bytes(body)
            (p/'s8-publication.json').write_text(json.dumps(dict(header=source['header'],samples=2,support_format=source['SupportFormat'])))
            rows=[dict(kind='workflow-result',rank=r,case='s8',n=65536,k=7381,samples=2,root_counts=[3442,1507,1541,895],identity_verified=True,reference_input=False,refinement=dict(Steps=16,RootCloneWordsPerStep=0,UploadedWords=40348),checkpoint_mode='commit')for r in range(4)]
            for stage,prefix in [('root',7382)]+[(f'newton-{v}',v)for v in H['WORKFLOW']['OWN']['PILOT']['prefixes'](58155)]:
                digests=[]
                for rank in range(4):
                    data=bytes(4*min(16384,max(0,prefix-rank*16384)));(p/f's8-rank-{rank}-{stage}.bin').write_bytes(data);digests.append(hashlib.sha256(data).hexdigest())
                record=dict(layout='prefix',s=8,k=7381,n=65536,prefix=prefix,ranks=4,support_sha256=source['SupportSHA'],stage=stage,source_sha256=hashlib.sha256(body).hexdigest(),sha256=digests)
                (p/f's8-{stage}-commit.json').write_text(json.dumps(record));rows.extend(dict(kind='workflow-checkpoint',case='s8',rank=rank,record=record)for rank in range(4))
            result=H['checkpoint_review'](rows,p);self.assertEqual(len(result['commits']),17)
            path=p/'s8-root-commit.json';record=json.loads(path.read_text());record['source_sha256']='0'*64;path.write_text(json.dumps(record))
            with self.assertRaises(AssertionError):H['checkpoint_review'](rows,p)
            record['source_sha256']=result['source_sha256'];path.write_text(json.dumps(record));(p/'s8-rank-0-root.bin').write_bytes(b'bad')
            with self.assertRaises(AssertionError):H['checkpoint_review'](rows,p)
if __name__=='__main__':unittest.main()
