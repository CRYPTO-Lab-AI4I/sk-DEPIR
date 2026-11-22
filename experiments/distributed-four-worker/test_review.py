#!/usr/bin/env python3
from pathlib import Path
import runpy
import tempfile
import unittest
HERE=Path(__file__).resolve().parent
H=runpy.run_path(str(HERE/'run.py'))


class ReviewTest(unittest.TestCase):
    def test_four_worker_fft_peer_coverage(self):
        fixture=runpy.run_path(str(HERE.parent/'distributed-gpu-fft/test_review.py'))['ReviewTest']
        for small in (True,False):
            with tempfile.TemporaryDirectory()as tmp:
                d=Path(tmp);cases=[(18,v)for v in ((5,)if small else(5,12))]
                rows=fixture().fixture(d,cases,ranks=4)
                for r in rows:
                    if r.get('kind')=='measurement':r.update(resident=True,third_preserved=True,reused=True,range_rejection_preserved=True,local_taylor=True)
                spec=dict(kind='fft',small=small)
                self.assertEqual(len(H['review'](rows,d,spec)['measurements']),8*len(cases))
                i=next(i for i,r in enumerate(rows)if r.get('kind')=='tile-proof'and r['rank']==0 and r['peer']==2)
                with self.assertRaises(AssertionError):H['review'](rows[:i]+rows[i+1:],d,spec)
                rows[i]['round']=1
                with self.assertRaises(AssertionError):H['review'](rows,d,spec)

    def test_four_worker_product_and_missing_window(self):
        fixture=runpy.run_path(str(HERE.parent/'distributed-gpu-product/test_review.py'))['ReviewTest']
        with tempfile.TemporaryDirectory()as tmp:
            d=Path(tmp);rows=fixture().fixture(d,cases=[(18,5)],ranks=4)
            for r in rows:
                if r.get('kind')=='measurement':r.update(resident=True,right_spectrum_preserved=True,reused=True,local_taylor=True)
                if r.get('kind')=='pointwise-check':r.update(words=8,checked_words=8,resident=True,alias_rejected=True,layout_rejected=True,invalid_rejected=True,third_preserved=True)
            spec=dict(kind='product',small=True)
            self.assertEqual(len(H['review'](rows,d,spec)['measurements']),8)
            i=next(i for i,r in enumerate(rows)if r.get('kind')=='taylor-proof')
            with self.assertRaises(AssertionError):H['review'](rows[:i]+rows[i+1:],d,spec)

    def test_shared_gpu_launch_admission(self):
        class Captured(Exception):pass
        with tempfile.TemporaryDirectory()as tmp:
            root=Path(tmp);source=root/'source';cache=root/'cache';out=root/'logs'
            for p in (source/'experiments/distributed-root',cache,out):p.mkdir(parents=True)
            (source/'experiments/distributed-root/rank.sh').write_text('# shared wrapper\n');(out/'ucx-config.log').write_text('UCX_RCACHE_ENABLE=n\n');resources={}
            def run(label,argv,timeout=30):
                if label=='normal':
                    self.assertIn('--ntasks=4',argv);self.assertIn('--ntasks-per-node=2',argv);self.assertIn('--cpus-per-task=1',argv)
                    self.assertIn('--gpu-bind=single:2',argv);self.assertIn('--gres-flags=allow-task-sharing',argv)
                    self.assertNotIn('--gpus-per-task=h200:1',argv);self.assertEqual(resources['owned_cap_MiB_per_worker'],64);raise Captured()
            with self.assertRaises(Captured):H['VECTORS']['suite'](dict(run=run,go='go',source=source,cache=cache,out=out,env={},resources=resources,save=lambda:None),ranks=4)


if __name__=='__main__':unittest.main()
