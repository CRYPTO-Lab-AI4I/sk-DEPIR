#!/usr/bin/env python3
from pathlib import Path
import copy
import hashlib
import json
import runpy
import tempfile
import unittest
HERE=Path(__file__).resolve().parent
H=runpy.run_path(str(HERE/'run.py'))


class ReviewTest(unittest.TestCase):
    def test_retained_manifest_bounds_and_copy_corruption(self):
        case=dict(label='test',s=4,count=4,counts=[2]*4)
        names={}
        for rank in range(4):
            names[f'test-rank-{rank}-leaf.bin']=8
            names[f'test-rank-{rank}-final.bin']=16
        for stage in H['TREE']['stages'](case['counts']):
            for rank in range(stage['Ranks']):names[f'test-level-{stage["Level"]}-group-{stage["First"]}-rank-{rank}-output.bin']=4*stage['Domain']//stage['Ranks']
        manifest=dict(cases=[case],files={name:dict(source=f'.cache/liftrs/oracle/{name}',bytes=size,sha256=hashlib.sha256(bytes(size)).hexdigest())for name,size in names.items()})
        self.assertEqual(H['reference_map'](manifest),names)
        for field,value in [('source','../outside'),('bytes',1),('sha256','broken')]:
            bad=copy.deepcopy(manifest);next(iter(bad['files'].values()))[field]=value
            with self.assertRaises((AssertionError,ValueError)):H['reference_map'](bad)
        with tempfile.TemporaryDirectory()as tmp:
            root=Path(tmp);inputs=root/'.cache/liftrs/oracle';inputs.mkdir(parents=True)
            for name,size in names.items():(inputs/name).write_bytes(bytes(size))
            self.assertEqual(H['copy_references'](root,manifest,root/'good'),sum(names.values()))
            name=next(iter(names));data=bytearray((inputs/name).read_bytes());data[0]^=1;(inputs/name).write_bytes(data)
            with self.assertRaises(AssertionError):H['copy_references'](root,manifest,root/'corrupt')

    def test_explicit_runtime_manifest_selection(self):
        policy=runpy.run_path(str(HERE.parent/'gpu-staged-transport/run.py'))
        self.assertTrue(policy['build_input']('experiments/distributed-owned-roots/root-references.json'))
        self.assertFalse(policy['build_input']('experiments/distributed-owned-roots/runs/owned-01/results.json'))
        self.assertFalse(policy['build_input']('experiments/distributed-owned-roots/unreviewed.json'))

    def test_four_arms_preserve_review_callback(self):
        old=H['WIDE']['FOUR']['suite'];captured=[];H['WIDE']['FOUR']['suite']=lambda c,**kw:captured.append(kw)or{}
        owned=dict(name='owned-wide',test='TestOwnedWideRoots',memcheck=False,profile=True)
        try:
            H['WIDE']['suite'](dict(run=lambda *a:None,resources={}),review_fn=H['review'],extra_arms=[owned])
            self.assertIs(captured[0]['review_fn'],H['review'])
            self.assertEqual([a['name']for a in captured[0]['extra_arms']],['wide','owned-wide'])
            self.assertTrue(all(not a['memcheck']and a['profile']for a in captured[0]['extra_arms']))
        finally:H['WIDE']['FOUR']['suite']=old


if __name__=='__main__':unittest.main()
