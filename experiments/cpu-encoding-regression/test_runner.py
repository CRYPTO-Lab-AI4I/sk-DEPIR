#!/usr/bin/env python3
"""Fast static admission and exact old-source logging-patch checks; no large run."""
import runpy,subprocess,tempfile,unittest
from pathlib import Path
HERE=Path(__file__).resolve().parent;ROOT=HERE.parents[1]
class RegressionRunnerTests(unittest.TestCase):
 def test_budget(self):runpy.run_path(str(HERE/'run.py'))['check']()
 def test_exact_progress_patch(self):
  instrument=runpy.run_path(str(HERE/'baseline_progress.py'))['instrument']
  with tempfile.TemporaryDirectory()as directory:
   baseline=Path(directory);cpp=baseline/'encoding/liftrs-cpp';cpp.mkdir(parents=True)
   original={}
   for name in('matrix.hpp','c_api.cpp'):
    original[name]=subprocess.check_output(['git','show','671bf5e:encoding/liftrs-cpp/'+name],cwd=ROOT,text=True)
    (cpp/name).write_text(original[name])
   instrument(baseline,ROOT)
   matrix=(cpp/'matrix.hpp').read_text();api=(cpp/'c_api.cpp').read_text()
   self.assertEqual(matrix.count('progress.advance();'),2)
   self.assertIn('WorkProgress progress("parallel-blocks",jobs);',matrix)
   self.assertIn('full-inverse',api);self.assertIn('progress.run(label,fn);',api)
   self.assertEqual((cpp/'progress.hpp').read_bytes(),(ROOT/'encoding/liftrs-cpp/progress.hpp').read_bytes())
   # Reject a second application instead of silently duplicating observers.
   with self.assertRaises(AssertionError):instrument(baseline,ROOT)
if __name__=='__main__':unittest.main()
