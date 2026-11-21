#!/usr/bin/env python3
from pathlib import Path
import runpy
import unittest
import tempfile
import hashlib
H=runpy.run_path(str(Path(__file__).with_name('run.py')))

class LargerTests(unittest.TestCase):
    def test_prefix_oracle_against_schoolbook(self):
        for ac in (1,3,8,17):
            for bc in (1,2,5,17):
                n=2
                while n<ac+bc+3:n*=2
                a=[1 if i in (0,ac-1)else(((i+1)*0x9e3779b9)^0x1337)&((1<<17)-1)for i in range(ac)]
                out=[0]*n
                for i,v in enumerate(a):
                    for j in range(bc):out[i+j]^=v
                for rank in (0,1):self.assertEqual(H['prefix_oracle'](n,ac,bc,rank),b''.join(v.to_bytes(4,'little')for v in out[rank*n//2:(rank+1)*n//2]))
    def test_plan_and_selection(self):
        plan=H['specs']();self.assertEqual(len(plan),22);self.assertEqual(len({s['name']for s in plan}),22)
        self.assertEqual(sum(s['memcheck']for s in plan),2)
        for dim in (14,16,18):
            for backend in ('host','resident-local'):
                selected=[s for s in plan if s['dim']==dim and s['sample']is not None and s['backend']==backend]
                self.assertEqual([s['sample']for s in selected],[0,1,2]);self.assertTrue(all(s['shapes']==[0]and s['pattern']=='ones'for s in selected))
        old=H['VECTORS']['suite'];seen=[];H['VECTORS']['suite']=lambda c,**kw:seen.append(kw)
        try:
            H['suite'](dict(run=lambda *a:None,resources={}))
            self.assertEqual(len(seen[0]['arms']),22)
            self.assertTrue(all(a['test']=='TestScaledGPUProducts'for a in seen[0]['arms']))
            self.assertTrue(all(a['env']['PIR_SCALED_SHAPES']=='1'for a in seen[0]['arms'][4:]))
        finally:H['VECTORS']['suite']=old

    def test_subset_output_and_inverse_reuse(self):
        fixture=runpy.run_path(str(Path(__file__).resolve().parent.parent/'distributed-gpu-product/test_review.py'))['ReviewTest']
        with tempfile.TemporaryDirectory()as tmp:
            directory=Path(tmp);rows=fixture().fixture(directory);label='s17-d12-shape-0'
            rows=[r for r in rows if r.get('kind')in ('host','completed')or r.get('case')==label or(r.get('kind')=='pointwise-check'and r['s']==17 and r['dimension']==12)]
            measured=[r for r in rows if r.get('kind')=='measurement'];shift=measured[0]['operation_start']-1
            merged={};rest=[]
            for r in rows:
                if r.get('kind')=='completed':r['cases']=1
                if r.get('kind')=='measurement':
                    r.update(operation_start=r['operation_start']-shift,oracle_pattern='ones',resident=False,local_taylor=False,tile_words=4096,setup_seconds=.01,inverse_reused=True,geometry_reuse_checked=False,work_seconds=1.,validation_capture_seconds=0.,local_seconds=1.,local_work_seconds=1.,local_capture_seconds=0.)
                    data=H['prefix_oracle'](4096,r['a_count'],r['b_count'],r['rank'])
                    for suffix in ('output','expected'):(directory/f"rank-{r['rank']}-{label}-{suffix}.bin").write_bytes(data)
                    r['output_sha256']=hashlib.sha256(data).hexdigest()
                if r.get('kind')=='taylor-plan':r['operation_start']-=shift
                if r.get('kind')=='taylor-proof':r['packet']=dict(r['packet']);r['packet']['Operation']-=shift
                if r.get('kind')!='tile-proof':rest.append(r);continue
                r['operation']-=shift;key=(r['rank'],r['pass'],r['operation'],r['direction'])
                if key not in merged:merged[key]=dict(r,offset=0,words=0)
                merged[key]['words']+=r['words']
            rows=rest+list(merged.values())
            for r in measured:r['proof_events']=sum(q['rank']==r['rank']for q in merged.values())
            spec=dict(backend='host',dim=12,pattern='ones',shapes=[0])
            self.assertEqual(len(H['review'](rows,directory,spec)['measurements']),2)
            measured[0]['inverse_reused']=False
            with self.assertRaises(AssertionError):H['review'](rows,directory,spec)
            measured[0]['inverse_reused']=True
            row=measured[0];path=directory/f"rank-{row['rank']}-{label}-output.bin";bad=bytearray(path.read_bytes());bad[40]^=1
            for suffix in ('output','expected'):(directory/f"rank-{row['rank']}-{label}-{suffix}.bin").write_bytes(bad)
            row['output_sha256']=hashlib.sha256(bad).hexdigest()
            with self.assertRaises(AssertionError):H['review'](rows,directory,spec)
    def test_proof_file_budget(self):
        fft=H['PRODUCT']['FFT'];dim=18;n=1<<dim;cols=1<<(dim//2);capacity=3*cols;events=0
        for level in range(dim//2):
            degrees=[(1<<k)<<level for k in range(dim//2)]
            for quotient in (True,False):events+=len(fft['expected_windows'](n,cols,degrees,level,quotient,capacity))
        events*=3
        self.assertEqual(events,20604);self.assertLess(events*800+1*1024**2+192*800,24*1024**2)

if __name__=='__main__':unittest.main()
