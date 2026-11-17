#!/usr/bin/env python3
"""Check truncation/backing lifetimes and the distinction from measured latency."""
import runpy
from pathlib import Path
import unittest
M = runpy.run_path(str(Path(__file__).with_name('model.py')))


class WorkflowModel(unittest.TestCase):
    def test_prefix_requires_full_product_padding(self):
        trace = M['newton_trace'](64, 15)
        last = trace[-1]
        self.assertEqual((last['old'], last['square_backing'], last['output']), (32, 63, 49))
        # Returning only 49 coefficients must not permit a 32-point convolution.
        self.assertEqual(last['resident_domain'], 64)
        self.assertEqual(last['split_prefix']['calls'], 13)
        self.assertEqual(last['split_prefix']['max_domain'], 16)
        # Full square backing (63), not its truncated 49-element view, stays live.
        self.assertEqual(last['product_stage_host_bytes'], 8*15 + 4*(16+63+49))
        self.assertEqual(last['square_stage_host_bytes'], 8*15 + 4*(16+32+63))
        self.assertEqual(last['old_prefix_retention_envelope_bytes'], 8*15 + 4*(16+32+63+49))

    def test_routes_are_scenarios_and_need_more_than_device_capacity(self):
        result = M['build']()
        M['check'](result)
        hybrid = result['routes']['hybrid']
        self.assertGreater(hybrid['old_prefix_retention_envelope_bytes'], 800*(1<<30))
        self.assertLess(hybrid['split_newton_host_principal_bytes'], 700*(1<<30))
        self.assertEqual(hybrid['candidate_early_support_release_newton_bytes'], 576*(1<<30)-4)
        self.assertIsNone(result['complete_route_runtime_seconds'])
        self.assertTrue(result['routes']['distributed']['missing'])


if __name__ == '__main__':
    unittest.main()
