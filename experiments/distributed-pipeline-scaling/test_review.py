#!/usr/bin/env python3
from pathlib import Path
import json
import runpy
import tempfile
import unittest
HERE=Path(__file__).resolve().parent
H=runpy.run_path(str(HERE/'run.py'))


class ReviewTest(unittest.TestCase):
    def test_native_manifest_exact_geometry(self):
        with tempfile.TemporaryDirectory()as tmp:
            root=Path(tmp);d=root/'.cache/liftrs/native';d.mkdir(parents=True)
            case=dict(label='s8',s=8,count=4,counts=[2]*4)
            (d/'s8-roots.json').write_text(json.dumps(case))
            for rank in range(4):
                (d/f's8-rank-{rank}-leaf.bin').write_bytes(bytes(8))
                (d/f's8-rank-{rank}-final.bin').write_bytes(bytes(16))
            for stage in H['OWN']['TREE']['stages'](case['counts']):
                for rank in range(stage['Ranks']):(d/f's8-level-{stage["Level"]}-group-{stage["First"]}-rank-{rank}-output.bin').write_bytes(bytes(4*stage['Domain']//stage['Ranks']))
            manifest=H['reference_manifest'](d,root)
            self.assertEqual(manifest['cases'],[case]);self.assertEqual(len(manifest['files']),16)
            (d/'s8-rank-0-final.bin').write_bytes(bytes(12))
            with self.assertRaises(AssertionError):H['reference_manifest'](d,root)

    def test_shared_review_receives_explicit_cases_and_native_hash(self):
        old=H['OWN']['review'];calls=[]
        H['OWN']['review']=lambda *a,**kw:calls.append(kw)or dict(measurements=[dict(checkpoint_layout='prefix')])
        try:
            with tempfile.TemporaryDirectory()as tmp:
                root=Path(tmp);d=root/'reference-inputs';d.mkdir()
                (d/'s8-reference.json').write_text(json.dumps(dict(sha256={'s8-reference-code.bin':'a'*64})))
                H['review']([],root/'cold')
                self.assertEqual(calls[0],dict(cases=(8,),canonical={8:'a'*64},curve=5))
        finally:H['OWN']['review']=old


if __name__=='__main__':unittest.main()
