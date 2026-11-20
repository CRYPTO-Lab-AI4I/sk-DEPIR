#!/usr/bin/env python3
"""Exercise output, layout, proof and reuse rejection without GPUs/MPI."""
from pathlib import Path
import copy
import runpy
import tempfile
import unittest
H=runpy.run_path(str(Path(__file__).with_name('run.py')))

class ReviewTests(unittest.TestCase):
    def fixture(self,directory):
        rows=[dict(kind='host',rank=r,host=f'node{r}',visible_devices='0')for r in (0,1)];op=1
        for s,nr,nc in H['CASES']:
            label=f's{s}-r{nr}-c{nc}';width=4 if s==4 else 8
            for pass_id in range(8):
                a,b=(nr,nc)if pass_id%2==0 else(nc,nr)
                for rank in (0,1):
                    first,count=H['band'](b,rank);_,in_count=H['band'](a,rank)
                    out=b''.join(H['pattern']((column*nc+row if pass_id%2==0 else row*nc+column),width).to_bytes(width,'little')for row in range(first,first+count)for column in range(a))
                    for suffix in ('output','expected'):(directory/f'rank-{rank}-{label}-pass-{pass_id}-{suffix}.bin').write_bytes(out)
                    _,peer_cols=H['band'](b,1-rank);_,peer_rows=H['band'](a,1-rank);events=0
                    for direction,words in (('send',in_count*peer_cols),('receive',count*peer_rows)):
                        for offset in range(0,words,127):
                            rows.append(dict(kind='tile-proof',rank=rank,case=label,pass_id=pass_id,**{'pass':pass_id},operation=op,offset=offset,words=min(127,words-offset),peer=1-rank,round=1,direction=direction,own_proof='ab'*64,remote_proof='ab'*64));events+=1
                    rows.append(dict(kind='measurement',rank=rank,case=label,**{'pass':pass_id},s=s,rows=a,columns=b,operation=op,word_bytes=width,words=count*a,matches=True,source_preserved=True,third_preserved=True,reused=True,row_interleaved=bool(pass_id//2&1),column_interleaved=bool(pass_id//2&2),owned_device_bytes=1024,peak_rss_bytes=1024,seconds=.1,proof_events=events))
                op+=1
        rows.extend(dict(kind='completed',rank=r,cases=8*len(H['CASES']))for r in (0,1));return rows
    def test_pass(self):
        with tempfile.TemporaryDirectory()as d:
            result=H['review'](self.fixture(Path(d)),Path(d));self.assertEqual(len(result['measurements']),208)
    def test_reject_corrupt(self):
        with tempfile.TemporaryDirectory()as d:
            directory=Path(d);rows=self.fixture(directory);path=next(p for p in directory.glob('*-output.bin')if p.stat().st_size)
            path.write_bytes(b'\x00'*path.stat().st_size)
            with self.assertRaises(AssertionError):H['review'](rows,directory)
    def test_reject_metadata(self):
        with tempfile.TemporaryDirectory()as d:
            directory=Path(d);rows=self.fixture(directory)
            for key,value in (('source_preserved',False),('third_preserved',False),('reused',False),('owned_device_bytes',999),('row_interleaved',False)):
                bad=copy.deepcopy(rows);item=next(r for r in bad if r.get('kind')=='measurement'and r['pass']==2);item[key]=value
                with self.subTest(key=key),self.assertRaises(AssertionError):H['review'](bad,directory)
            bad=copy.deepcopy(rows);next(r for r in bad if r.get('kind')=='tile-proof')['words']+=1
            with self.assertRaises(AssertionError):H['review'](bad,directory)

    def test_selects_resident_driver(self):
        original=H['VECTORS']['suite'];calls=[]
        def fake(context,**kwargs):calls.append(kwargs);return 'prepared'
        H['VECTORS']['suite']=fake
        try:
            self.assertEqual(H['suite'](dict(run=lambda *args:None,resources={})), 'prepared')
            self.assertEqual(calls[0]['test_name'],'TestDistributedGPUStorage')
            self.assertIs(calls[0]['review_fn'],H['review'])
            self.assertNotIn('full_fft',calls[0])
        finally:H['VECTORS']['suite']=original

if __name__=='__main__':unittest.main()
