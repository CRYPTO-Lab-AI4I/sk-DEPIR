#!/usr/bin/env python3
from pathlib import Path
import hashlib
import json
import runpy
import tempfile
import unittest
HERE=Path(__file__).resolve().parent
H=runpy.run_path(str(HERE/'run.py'))


class ReviewTest(unittest.TestCase):
    def test_compact_four_rank_artifacts_and_storage(self):
        fixture=runpy.run_path(str(HERE.parent/'distributed-preprocessing-pilot/test_review.py'))['ReviewTest']
        with tempfile.TemporaryDirectory()as tmp:
            d=Path(tmp);s,k=fixture().fixture(d,ranks=4)
            for path in d.glob('s4-*-commit.json'):
                record=json.loads(path.read_text());record['layout']='prefix'
                for rank in range(4):
                    datafile=d/f's4-rank-{rank}-{record["stage"]}.bin';raw=datafile.read_bytes();first=rank*64;stored=min(64,max(0,record['prefix']-first));raw=raw[:4*stored]
                    datafile.write_bytes(raw);record['sha256'][rank]=hashlib.sha256(raw).hexdigest()
                path.write_text(json.dumps(record))
            result=H['OWN']['PILOT']['artifacts'](d,s,k,ranks=4)
            self.assertEqual(result['checkpoint_prefixes'][-1],248)
            # Generic storage summary also expects s6;test exact interval decoding
            # and savings directly here without fabricating another code fixture.
            root=json.loads((d/'s4-root-commit.json').read_text());self.assertEqual(sum((d/f's4-rank-{r}-root.bin').stat().st_size for r in range(4)),4*(k+1))
            self.assertEqual(len(H['OWN']['PILOT']['checkpoint_bytes'](d,'s4','root',root)),1024)
            empty=d/'s4-rank-3-root.bin';empty.write_bytes(bytes(4))
            with self.assertRaises(AssertionError):H['OWN']['PILOT']['artifacts'](d,s,k,ranks=4)

    def test_explicit_compact_selection_preserves_owned_suite(self):
        old=H['OWN']['suite'];captured=[]
        H['OWN']['suite']=lambda c:captured.append(dict(c['env']))or{}
        globals_=H['suite'].__globals__;original=globals_['storage'];globals_['storage']=lambda p:dict(checked=True)
        try:
            result=H['suite'](dict(run=lambda *a:None,env={},cache=Path('/cache'),resources={}))
            self.assertEqual(captured,[dict(PIR_PIPELINE_CHECKPOINT_LAYOUT='prefix')])
            self.assertEqual(set(result['checkpoint_storage']),{'cold','resume'})
        finally:H['OWN']['suite']=old;globals_['storage']=original


if __name__=='__main__':unittest.main()
