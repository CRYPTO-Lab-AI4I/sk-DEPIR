"""Check the resource policy that previously discarded a charged-battery run."""
import importlib.util
from pathlib import Path
import unittest
spec=importlib.util.spec_from_file_location('checkpoint_run',Path(__file__).with_name('run.py'))
runner=importlib.util.module_from_spec(spec);spec.loader.exec_module(runner)
class PowerPolicy(unittest.TestCase):
 def test_charged_battery_continues(self):
  self.assertIsNone(runner.power_action("Now drawing from 'Battery Power'\n 40%; discharging",False))
 def test_low_battery_preserves_work(self):
  self.assertEqual(runner.power_action("Now drawing from 'Battery Power'\n 10%; discharging",False),'pause')
  self.assertIsNone(runner.power_action("Now drawing from 'Battery Power'\n 9%; discharging",True))
 def test_ac_resumes(self):
  self.assertEqual(runner.power_action("Now drawing from 'AC Power'\n 8%; charging",True),'resume')
  self.assertIsNone(runner.power_action("Now drawing from 'AC Power'\n 8%; charging",False))
 def test_unknown_reading_does_not_kill_work(self):
  self.assertIsNone(runner.power_action('',False))
if __name__=='__main__':unittest.main()
