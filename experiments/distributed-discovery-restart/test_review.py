#!/usr/bin/env python3
"""Real larger-case recovery wiring, summary-footer rejection and reference bounds."""
from pathlib import Path
import copy,os,runpy,tempfile,unittest
from unittest.mock import patch
H=runpy.run_path(str(Path(__file__).with_name('run.py')))
class ReviewTest(unittest.TestCase):
 def test_actual_suite(self):
  class Captured(Exception):pass
  captured={};original=H['VECTORS']['suite']
  def capture(c,**kw):captured.update(kw);raise Captured()
  H['VECTORS']['suite']=capture
  try:
   with tempfile.TemporaryDirectory()as tmp,patch.dict(os.environ,SLURM_NTASKS='8'):
    c=dict(env={},resources={},run=lambda *a:None,go='go',source=Path(__file__).resolve().parents[2],cache=Path(tmp),out=Path(tmp))
    with self.assertRaises(Captured):H['suite'](c)
    self.assertEqual([a['name']for a in captured['arms']],['cold','discovery-resume','root-resume','prefix-resume'])
    import json
    cfg=json.loads((Path(tmp)/'cold.json').read_text());self.assertEqual(cfg['spec']['FieldDegree'],10);self.assertEqual(cfg['proof_log'],'summary');self.assertEqual(cfg['max_local_words'],131072)
    self.assertEqual(cfg['device_bytes'],32<<20);self.assertEqual(cfg['rss_bytes'],1<<30)
  finally:H['VECTORS']['suite']=original
 def test_summary_checks(self):
  rows=[dict(kind='proof-summary',rank=rank,proof_kind=kind,complete=True,mode='summary',codec='go-json-v1-length64le',records=1,semantic_bytes=100,sha256='01'*32)for rank in range(8)for kind in ('tile-proof','taylor-proof','operand-proof')]
  self.assertEqual(H['HOST']['recording_review'](rows,'summary')['records'],24)
  bad=copy.deepcopy(rows);bad[0]['complete']=False
  with self.assertRaises(AssertionError):H['HOST']['recording_review'](bad,'summary')
  with self.assertRaises(AssertionError):H['HOST']['recording_review'](rows+[dict(kind='tile-proof')],'summary')
 def test_reference_capacity(self):
  r=H['REFERENCE'];h=r['header'];self.assertEqual((h['K'],h['N']),(146641,1048576));self.assertEqual(len(r['stages']),21)
  self.assertLessEqual(max(r['root_counts']),h['N']//8);self.assertLess(max(v['bytes']for v in r['files'].values()),32<<20)
  total=sum(v['bytes']for v in r['files'].values())+r['files']['s10-archive.raw']['bytes']+sum(v['bytes']for v in r['stages'].values())+23*65536
  self.assertLess(total,32<<20)
if __name__=='__main__':unittest.main()
