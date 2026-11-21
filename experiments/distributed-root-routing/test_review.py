#!/usr/bin/env python3
"""Operand proofs must match exact source coefficients and target ownership."""
from pathlib import Path
import copy
import hashlib
import runpy
import tempfile
import unittest
HERE=Path(__file__).resolve().parent
H=runpy.run_path(str(HERE/'run.py'))


class ReviewTest(unittest.TestCase):
    def fixture(self,directory,layout='routed'):
        previous=runpy.run_path(str(HERE.parent/'distributed-root-merge/test_review.py'))['ReviewTest']
        rows=previous().fixture(directory);n=4096;count=n//2-1;label='s17-d12-shape-0'
        source=(directory/f'rank-1-{label}-local-root.bin').read_bytes()
        for rank in (0,1):
            for operand,suffix in enumerate(('left','right')):
                local=(directory/f'rank-{operand}-{label}-local-root.bin').read_bytes()
                data=(local+b'\0'*(4*n-len(local)))[rank*2*n:(rank+1)*2*n]
                (directory/f'rank-{rank}-{label}-{suffix}-input.bin').write_bytes(data)
            r=next(r for r in rows if r.get('kind')=='root-input'and r['rank']==rank)
            sent,received=(count,count)if layout=='peer'else((0,count)if rank==0 else(count,0))
            r.update(input_layout=layout,peer_materialized=layout=='peer',sent_words=sent,received_words=received,route_tile_words=4096)
            if layout=='routed':(directory/f'rank-{rank}-{label}-peer-root.bin').unlink()
        if layout=='routed':
            proof=hashlib.sha256(source).hexdigest()*2
            packet=dict(Operation=1,Domain=n,Operand=1,Offset=0,Words=count,Source=1,Target=0)
            rows.extend(dict(kind='operand-proof',rank=rank,direction=direction,packet=packet,own_proof=proof,remote_proof=proof)for rank,direction in ((1,'send'),(0,'receive')))
        return rows

    def test_routed_and_peer(self):
        for layout in ('peer','routed'):
            with tempfile.TemporaryDirectory()as tmp:
                d=Path(tmp);rows=self.fixture(d,layout)
                result=H['review'](rows,d,dict(dim=12,layout=layout))
                self.assertEqual(result['operand_proof_events'],2 if layout=='routed'else 0)

    def test_bad_route_or_payload(self):
        with tempfile.TemporaryDirectory()as tmp:
            d=Path(tmp);rows=self.fixture(d);spec=dict(dim=12,layout='routed')
            index=next(i for i,r in enumerate(rows)if r.get('kind')=='operand-proof')
            with self.assertRaises(AssertionError):H['review'](rows[:index]+rows[index+1:],d,spec)
            bad=copy.deepcopy(rows);bad[index]['packet']['Operand']=0
            with self.assertRaises(AssertionError):H['review'](bad,d,spec)
            # Both peers can agree on the wrong digest: compare against saved
            # source coefficients as well as the paired proof endpoints.
            bad=copy.deepcopy(rows)
            for r in bad:
                if r.get('kind')=='operand-proof':r.update(own_proof='00'*64,remote_proof='00'*64)
            with self.assertRaises(AssertionError):H['review'](bad,d,spec)
            p=d/'rank-0-s17-d12-shape-0-right-input.bin';data=bytearray(p.read_bytes());data[40]^=1;p.write_bytes(data)
            with self.assertRaises(AssertionError):H['review'](rows,d,spec)

    def test_plan_and_race_environment(self):
        plan=H['specs']();self.assertEqual(len(plan),5)
        self.assertEqual([(s['layout'],s['dim'])for s in plan],[('peer',16),('routed',12),('routed',12),('routed',14),('routed',16)])
        original=H['VECTORS']['suite'];selected=[];calls=[];env={}
        H['VECTORS']['suite']=lambda c,**kw:selected.append(kw)
        try:
            H['suite'](dict(run=lambda label,*args:calls.append((label,env.get('CGO_ENABLED'))),go='go',resources={},env=env))
            self.assertIn(('operand-routing-cpu','0'),calls);self.assertIn(('operand-routing-race','1'),calls)
            self.assertEqual([a['env']['PIR_ROOT_INPUT_LAYOUT']for a in selected[0]['arms']],['peer']+['routed']*4)
        finally:H['VECTORS']['suite']=original


if __name__=='__main__':unittest.main()
