#!/usr/bin/env python3
"""Small policy checks; importing the monitor must not signal any process."""
import unittest
from extend_guard import elapsed_seconds, power_action, budget_origin


class GuardPolicy(unittest.TestCase):
    def test_elapsed(self):
        for value, expected in [('01:25:36',5136), ('02:05',125), ('1-01:00:00',90000)]:
            self.assertEqual(elapsed_seconds(value),expected)

    def test_manual_pause_budget(self):
        # A long sleep/awake break and a new per-process clock origin must not
        # consume the remaining active budget or reset already consumed time.
        for now in [0.01, 100000]:
            origin, paused = budget_origin(now, 50000, 0, {'active_elapsed_s':10579})
            self.assertEqual(paused, 0)
            self.assertAlmostEqual(now-origin, 10579)
            self.assertAlmostEqual(19800-(now-origin), 9221)
        origin, paused = budget_origin(200, 100, 10, None)
        self.assertEqual(200-origin-paused, 90)

    def test_power(self):
        self.assertIsNone(power_action("'Battery Power' 40%;",False))
        self.assertEqual(power_action("'Battery Power' 10%;",False),'pause')
        self.assertEqual(power_action("'AC Power' 12%;",True),'resume')
        self.assertIsNone(power_action('unknown',False))


if __name__=='__main__':unittest.main()
