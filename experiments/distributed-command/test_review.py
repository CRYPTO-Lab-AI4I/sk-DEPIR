#!/usr/bin/env python3
"""Exercise real command-arm wiring and per-rank profile flag separation."""
from pathlib import Path
import json
import runpy
import os
import subprocess
import tempfile
import unittest
H=runpy.run_path(str(Path(__file__).with_name('run.py')))
class ReviewTest(unittest.TestCase):
    def test_real_rank_wrapper(self):
        with tempfile.TemporaryDirectory()as tmp:
            directory=Path(tmp);program=directory/'fake command.py';prefix=directory/'rank output'
            program.write_text('import json,pathlib,sys\nassert sys.argv[-1].startswith("--profile=")\nassert not any(a.startswith("-test.")for a in sys.argv)\npathlib.Path(sys.argv[-1].split("=",1)[1]).write_text("profile")\nprint(json.dumps(sys.argv[1:]))\n')
            wrapper=Path(__file__).resolve().parents[2]/'cmd/pir-preprocess-mpi/rank.sh'
            result=subprocess.run(['bash',str(wrapper),str(prefix),'python3',str(program),'--config','path with spaces'],env=dict(os.environ,SLURM_PROCID='3'),capture_output=True,text=True,timeout=10)
            self.assertEqual(result.returncode,0,result.stderr)
            self.assertEqual(json.loads(Path(str(prefix)+'-3.stdout.log').read_text())[:-1],['--config','path with spaces'])
            self.assertEqual(Path(str(prefix)+'-3.pprof').read_text(),'profile')
    def test_actual_command_suite(self):
        class Captured(Exception):pass
        original=H['SCALE']['suite'];captured={}
        def capture(c,**kw):captured.update(kw);raise Captured()
        H['SCALE']['suite']=capture
        try:
            with tempfile.TemporaryDirectory()as tmp:
                root=Path(__file__).resolve().parents[2];cache=Path(tmp)
                with self.assertRaises(Captured):H['suite'](dict(env={},run=lambda *args:None,go='go',cache=cache,out=cache,source=root))
                arms=captured['extra_arms'];self.assertEqual([a['name']for a in arms],['command-cold','command-resume'])
                for a in arms:
                    self.assertEqual(a['command'][0],cache/'pir-preprocess-mpi');self.assertNotIn('-test.run',str(a['command']))
                    self.assertEqual(a['wrapper'],cache/'preprocess-rank.sh')
                config=json.loads((cache/'command-cold.json').read_text())
                self.assertEqual(config['ranks'],4);self.assertEqual(config['device_bytes'],64<<20);self.assertEqual(config['proof_log'],'trace')
                wrapper=(cache/'preprocess-rank.sh').read_text();self.assertIn('--profile=',wrapper);self.assertNotIn('-test.cpuprofile',wrapper)
        finally:H['SCALE']['suite']=original
if __name__=='__main__':unittest.main()
