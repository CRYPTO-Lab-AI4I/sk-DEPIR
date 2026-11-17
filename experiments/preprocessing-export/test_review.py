#!/usr/bin/env python3
"""Expected negative tests cannot mask unexpected failures or resource guards."""
import copy
from pathlib import Path
import runpy
import unittest
CHECK=runpy.run_path(str(Path(__file__).resolve().parents[1]/'s17-arithmetic/analyze.py'))['check_commands']

class ReviewTests(unittest.TestCase):
    def test_default_stays_strict(self):
        CHECK([dict(label='ordinary',exit_code=0,guard=None)])
        with self.assertRaises(AssertionError):
            CHECK([dict(label='negative',exit_code=1,expected_exit_code=1,guard=None)])

    def test_declared_failures_only(self):
        commands=[dict(label='ordinary',exit_code=0,guard=None),
                  dict(label='negative',exit_code=1,expected_exit_code=1,guard=None)]
        expected={'negative':1}
        CHECK(commands,expected)
        for field,value in [('exit_code',0),('exit_code',2),('exit_code',-9),
                            ('guard','RSS guard'),('expected_exit_code',0)]:
            bad=copy.deepcopy(commands);bad[1][field]=value
            with self.assertRaises(AssertionError):CHECK(bad,expected)
        with self.assertRaises(AssertionError):CHECK(commands[:1],expected)
        with self.assertRaises(AssertionError):CHECK(commands+[commands[1]],expected)
        bad=copy.deepcopy(commands);bad[0]['exit_code']=1
        with self.assertRaises(AssertionError):CHECK(bad,expected)

if __name__=='__main__':unittest.main()
