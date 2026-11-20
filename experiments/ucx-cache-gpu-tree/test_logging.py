#!/usr/bin/env python3
from pathlib import Path
import json
import os
import runpy
import subprocess
import sys
import tempfile
import unittest
HERE=Path(__file__).resolve().parent
RECOVER=runpy.run_path(str(HERE/'review_interleaved.py'))
BASE=runpy.run_path(str(HERE.parent/'gpu-staged-transport/run.py'))


class Logging(unittest.TestCase):
    def broken(self):
        return ('0: {"phase":"local-vector-transpose","event":"completed","completed":65536,"t'
                '[node:123] 1 more process has sent help message help-mpi-btl-openib.txt / no device params found\n'
                '0: otal":65536,"unit":"elements"}\n')
    def test_explicit_recovery_preserves_warning_and_all_event_fields(self):
        fixed,manifest=RECOVER['recover'](self.broken())
        self.assertEqual(json.loads(fixed.splitlines()[0][3:])['total'],65536)
        self.assertIn(manifest[0]['warning'],fixed)
        self.assertEqual(len(manifest),1)
    def test_ambiguous_or_unrelated_interleave_is_rejected(self):
        for text in (self.broken().replace('0: otal','1: otal'),self.broken().splitlines()[0],self.broken().replace('local-vector-transpose','other-phase'),self.broken().replace('no device params found','unrelated warning')):
            with self.assertRaises((AssertionError,json.JSONDecodeError)):RECOVER['recover'](text)
    def test_rank_streams_are_complete_and_strict(self):
        with tempfile.TemporaryDirectory()as tmp:
            path=Path(tmp)/'normal.log';path.write_text('launcher diagnostic {not-json}\n')
            files=[]
            for rank in range(4):
                for stream in ('stdout','stderr'):
                    file=Path(tmp)/f'normal-{rank}.{stream}.log';file.write_text(json.dumps(dict(rank=rank,stream=stream))+'\n');files.append(file)
            self.assertEqual(len(BASE['parse_rank_logs'](path)),8)
            files[-1].write_text('{"kind":')
            with self.assertRaises(json.JSONDecodeError):BASE['parse_rank_logs'](path)
            files[-1].unlink()
            with self.assertRaises(AssertionError):BASE['parse_rank_logs'](path)
    def test_wrapper_isolates_worker_streams_and_preserves_legacy_mode(self):
        wrapper=HERE.parent/'distributed-root/rank.sh'
        with tempfile.TemporaryDirectory()as tmp:
            prefix=Path(tmp)/'normal';env=dict(os.environ,SLURM_PROCID='0',PIR_RANK_LOG_FILES='1')
            command=['bash',str(wrapper),str(prefix),sys.executable,'-c','import sys;print("stdout event");print("stderr diagnostic",file=sys.stderr)']
            result=subprocess.run(command,env=env,capture_output=True,text=True,check=True)
            self.assertEqual((result.stdout,result.stderr),('',''))
            self.assertEqual(Path(str(prefix)+'-0.stdout.log').read_text(),'stdout event\n')
            self.assertEqual(Path(str(prefix)+'-0.stderr.log').read_text(),'stderr diagnostic\n')
            env['PIR_RANK_LOG_FILES']='0';result=subprocess.run(command,env=env,capture_output=True,text=True,check=True)
            self.assertEqual((result.stdout,result.stderr),('stdout event\n','stderr diagnostic\n'))


if __name__=='__main__':unittest.main()
