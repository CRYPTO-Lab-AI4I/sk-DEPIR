#!/usr/bin/env python3
"""Check complete FFT pass/window coverage, including inactive source prefixes."""
from pathlib import Path
import runpy
import tempfile
import unittest
HERE=Path(__file__).resolve().parent
H=runpy.run_path(str(HERE/'run.py'))
OLD=runpy.run_path(str(HERE.parent/'distributed-gpu-vectors/test_review.py'))


class ReviewTest(unittest.TestCase):
    def fixture(self,d):
        rows=OLD['ReviewTest']().fixture(d);remap={};starts={};operation=1;extra=[];counts={}
        for index,(s,dim)in enumerate(H['VECTORS']['CASES']):
            label=f's{s}-d{dim}';n=1<<dim;cols=1<<(dim//2);levels=dim-dim//2;capacity=min(n,3*max(cols,n//cols))
            for p in (0,1):
                starts[(label,p)]=operation;matrix=operation+(2*levels if p==0 else 0)
                for j in (0,1):remap[4*index+2*p+1+j]=matrix+j
                counts[(label,p,0)]=counts[(label,p,1)]=0
                for step in range(levels):
                    level=levels-1-step if p==0 else step;degrees=[1<<level]
                    for rank in (0,1):extra.append(dict(kind='taylor-plan',rank=rank,case=label,level=level,degrees=degrees,capacity=capacity,operation_start=operation,**{'pass':p}))
                    for stage in (0,1):
                        q=(stage==0)!=(p==1);op=operation+(2 if p==1 else 0)+2*step+stage
                        for target,source,block,first,count,slot,a,words in H['expected_windows'](n,cols,degrees,level,q,capacity):
                            packet=dict(Operation=op,Domain=n,Columns=cols,Level=level,Quotient=q,Target=dict(Block=block,First=first,Count=count),Slot=slot,Source=a,Words=words)
                            for rank,peer,direction in ((source,target,'send'),(target,source,'receive')):
                                extra.append(dict(kind='taylor-proof',rank=rank,case=label,peer=peer,direction=direction,packet=packet,own_proof='00'*64,remote_proof='00'*64,**{'pass':p}));counts[(label,p,rank)]+=1
                operation+=2+2*levels
        for r in rows:
            if r.get('kind')=='tile-proof':r['operation']=remap[r['operation']]
            if r.get('kind')=='measurement':r.update(full_fft=True,operation_start=starts[(r['case'],r['pass'])],taylor_proof_events=counts[(r['case'],r['pass'],r['rank'])])
        return rows+extra

    def test_complete_and_missing_window(self):
        with tempfile.TemporaryDirectory()as tmp:
            d=Path(tmp);rows=self.fixture(d);self.assertEqual(len(H['review'](rows,d)['measurements']),24)
            i=next(i for i,r in enumerate(rows)if r.get('kind')=='taylor-proof')
            with self.assertRaises(AssertionError):H['review'](rows[:i]+rows[i+1:],d)

    def test_wrong_source(self):
        with tempfile.TemporaryDirectory()as tmp:
            d=Path(tmp);rows=self.fixture(d);r=next(r for r in rows if r.get('kind')=='taylor-proof');r['packet']=dict(r['packet']);r['packet']['Source']+=1
            with self.assertRaises(AssertionError):H['review'](rows,d)

    def test_full_runner_baseline_preflight(self):
        class Captured(Exception):pass
        with tempfile.TemporaryDirectory()as tmp:
            root=Path(tmp);source=root/'local-source';cache=root/'shared';out=root/'logs'
            for p in (source/'experiments/distributed-root',cache,out):p.mkdir(parents=True)
            (source/'experiments/distributed-root/rank.sh').write_text('# frozen wrapper\n')
            (out/'ucx-config.log').write_text('UCX_RCACHE_ENABLE=n\n');env={};preflight=[]
            def run(label,argv,timeout=30):
                if label=='rank-preflight':preflight.append(label)
                if label=='separable':
                    self.assertEqual(preflight,['rank-preflight']);self.assertEqual(env['PIR_GPU_MATRIX_STAGE_TAYLOR'],'0')
                    self.assertIn('--chdir='+str(cache),argv);self.assertNotIn(str(source),' '.join(map(str,argv)))
                    self.assertEqual(Path(argv[argv.index('bash')+1]),cache/'rank.sh');raise Captured()
            with self.assertRaises(Captured):H['suite'](dict(run=run,go='go',source=source,cache=cache,out=out,env=env,resources={},save=lambda:None))

    def test_operative_source_intervals(self):
        got=H['expected_windows'](16,4,[2,4],1,False,12)
        self.assertEqual(got,[(0,1,0,2,4,0,8,4),(0,1,0,2,4,1,8,2),(0,1,0,6,2,0,12,2),(0,1,0,6,2,1,10,2)])
        self.assertEqual(H['expected_windows'](16,4,[2,4],1,True,12),[])


if __name__=='__main__':unittest.main()
