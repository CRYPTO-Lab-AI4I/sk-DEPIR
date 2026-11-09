"""Check Linux process-group accounting and acceptance against retained evidence."""
import copy
import json
import os
from pathlib import Path
import runpy
import subprocess
import sys
import unittest

HERE = Path(__file__).resolve().parent
MODULE = runpy.run_path(str(HERE / 'run.py'))


class RunnerChecks(unittest.TestCase):
    def test_process_group_rss(self):
        child = subprocess.Popen([sys.executable, '-c',
                                  'import time; x=bytearray(8<<20); print("ready",flush=True); time.sleep(10)'],
                                 stdout=subprocess.PIPE, text=True, start_new_session=True)
        try:
            self.assertEqual(child.stdout.readline().strip(), 'ready')
            self.assertGreater(MODULE['group_rss'](child.pid), 8 << 20)
            self.assertEqual(MODULE['group_rss'](-1), 0)
        finally:
            child.terminate()
            child.wait()
            child.stdout.close()

    def test_retained_evidence_and_corruption(self):
        root = HERE.parents[1]
        phases = json.loads((root / 'experiments/parallel-s15/cached/phases.json').read_text())
        pack = next(p for p in phases if p['phase'] == 'pack-output')
        reference = {'canonical_sha256': pack['canonical_sha256'], 'sha256': pack['sha256']}
        MODULE['verify'](phases, reference)
        bad = copy.deepcopy(phases)
        next(p for p in bad if p['phase'] == 'verify-0')['canonical_sha256'] = 'bad'
        with self.assertRaises(AssertionError):
            MODULE['verify'](bad, reference)


if __name__ == '__main__':
    unittest.main()
