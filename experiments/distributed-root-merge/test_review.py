#!/usr/bin/env python3
"""Reject incorrect root ownership/references without weakening FFT route review."""
from pathlib import Path
import copy
import hashlib
import runpy
import tempfile
import unittest

HERE=Path(__file__).resolve().parent
H=runpy.run_path(str(HERE/'run.py'))


class ReviewTest(unittest.TestCase):
    def fixture(self,directory):
        old=runpy.run_path(str(HERE.parent/'distributed-gpu-product/test_review.py'))['ReviewTest']
        rows=old().fixture(directory)
        label='s17-d12-shape-0';n=4096;ac=bc=n//2-1
        rows=[r for r in rows if r.get('kind')in ('host','completed')or r.get('case')==label or (r.get('kind')=='pointwise-check'and r['s']==17 and r['dimension']==12)]
        values=[r for r in rows if r.get('kind')=='measurement'];shift=values[0]['operation_start']-1
        merged={};rest=[]
        full=[max(0,min(k,ac-1)-max(0,k-bc+1)+1)%2 if k<ac+bc-1 else 0 for k in range(n)]
        full_data=b''.join(v.to_bytes(4,'little')for v in full)
        polynomial=b''.join((1).to_bytes(4,'little')for _ in range(ac))
        for r in rows:
            kind=r['kind']
            if kind=='completed':r['cases']=1
            if kind=='pointwise-check':r.update(resident=True,alias_rejected=True,layout_rejected=True,invalid_rejected=True,third_preserved=True,checked_words=n//2)
            if kind=='measurement':
                r.update(a_count=ac,b_count=bc,operation_start=1,oracle_pattern='roots',leading_coefficient=1,local_taylor=True,inverse_reused=True,geometry_reuse_checked=False,resident=True,right_spectrum_preserved=True,reused=True)
                data=full_data[r['rank']*2*n:(r['rank']+1)*2*n]
                for suffix in ('output','expected','reference'):(directory/f"rank-{r['rank']}-{label}-{suffix}.bin").write_bytes(data)
                r['output_sha256']=hashlib.sha256(data).hexdigest()
            if kind=='taylor-plan':r['operation_start']-=shift
            if kind=='taylor-proof':r['packet']=dict(r['packet']);r['packet']['Operation']-=shift
            if kind!='tile-proof':rest.append(r);continue
            r['operation']-=shift;key=(r['rank'],r['pass'],r['operation'],r['direction'])
            if key not in merged:merged[key]=dict(r,offset=0,words=0)
            merged[key]['words']+=r['words']
        rows=rest+list(merged.values())
        for r in values:r['proof_events']=sum(q['rank']==r['rank']for q in merged.values())
        for rank in (0,1):
            own=[]
            for i in range(rank*1023,(rank+1)*1023):
                v=(1<<33)+i+1;own.extend((v,(v*(1<<17))%((1<<34)-1)))
            data=b''.join(v.to_bytes(8,'little')for v in sorted(own))
            for suffix,payload in (('local-root',polynomial),('peer-root',polynomial),('support',data)):
                (directory/f'rank-{rank}-{label}-{suffix}.bin').write_bytes(payload)
            rows.append(dict(kind='root-input',rank=rank,case=label,s=17,dimension=12,support_count=n-4,local_count=n//2-2,a_count=ac,b_count=bc,scalar_witnesses=4,local_preserved=True,root_owned_device_bytes=1,peak_rss_bytes=1,reference_seconds=.1,local_root_seconds=.1,exchange_seconds=.1,input_seconds=.4,local_sha256=hashlib.sha256(polynomial).hexdigest(),peer_sha256=hashlib.sha256(polynomial).hexdigest(),support_sha256=hashlib.sha256(data).hexdigest(),leading_coefficient=1,reference_sha256=hashlib.sha256(full_data[:4*(ac+bc-1)]).hexdigest()))
        return rows

    def test_root_review_and_bad_peer(self):
        with tempfile.TemporaryDirectory()as tmp:
            directory=Path(tmp);rows=self.fixture(directory);spec=dict(reference=False,dim=12)
            self.assertEqual(len(H['review'](rows,directory,spec)['measurements']),2)
            bad=copy.deepcopy(rows);r=next(r for r in bad if r.get('kind')=='root-input');r['peer_sha256']='0'*64
            with self.assertRaises(AssertionError):H['review'](bad,directory,spec)
            # Output/producer expectation/hash agree on a bad coefficient; the
            # separately saved CPU reference must still reject the result.
            r=next(r for r in rows if r.get('kind')=='measurement');prefix=f"rank-{r['rank']}-{r['case']}"
            data=bytearray((directory/(prefix+'-output.bin')).read_bytes());data[40]^=1
            for suffix in ('output','expected'):(directory/(prefix+f'-{suffix}.bin')).write_bytes(data)
            r['output_sha256']=hashlib.sha256(data).hexdigest()
            with self.assertRaises(AssertionError):H['review'](rows,directory,spec)

    def test_missing_source_still_rejected(self):
        with tempfile.TemporaryDirectory()as tmp:
            directory=Path(tmp);rows=self.fixture(directory)
            index=next(i for i,r in enumerate(rows)if r.get('kind')=='taylor-proof')
            with self.assertRaises(AssertionError):H['review'](rows[:index]+rows[index+1:],directory,dict(reference=False,dim=12))

    def test_plan_and_driver(self):
        plan=H['specs']();self.assertEqual(len(plan),5)
        self.assertEqual([(s['dim'],s['memcheck'])for s in plan if not s['reference']],[(12,True),(12,False),(14,False),(16,False)])
        original=H['VECTORS']['suite'];calls=[]
        H['VECTORS']['suite']=lambda c,**kw:calls.append(kw)
        try:
            H['suite'](dict(run=lambda *a:None,resources={}))
            self.assertEqual([a['test']for a in calls[0]['arms']],['TestScaledGPUProducts']+['TestDistributedRootMerge']*4)
            self.assertTrue(calls[0]['full_fft'])
        finally:H['VECTORS']['suite']=original


if __name__=='__main__':unittest.main()
