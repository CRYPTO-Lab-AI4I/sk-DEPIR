#!/usr/bin/env python3
"""Test eight-worker launch policy and independent frozen reference shape."""
from pathlib import Path
import os,runpy,tempfile,unittest
from unittest.mock import patch
H=runpy.run_path(str(Path(__file__).with_name('run.py')))
class ReviewTest(unittest.TestCase):
 def test_actual_suite_budget(self):
  class Captured(Exception):pass
  captured={};original=H['VECTORS']['suite']
  def capture(c,**kw):captured.update(kw,resources=c['resources']);raise Captured()
  H['VECTORS']['suite']=capture
  try:
   with tempfile.TemporaryDirectory()as tmp,patch.dict(os.environ,SLURM_NTASKS='8'):
    c=dict(env={},resources={},run=lambda *a:None,go='go',source=Path(__file__).resolve().parents[2],cache=Path(tmp),out=Path(tmp))
    with self.assertRaises(Captured):H['suite'](c)
   self.assertEqual(captured['ranks'],8);self.assertEqual(captured['cpu_cores'],8)
   self.assertEqual(captured['resources']['host_request_GiB_per_node'],8)
   self.assertEqual([a['name']for a in captured['arms']],['cold','resume'])
  finally:H['VECTORS']['suite']=original
 def test_reference_and_native_worker_extent(self):
  r=H['REFERENCE'];self.assertEqual(sum(r['root_counts']),7381+8);self.assertEqual(len(r['stages']),17)
  for stage in r['stages'].values():self.assertEqual(stage['bytes'],4*stage['prefix'])
  row=dict(kind='workflow-result',case='s8',n=65536)
  plans=[dict(kind='taylor-plan',case='s8/x',rank=rank,degrees=[0,1],level=0,capacity=768)for rank in range(8)]
  result=H['PROOF']['native_expected']([row,*plans],workers=8)
  self.assertEqual(set(result),set(range(8)));self.assertTrue(all(v['windows']>0 for v in result.values()))
if __name__=='__main__':unittest.main()
