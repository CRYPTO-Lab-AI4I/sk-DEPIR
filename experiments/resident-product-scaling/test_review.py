#!/usr/bin/env python3
from pathlib import Path
import runpy
import tempfile
import unittest
HERE=Path(__file__).resolve().parent
H=runpy.run_path(str(HERE/'run.py'))
Fixture=runpy.run_path(str(HERE.parent/'distributed-gpu-product/test_review.py'))['ReviewTest']

class ScalingTests(unittest.TestCase):
    def test_plan_and_selection(self):
        plan=H['specs']();self.assertEqual(len(plan),14);self.assertEqual(len({s['name']for s in plan}),14)
        for tile in (127,4096):
            for backend in ('host','resident'):self.assertEqual([s['sample']for s in plan if not s['memcheck']and s['tile']==tile and s['backend']==backend],[0,1,2])
        self.assertTrue(all(s['dims']==[12]and s['tile']==4096 for s in plan if s['memcheck']))
        old=H['VECTORS']['suite'];seen=[];H['VECTORS']['suite']=lambda c,**kwargs:seen.append(kwargs)
        try:
            H['suite'](dict(run=lambda *a:None,resources={}))
            self.assertEqual(len(seen[0]['arms']),14)
            self.assertTrue(all(a['test']=='TestMatchedGPUProducts'for a in seen[0]['arms']))
            self.assertEqual(sum(a['memcheck']for a in seen[0]['arms']),2)
        finally:H['VECTORS']['suite']=old
    def test_case_subset_and_tile(self):
        # The original small fixture provides a d12 subset with nine127-word
        # messages per4096-point stream. Merge expected endpoints independently
        # to verify the4096-word admission, not only metadata flags.
        with tempfile.TemporaryDirectory()as tmp:
            directory=Path(tmp);rows=Fixture().fixture(directory)
            rows=[r for r in rows if r.get('kind')in ('host','completed')or(r.get('s')==17 and r.get('dimension')==12)or 's17-d12-'in r.get('case','')]
            measurements=[r for r in rows if r.get('kind')=='measurement'];first=min(r['operation_start']for r in measurements);shift=first-1
            for r in rows:
                if r.get('kind')=='completed':r['cases']=2
                if r.get('kind')=='measurement':r.update(operation_start=r['operation_start']-shift,resident=True,right_spectrum_preserved=True,reused=True,tile_words=4096,setup_seconds=.01,work_seconds=.1,validation_capture_seconds=0.,local_seconds=.1,local_work_seconds=.1,local_capture_seconds=0.)
                if r.get('kind')=='taylor-plan':r['operation_start']-=shift
                if r.get('kind')=='taylor-proof':r['packet']=dict(r['packet']);r['packet']['Operation']-=shift
                if r.get('kind')=='pointwise-check':r.update(words=min(2048,r['words']),checked_words=2048,resident=True,alias_rejected=True,layout_rejected=True,invalid_rejected=True,third_preserved=True)
            merged={};rest=[]
            for r in rows:
                if r.get('kind')!='tile-proof':rest.append(r);continue
                r['operation']-=shift
                key=(r['rank'],r['case'],r['pass'],r['operation'],r['direction'])
                if key not in merged:merged[key]=dict(r,offset=0,words=0)
                merged[key]['words']+=r['words']
            rows=rest+list(merged.values())
            for r in measurements:r['proof_events']=sum(q['rank']==r['rank']and q['case']==r['case']for q in merged.values())
            spec=dict(name='fixture',backend='resident',tile=4096,dims=[12],sample=0,memcheck=False)
            self.assertEqual(len(H['review'](rows,directory,spec)['measurements']),4)
            measurements[0]['tile_words']=127
            with self.assertRaises(AssertionError):H['review'](rows,directory,spec)

    def test_named_arm_profile_paths(self):
        suite=H['VECTORS']['suite'];g=suite.__globals__;control=g['CONTROL'];base=g['BASE']
        old_config,old_parse=control['configure'],base['parse_rank_logs']
        control['configure']=lambda c:None;base['parse_rank_logs']=lambda *a,**kw:[]
        try:
            with tempfile.TemporaryDirectory()as tmp:
                root=Path(tmp);source=root/'source';cache=root/'cache';out=root/'out'
                for p in (source/'experiments/distributed-root',cache,out):p.mkdir(parents=True)
                (source/'experiments/distributed-root/rank.sh').write_text('# shared wrapper\n')
                seen=[];env={}
                def run(label,argv,timeout=30):
                    if label=='sample-test':
                        self.assertEqual(env['PIR_PRODUCT_TILE_WORDS'],'4096')
                        self.assertIn('-test.run=^TestMatchedGPUProducts$',argv)
                        for rank in (0,1):(out/f'sample-test-{rank}.pprof').write_bytes(b'fixture')
                    if '-profile-'in label:
                        self.assertEqual(Path(argv[-1]),out/f'sample-test-{len(seen)}.pprof');self.assertTrue(Path(argv[-1]).is_file());seen.append(label)
                context=dict(run=run,go='go',source=source,cache=cache,out=out,env=env,resources={},save=lambda:None)
                suite(context,arms=[dict(name='sample-test',test='TestMatchedGPUProducts',review=lambda *a:dict(matches=True),profile=True,env=dict(PIR_PRODUCT_TILE_WORDS='4096'))])
                self.assertEqual(seen,['sample-test-profile-0','sample-test-profile-1'])
        finally:control['configure']=old_config;base['parse_rank_logs']=old_parse

if __name__=='__main__':unittest.main()
