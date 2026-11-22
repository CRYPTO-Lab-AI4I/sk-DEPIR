#!/usr/bin/env python3
import runpy
from pathlib import Path
import unittest
H=runpy.run_path(str(Path(__file__).with_name('audit.py')))


class AuditTest(unittest.TestCase):
    def test_large_arrays_and_checkpoint_units(self):
        r=H['project'](18,k=16300262095)
        self.assertEqual(r['device_three_arrays_bytes_per_gpu'],96<<30)
        self.assertEqual(r['host_base_array_bytes_per_worker'],16<<30)
        self.assertEqual(r['host_extension_array_bytes_per_worker'],32<<30)
        self.assertEqual(r['newton_root_clone_bytes_per_step_per_worker'],16<<30)
        self.assertEqual(r['checkpoint_prefix_payload_bytes'],(512<<30)-4)
        self.assertEqual(r['newton_dense_input_upload_bytes_total'],36<<40)
        self.assertLess(r['newton_prefix_input_upload_bytes_total'],2<<40)
        self.assertEqual(r['payload_only_min_power2_gpus_at112GiB'],16)
        self.assertEqual(r['main_transpose_proof_rows_per_worker'],440401920)

    def test_exploratory_limits_remain_visible(self):
        for s,minimum in ((19,64),(20,256)):
            r=H['project'](s)
            self.assertEqual(r['payload_only_min_power2_gpus_at112GiB'],minimum)
            self.assertFalse(r['backend_degree_admitted']);self.assertFalse(r['full_distributed_pipeline_run'])
            self.assertGreater(r['taylor_capacity_words'],r['taylor_current_capacity_limit_words'])
            self.assertNotIn('newton_steps',r)


if __name__=='__main__':unittest.main()
