#!/usr/bin/env python3
from pathlib import Path
import runpy
import tempfile
import unittest
HERE=Path(__file__).resolve().parent
H=runpy.run_path(str(HERE/'run.py'))
Fixture=runpy.run_path(str(HERE.parent/'distributed-gpu-product/test_review.py'))['ReviewTest']

class LocalTests(unittest.TestCase):
    def test_local_guard_review(self):
        with tempfile.TemporaryDirectory()as tmp:
            directory=Path(tmp);rows=Fixture().fixture(directory)
            # Fixture tiles for N<=4096 merge into one endpoint per stream.
            merged={};rest=[]
            for r in rows:
                if r.get('kind')=='measurement':r.update(resident=True,right_spectrum_preserved=True,reused=True,local_taylor=True)
                if r.get('kind')=='pointwise-check':r.update(words=min(1<<(r['dimension']-1),r['words']),checked_words=1<<(r['dimension']-1),resident=True,alias_rejected=True,layout_rejected=True,invalid_rejected=True,third_preserved=True)
                if r.get('kind')!='tile-proof':rest.append(r);continue
                key=(r['rank'],r['case'],r['pass'],r['operation'],r['direction'])
                if key not in merged:merged[key]=dict(r,offset=0,words=0)
                merged[key]['words']+=r['words']
            rows=rest+list(merged.values())
            for r in rows:
                if r.get('kind')=='measurement':r['proof_events']=sum(q['rank']==r['rank']and q['case']==r['case']for q in merged.values())
            rows.extend(dict(kind='local-window-check',mutation_kinds=['upload','scatter','localcopy'],rank=rank,s=s,dimension=d,canonical_proof_invalidated=True,rejection_preserved=True,both_layouts=True,host_snapshot_preserved=True)for rank in (0,1)for s,d in H['VECTORS']['CASES'])
            spec=dict(kind='local');self.assertEqual(H['review'](rows,directory,spec)['local_guard_checks'],12)
            rows[-1]['canonical_proof_invalidated']=False
            with self.assertRaises(AssertionError):H['review'](rows,directory,spec)
    def test_plan(self):
        plan=H['specs']();self.assertEqual(len(plan),14)
        self.assertEqual([r['kind']for r in plan[:5]],['legacy','fft','fft','local','local'])
        self.assertTrue(plan[2]['memcheck']);self.assertEqual(sum(s['memcheck']for s in plan),2)
        for backend in ('host','resident','resident-local'):self.assertEqual([s['sample']for s in plan if s.get('backend')==backend],[0,1,2])
        old=H['VECTORS']['suite'];seen=[];H['VECTORS']['suite']=lambda c,**kw:seen.append(kw)
        try:
            H['suite'](dict(run=lambda *a:None,resources={}))
            self.assertEqual(seen[0]['arms'][4]['test'],'TestLocalTaylorGPUProducts')
            self.assertTrue(all(a['env'].get('PIR_TAYLOR_LOCAL_GUARD')=='0'for a in seen[0]['arms'][5:]))
        finally:H['VECTORS']['suite']=old

if __name__=='__main__':unittest.main()
