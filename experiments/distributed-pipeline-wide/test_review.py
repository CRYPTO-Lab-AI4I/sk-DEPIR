#!/usr/bin/env python3
from pathlib import Path
import runpy
import tempfile
import unittest
HERE=Path(__file__).resolve().parent
H=runpy.run_path(str(HERE/'run.py'))


class ReviewTest(unittest.TestCase):
    def test_independent_reciprocal_and_full_convolution(self):
        for count,target in ((9,16),(17,64)):
            p,g=H['reciprocal'](count,target);out=[0]*(count+target-1)
            for i,a in enumerate(p):
                for j,b in enumerate(g):out[i+j]^=H['multiply'](a,b)
            self.assertEqual(out[:target],[1]+[0]*(target-1))
            self.assertTrue(any(out[target:]))
            self.assertLess(max(g),1<<18)

    def test_stream_width_and_truncated_high_bits(self):
        with tempfile.TemporaryDirectory()as tmp:
            d=Path(tmp);n=32;target=16;label='bounded'
            generator=(1).to_bytes(4,'little')+bytes(4*(n-1));spectrum=b''.join(i.to_bytes(8,'little')for i in range(n))
            high=b''.join((((1<<35)^((i+1)*0x9e3779b9))&((1<<36)-1)).to_bytes(8,'little')for i in range(n))
            for suffix,data in [('generator',generator[:4*target]),('spectrum',spectrum),('high',high)]:
                (d/f'{label}-stream-{suffix}.bin').write_bytes(data)
            H['review_streams'](d,label,n,target,generator,spectrum,high)
            p=d/f'{label}-stream-high.bin';p.write_bytes(b''.join((int.from_bytes(high[i:i+8],'little')&0xffffffff).to_bytes(8,'little')for i in range(0,len(high),8)))
            with self.assertRaises(AssertionError):H['review_streams'](d,label,n,target,generator,spectrum,high)
            p.write_bytes(high);(d/f'{label}-stream-spectrum.bin').write_bytes(spectrum[:4*n])
            with self.assertRaises(AssertionError):H['review_streams'](d,label,n,target,generator,spectrum,high)

    def test_three_arms_share_admitted_runner(self):
        old=H['FOUR']['suite'];captured=[]
        H['FOUR']['suite']=lambda c,**kw:captured.append(kw)or{}
        try:
            H['suite'](dict(run=lambda *a:None,resources={}))
            arms=captured[0]['extra_arms'];self.assertEqual(len(arms),1)
            self.assertEqual((arms[0]['name'],arms[0]['test'],arms[0]['memcheck'],arms[0]['profile']),('wide','TestWidePipelineTail',False,True))
            self.assertIs(arms[0]['review'],H['wide_review'])
        finally:H['FOUR']['suite']=old


if __name__=='__main__':unittest.main()
