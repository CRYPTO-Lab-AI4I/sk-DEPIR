#!/usr/bin/env python3
"""Bounded model/parser tests with no GPU allocation or large arrays."""
import json
from pathlib import Path
import runpy
import tempfile
import unittest

HERE = Path(__file__).resolve().parent
M = runpy.run_path(str(HERE/'memory.py'))['model']
R = runpy.run_path(str(HERE/'run.py'))


class AdmissionTests(unittest.TestCase):
    def test_format_and_array_model(self):
        for k in (1, 3876566203, 1 << 35):
            result = M(18, k)
            n = 1 << 36
            self.assertEqual(result['canonical_retained_bytes'], 12*k+12*n+61)
            self.assertEqual(result['discovery_grid_bytes'], 512 << 30)
            self.assertEqual(result['device_product_arrays_bytes'], 1536 << 30)
            self.assertEqual(result['power_of_two_device_count_for_112gib_cap'], 16)
        self.assertEqual(M(17, 3876566203)['canonical_retained_bytes'], 252677224705)
        for k in (0, (1 << 35)+1):
            with self.assertRaises(ValueError):
                M(18, k)

    def test_count_requires_complete_classes(self):
        row = dict(phase='exact-weighted-dimension', event='completed', s=4,
                   t=3, k=37, n=256, completed=81, total=81, owned_device_bytes=8)
        with tempfile.TemporaryDirectory() as directory:
            path = Path(directory)/'log'
            path.write_text(json.dumps(row)+'\n')
            self.assertEqual(R['completed_dimension'](path)['k'], 37)
            row['completed'] -= 1
            path.write_text(json.dumps(row)+'\n')
            with self.assertRaises(AssertionError):
                R['completed_dimension'](path)


if __name__ == '__main__':
    unittest.main()
