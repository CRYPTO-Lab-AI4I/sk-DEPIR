#!/usr/bin/env python3
"""Small fixtures verify streaming event selection and repeated-phase timings."""
import hashlib,json,runpy,tempfile,unittest
from unittest.mock import patch
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
 def test_seals_use_bounded_reads(self):
  helper=runpy.run_path(str(HERE.parent/'s17-arithmetic/analyze.py'))['reconcile']
  with tempfile.TemporaryDirectory()as directory:
   out=Path(directory);cache=out/'cache';cache.mkdir()
   data=b'x'*((1<<20)+17);(out/'large.log').write_bytes(data);(cache/'small').write_bytes(b'cache')
   record=dict(status='passed',job_id='test',revision='HEAD',cache=str(cache),source_sha256={},evidence_sha256={'large.log':hashlib.sha256(data).hexdigest()},cache_sha256={'small':hashlib.sha256(b'cache').hexdigest()},commands=[])
   (out/'results.json').write_text(json.dumps(record))
   with patch.object(Path,'read_bytes',side_effect=AssertionError('whole-file read')):
    self.assertEqual(helper(out)['evidence_files_verified'],1)
 def test_repeated_checkpoint_timings(self):
  rows=[dict(phase='recovery-roots',event=event,elapsed_seconds=t)for event,t in [('started',0),('completed',2),('started',10),('completed',13)]]
  self.assertEqual(M['phase_durations'](rows),{'recovery-roots':[2,3]})
if __name__=='__main__':unittest.main()
