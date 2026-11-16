#!/usr/bin/env python3
"""Small fixtures verify streaming event selection and repeated-phase timings."""
import json,runpy,tempfile,unittest
from pathlib import Path
HERE=Path(__file__).resolve().parent
M=runpy.run_path(str(HERE/'analyze.py'))
class ReviewTests(unittest.TestCase):
 def test_large_noise_and_selected_events(self):
  with tempfile.TemporaryDirectory()as directory:
   path=Path(directory)/'log'
   with path.open('w')as stream:
    for _ in range(1000):stream.write('{"phase":"GPU-pointwise","event":"completed","completed":65536}\n')
    for row in [dict(phase='identity',event='started',elapsed_seconds=10),dict(phase='identity',event='completed',elapsed_seconds=15),dict(phase='workflow',event='completed',elapsed_seconds=20)]:stream.write(json.dumps(row)+'\n')
   rows=M['review_events'](path);self.assertEqual(len(rows),3)
   self.assertEqual(M['phase_durations'](rows),{'identity':[5]})
 def test_repeated_checkpoint_timings(self):
  rows=[dict(phase='recovery-roots',event=event,elapsed_seconds=t)for event,t in [('started',0),('completed',2),('started',10),('completed',13)]]
  self.assertEqual(M['phase_durations'](rows),{'recovery-roots':[2,3]})
if __name__=='__main__':unittest.main()
