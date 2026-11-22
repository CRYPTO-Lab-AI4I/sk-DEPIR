#!/usr/bin/env python3
"""Focused checks for new pipeline checkpoint/publication boundaries."""
from pathlib import Path
import hashlib
import json
import runpy
import tempfile
import unittest
H=runpy.run_path(str(Path(__file__).with_name('run.py')))


class ReviewTest(unittest.TestCase):
    def fixture(self,d,ranks=2):
        # Synthetic parser fixture only; real polynomial identities are checked
        # by the GPU pilot against the complete native preprocessing oracle.
        s=4;k=8;n=256;label='s4';word=lambda v:v.to_bytes(4,'little')
        p=word(1)+word(0)*(k-1)+word(2);g=word(1)*(n-k);spectrum=word(3)*n
        reference=dict(p=p,g=g,spectrum=spectrum)
        for stage,data in reference.items():(d/f'{label}-rank-0-reference-{stage}.bin').write_bytes(data)
        (d/'s4-support.bin').write_bytes(b'synthetic parser support')
        binding=hashlib.sha256(b'synthetic parser support').hexdigest()
        for rank in range(ranks):(d/f's4-rank-{rank}-support.bin').write_bytes(b''.join(i.to_bytes(8,'little')for i in range(1+rank*(k//ranks),1+(rank+1)*(k//ranks))))
        full=dict(p=p+b'\0'*(4*(n-k-1)),g=g+b'\0'*(4*k),spectrum=spectrum,identity=word(1)+word(0)*(n-2)+word(1))
        for stage,data in full.items():
            for rank in range(ranks):(d/f'{label}-rank-{rank}-{stage}.bin').write_bytes(data[rank*4*n//ranks:(rank+1)*4*n//ranks])
        for stage,count,data in [('root',k+1,p)]+[(f'newton-{v}',v,g[:4*v])for v in H['prefixes'](n-k)]:
            data+=b'\0'*(4*(n-count));hashes=[]
            for rank in range(ranks):
                block=data[rank*4*n//ranks:(rank+1)*4*n//ranks];(d/f'{label}-rank-{rank}-{stage}.bin').write_bytes(block);hashes.append(hashlib.sha256(block).hexdigest())
            (d/f'{label}-{stage}-commit.json').write_text(json.dumps(dict(s=s,k=k,n=n,prefix=count,ranks=ranks,support_sha256=binding,stage=stage,sha256=hashes)))
        code=b'LRSprobe1'+b''.join(word(v)for v in (s,2,k,n))+p+spectrum
        for stage in ('code','reference-code'):(d/f's4-{stage}.bin').write_bytes(code)
        return s,k

    def test_artifact_and_bad_publication(self):
        with tempfile.TemporaryDirectory()as tmp:
            d=Path(tmp);s,k=self.fixture(d)
            self.assertEqual(H['artifacts'](d,s,k)['checkpoint_prefixes'][-1],248)
            p=d/'s4-code.bin';data=bytearray(p.read_bytes());data[-1]^=1;p.write_bytes(data)
            with self.assertRaises(AssertionError):H['artifacts'](d,s,k)

    def test_checkpoint_binding_and_independent_prefix(self):
        with tempfile.TemporaryDirectory()as tmp:
            d=Path(tmp);s,k=self.fixture(d);p=d/'s4-newton-8-commit.json';record=json.loads(p.read_text());original=record['support_sha256'];record['support_sha256']='00'*32;p.write_text(json.dumps(record))
            with self.assertRaises(AssertionError):H['artifacts'](d,s,k)
            record['support_sha256']=original
            block=d/'s4-rank-0-newton-8.bin';data=bytearray(block.read_bytes());data[4]^=1;block.write_bytes(data);record['sha256'][0]=hashlib.sha256(data).hexdigest();p.write_text(json.dumps(record))
            with self.assertRaises(AssertionError):H['artifacts'](d,s,k)

    def test_only_two_integrated_arms(self):
        captured=[];old=H['VECTORS']['suite'];H['VECTORS']['suite']=lambda c,**kw:captured.append(kw)
        try:
            H['suite'](dict(run=lambda *a:None,env={},go='go'))
            self.assertFalse(captured[0]['full_fft'])
            self.assertEqual([(a['name'],a['memcheck'],a['env']['PIR_PIPELINE_SMALL'])for a in captured[0]['arms']],[('normal',False,'0'),('memcheck',True,'1')])
            self.assertTrue(all(a['test']=='TestDistributedPreprocessingPilot'for a in captured[0]['arms']))
        finally:H['VECTORS']['suite']=old


if __name__=='__main__':unittest.main()
