#!/usr/bin/env python3
from pathlib import Path
import hashlib
import runpy
import tempfile
import unittest
H=runpy.run_path(str(Path(__file__).with_name('run.py')))


class ReviewTest(unittest.TestCase):
    def test_runtime_manifest_is_a_snapshot_input(self):
        selector=H['VECTORS']['BASE']['build_input']
        self.assertTrue(selector('experiments/distributed-coordinator-stream/reference-inputs.json'))
        for name in ('experiments/distributed-coordinator-stream/readiness/checks.json',
                     'experiments/distributed-coordinator-stream/runs/example/results.json',
                     'experiments/other/reference-inputs.json'):
            self.assertFalse(selector(name))

    def test_reference_copy_and_corruption(self):
        with tempfile.TemporaryDirectory()as tmp:
            root=Path(tmp);source=root/'source';source.mkdir();manifest=dict(source='source',files={})
            for s in (4,6):
                for suffix in ('reference.json','support.bin','reference-code.bin','rank-0-reference-p.bin','rank-0-reference-g.bin','rank-0-reference-spectrum.bin'):
                    name=f's{s}-{suffix}';(source/name).write_bytes(b'abc');manifest['files'][name]=dict(bytes=3,sha256=hashlib.sha256(b'abc').hexdigest())
            self.assertEqual(H['copy_inputs'](root,manifest,root/'ok'),36)
            (source/'s4-support.bin').write_bytes(b'bad')
            with self.assertRaises(AssertionError):H['copy_inputs'](root,manifest,root/'corrupt')
            manifest['source']='../escape'
            with self.assertRaises(AssertionError):H['copy_inputs'](root,manifest,root/'escape')

    def test_race_environment_before_job_work(self):
        class Captured(Exception):pass
        old=H['VECTORS']['BASE']['ROOT']
        try:
            with tempfile.TemporaryDirectory()as tmp:
                root=Path(tmp);source=root/'source';source.mkdir();cache=root/'cache';cache.mkdir()
                metadata=root/'experiments/distributed-coordinator-stream';metadata.mkdir(parents=True)
                manifest=dict(source='source',files={})
                for size in (4,6):
                    for suffix in ('reference.json','support.bin','reference-code.bin','rank-0-reference-p.bin','rank-0-reference-g.bin','rank-0-reference-spectrum.bin'):
                        name=f's{size}-{suffix}';(source/name).write_bytes(b'abc');manifest['files'][name]=dict(bytes=3,sha256=hashlib.sha256(b'abc').hexdigest())
                import json
                (metadata/'reference-inputs.json').write_text(json.dumps(manifest))
                H['VECTORS']['BASE']['ROOT']=root;env=dict(CGO_ENABLED='0');events=[]
                def run(label,argv,timeout):
                    events.append(label)
                    if label=='bounded-reference-race':
                        self.assertEqual(env['CGO_ENABLED'],'1');self.assertEqual(env['CC'],'mpicc');self.assertIn('reference-preflight',events);raise Captured()
                with self.assertRaises(Captured):H['suite'](dict(run=run,source=root,cache=cache,env=env,resources={},save=lambda:None,go='go'))
        finally:H['VECTORS']['BASE']['ROOT']=old

    def test_shared_restart_suite_selects_retained_mode(self):
        restart=H['RESTART'];old=restart['VECTORS']['suite'];selected=[]
        restart['VECTORS']['suite']=lambda c,**kw:selected.append(kw)
        try:
            restart['suite'](dict(run=lambda *a:None,go='go',env={},cache=Path('/cache'),out=Path('/out')),reference_from=Path('/reference'),review_fn=lambda *a:None)
            arms=selected[0]['arms'];self.assertEqual([a['name']for a in arms],['cold','resume'])
            self.assertTrue(all(a['env']['PIR_PIPELINE_REFERENCE_FROM']=='/reference'and a['memcheck']is False for a in arms))
        finally:restart['VECTORS']['suite']=old


if __name__=='__main__':unittest.main()
