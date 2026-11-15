#!/usr/bin/env python3
"""Fast tests of graceful shutdown and the unkillable-driver wait boundary."""
import os
from pathlib import Path
import runpy
import signal
import subprocess
import unittest
from unittest.mock import Mock,patch
STOP=runpy.run_path(str(Path(__file__).with_name('run.py')))['stop_process']
class ShutdownTests(unittest.TestCase):
 def test_owned_process(self):
  process=subprocess.Popen(['sleep','10'],start_new_session=True)
  try:
   result=STOP(process,wait_seconds=1)
   self.assertTrue(result['reaped']);self.assertIsNotNone(process.poll())
  finally:
   if process.poll()is None:os.killpg(process.pid,signal.SIGKILL);process.wait(timeout=2)
 def test_unreaped_driver(self):
  process=Mock();process.pid=1234;process.poll.return_value=None
  process.wait.side_effect=subprocess.TimeoutExpired('simulated-driver',.01)
  with patch('os.killpg')as kill:
   result=STOP(process,wait_seconds=.01)
   self.assertFalse(result['reaped']);self.assertEqual(process.wait.call_count,2)
   self.assertEqual([call.args[1]for call in kill.call_args_list],[signal.SIGTERM,signal.SIGKILL])
   self.assertTrue(all(call.kwargs['timeout']==.01 for call in process.wait.call_args_list))
if __name__=='__main__':unittest.main()
