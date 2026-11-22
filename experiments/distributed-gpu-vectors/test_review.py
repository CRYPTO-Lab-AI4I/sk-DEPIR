#!/usr/bin/env python3
"""Validate exact operation coverage, output bytes and launcher shared paths."""
from pathlib import Path
import hashlib
import runpy
import tempfile
import unittest
H=runpy.run_path(str(Path(__file__).with_name('run.py')))


class ReviewTest(unittest.TestCase):
    def fixture(self,d,cases=None):
        cases=H['CASES']if cases is None else cases
        rows=[dict(kind='host',rank=r,host=f'node{r}',visible_devices='0')for r in (0,1)]
        for index,(s,dim)in enumerate(cases):
            label=f's{s}-d{dim}';nr,nc=1<<(dim-dim//2),1<<(dim//2);width=4 if s==4 else 8
            for p in (0,1):
                for rank in (0,1):
                    data=b''.join(((((i+1)*0x9e3779b9)^(1<<(2*s-1)))&((1<<(2*s))-1)).to_bytes(width,'little')for i in range(rank*nr*nc//2,(rank+1)*nr*nc//2))
                    for suffix in ('output','expected'):(d/f'rank-{rank}-{label}-pass-{p}-{suffix}.bin').write_bytes(data)
                    count=0
                    for op in (4*index+2*p+1,4*index+2*p+2):
                        for offset in range(0,nr*nc//4,127):
                            for direction in ('send','receive'):
                                count+=1;rows.append(dict(kind='tile-proof',rank=rank,case=label,pass_id=p,operation=op,offset=offset,words=min(127,nr*nc//4-offset),peer=1-rank,round=1,direction=direction,own_proof='00'*64,remote_proof='00'*64))
                                rows[-1]['pass']=rows[-1].pop('pass_id')
                    row=dict(kind='measurement',rank=rank,case=label,s=s,dimension=dim,rank_first=rank*nr//2,rank_rows=nr//2,word_bytes=width,words=nr*nc//2,matches=True,seconds=1,proof_events=count,owned_device_bytes=1,peak_rss_bytes=1,output_sha256=hashlib.sha256(data).hexdigest());row['pass']=p;rows.append(row)
        rows.extend(dict(kind='completed',rank=r,cases=2*len(cases))for r in (0,1));return rows

    def test_coverage_and_missing_packet(self):
        with tempfile.TemporaryDirectory()as tmp:
            d=Path(tmp);rows=self.fixture(d);self.assertEqual(len(H['review'](rows,d)['measurements']),24)
            i=next(i for i,r in enumerate(rows)if r.get('kind')=='tile-proof')
            with self.assertRaises(AssertionError):H['review'](rows[:i]+rows[i+1:],d)

    def test_corruption(self):
        with tempfile.TemporaryDirectory()as tmp:
            d=Path(tmp);rows=self.fixture(d);p=next(d.glob('*-output.bin'));a=bytearray(p.read_bytes());a[0]^=1;p.write_bytes(a)
            with self.assertRaises(AssertionError):H['review'](rows,d)

    def test_remote_launcher(self):
        class Captured(Exception):pass
        with tempfile.TemporaryDirectory()as tmp:
            root=Path(tmp);source=root/'local-source';cache=root/'shared';out=root/'logs'
            for p in (source/'experiments/distributed-root',cache,out):p.mkdir(parents=True)
            (source/'experiments/distributed-root/rank.sh').write_text('# frozen entry\n')
            (out/'ucx-config.log').write_text('UCX_RCACHE_ENABLE=n\n')
            checked=[]
            def run(label,argv,timeout=30):
                if label=='rank-preflight':
                    self.assertIn('--chdir='+str(cache),argv);self.assertNotIn(str(source),' '.join(map(str,argv)));checked.append(label)
                if label=='normal':
                    self.assertEqual(checked,['rank-preflight']);self.assertIn('--chdir='+str(cache),argv)
                    self.assertEqual(Path(argv[argv.index('bash')+1]),cache/'rank.sh')
                    self.assertEqual((cache/'rank.sh').read_text(),'# frozen entry\n')
                    self.assertNotIn(str(source),' '.join(map(str,argv)));raise Captured()
            with self.assertRaises(Captured):H['suite'](dict(run=run,go='go',source=source,cache=cache,out=out,env={},resources={},save=lambda:None))


if __name__=='__main__':unittest.main()
