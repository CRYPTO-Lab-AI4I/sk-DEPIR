#!/usr/bin/env python3
"""Product result/base-field/tail checks and exact three-FFT route coverage."""
from pathlib import Path
import hashlib
import runpy
import tempfile
import unittest
H=runpy.run_path(str(Path(__file__).with_name('run.py')))


class ReviewTest(unittest.TestCase):
    def fixture(self,d,cases=None,ranks=2):
        cases=H['VECTORS']['CASES']if cases is None else cases
        rows=[dict(kind='host',rank=r,host=f'node{r//(ranks//2)}',visible_devices='0')for r in range(ranks)];operation=1
        for s,dim in cases:
            n=1<<dim;cols=1<<(dim//2);levels=dim-dim//2;capacity=min(n,3*max(cols,n//cols));span=2+2*levels
            rows.extend(dict(kind='pointwise-check',rank=r,s=s,dimension=dim,words=min(capacity,17),matches=True)for r in range(ranks))
            for shape in (0,1):
                label=f's{s}-d{dim}-shape-{shape}';ac,bc=(n//4+1,3*n//4)if shape==0 else(n//4,n//2);proofs={r:0 for r in range(ranks)};windows={r:0 for r in range(ranks)}
                for child in (0,1,2):
                    start=operation+child*span;inverse=child==2;matrix=start+(0 if inverse else 2*levels)
                    for rank in range(ranks):
                        for op in (matrix,matrix+1):
                            for peer in range(ranks):
                                if peer==rank:continue
                                for offset in range(0,n//(ranks*ranks),127):
                                    for direction in ('send','receive'):
                                        round_id=((peer-rank)if direction=='send'else(rank-peer))%ranks
                                        rows.append(dict(kind='tile-proof',rank=rank,case=label,operation=op,offset=offset,words=min(127,n//(ranks*ranks)-offset),peer=peer,round=round_id,direction=direction,own_proof='00'*64,remote_proof='00'*64,**{'pass':child}));proofs[rank]+=1
                    for step in range(levels):
                        level=step if inverse else levels-1-step;degrees=[1<<level]
                        rows.extend(dict(kind='taylor-plan',rank=r,case=label,level=level,degrees=degrees,capacity=capacity,operation_start=start,**{'pass':child})for r in range(ranks))
                        for stage in (0,1):
                            q=(stage==0)!=inverse;op=start+(2 if inverse else 0)+2*step+stage
                            for target,source,block,first,count,slot,a,words in H['FFT']['expected_windows'](n,cols,degrees,level,q,capacity,ranks):
                                packet=dict(Operation=op,Domain=n,Columns=cols,Level=level,Quotient=q,Target=dict(Block=block,First=first,Count=count),Slot=slot,Source=a,Words=words)
                                for rank,peer,direction in ((source,target,'send'),(target,source,'receive')):
                                    rows.append(dict(kind='taylor-proof',rank=rank,case=label,peer=peer,direction=direction,packet=packet,own_proof='00'*64,remote_proof='00'*64,**{'pass':child}));windows[rank]+=1
                for rank in range(ranks):
                    # Schoolbook product of all-one operands: coefficient parity.
                    coeff=[]
                    for k in range(rank*n//ranks,(rank+1)*n//ranks):coeff.append(max(0,min(k,ac-1)-max(0,k-bc+1)+1)%2 if k<ac+bc-1 else 0)
                    data=b''.join(v.to_bytes(4,'little')for v in coeff)
                    for suffix in ('output','expected'):(d/f'rank-{rank}-{label}-{suffix}.bin').write_bytes(data)
                    rows.append(dict(kind='measurement',rank=rank,case=label,s=s,dimension=dim,shape=shape,a_count=ac,b_count=bc,first_coefficient=rank*n//ranks,words=n//ranks,slot_bytes=4 if s==4 else 8,output_word_bytes=4,seconds=1,operation_start=operation,proof_events=proofs[rank],taylor_proof_events=windows[rank],owned_device_bytes=1,peak_rss_bytes=1,output_sha256=hashlib.sha256(data).hexdigest(),matches=True))
                operation+=3*span
        rows.extend(dict(kind='completed',rank=r,cases=2*len(cases))for r in range(ranks));return rows

    def test_complete_and_missing_source(self):
        with tempfile.TemporaryDirectory()as tmp:
            d=Path(tmp);rows=self.fixture(d);self.assertEqual(len(H['review'](rows,d)['measurements']),24)
            i=next(i for i,r in enumerate(rows)if r.get('kind')=='taylor-proof')
            with self.assertRaises(AssertionError):H['review'](rows[:i]+rows[i+1:],d)

    def test_output_corruption(self):
        with tempfile.TemporaryDirectory()as tmp:
            d=Path(tmp);rows=self.fixture(d);p=next(d.glob('*-output.bin'));a=bytearray(p.read_bytes());a[0]^=1;p.write_bytes(a)
            with self.assertRaises(AssertionError):H['review'](rows,d)

    def test_product_runner_selection(self):
        class Captured(Exception):pass
        calls=[]
        def suite(c,**kwargs):
            self.assertEqual(kwargs['test_name'],'TestDistributedGPUProducts');self.assertTrue(kwargs['full_fft']);self.assertIs(kwargs['review_fn'],H['review']);raise Captured()
        original=H['VECTORS']['suite'];H['VECTORS']['suite']=suite
        try:
            with self.assertRaises(Captured):H['suite'](dict(run=lambda label,argv,timeout:calls.append(label)))
            self.assertEqual(calls,['product-review-tests','fft-review-tests'])
        finally:H['VECTORS']['suite']=original


if __name__=='__main__':unittest.main()
