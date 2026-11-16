#!/usr/bin/env python3
"""Short controller tests; never allocate full resources or start CUDA."""
import copy,json,runpy,subprocess,tempfile,unittest
from pathlib import Path
HERE=Path(__file__).resolve().parent
MODULE=runpy.run_path(str(HERE/'run.py'))
class FullControllerTests(unittest.TestCase):
 def setUp(self):self.plan=json.loads((HERE/'budget.json').read_text())
 def test_plan_and_rejection(self):
  MODULE['validate'](self.plan)
  for group,key in [('host_route','process_group_rss_cap_bytes'),('storage','new_retained_cap_bytes'),('deadlines','slurm_seconds')]:
   altered=copy.deepcopy(self.plan);altered[group][key]+=1
   with self.assertRaises(AssertionError):MODULE['validate'](altered)
  if not self.plan['approved']:
   process=subprocess.run(['python3','-B',str(HERE/'run.py'),'unapproved-test'],capture_output=True,text=True)
   self.assertNotEqual(process.returncode,0);self.assertIn('explicit numeric user budget approval',process.stderr)
 def test_worker_log_tail(self):
  with tempfile.TemporaryDirectory()as directory:
   path=Path(directory)/'worker.log'
   self.assertIsNone(MODULE['last_worker_event'](path))
   path.write_text('x'*70000+'\n'+json.dumps({'phase':'root-product','event':'progress','completed':3,'total':5,'unit':'exponents'})+'\n{"partial":')
   row=MODULE['last_worker_event'](path)
   self.assertEqual(row['completed'],3);self.assertEqual(row['total'],5)
if __name__=='__main__':unittest.main()
