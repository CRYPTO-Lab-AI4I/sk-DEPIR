#!/usr/bin/env python3
import runpy
from pathlib import Path
import tempfile
import unittest
HELP=runpy.run_path(str(Path(__file__).with_name('analyze.py')))

class Review(unittest.TestCase):
    def test_first_tail_word_and_length_difference(self):
        with tempfile.TemporaryDirectory()as path:
            a=Path(path)/'send';b=Path(path)/'receive'
            data=b'\0'*(65536*4)+(17).to_bytes(4,'little');a.write_bytes(data);b.write_bytes(data)
            self.assertTrue(HELP['compare_snapshots'](a,b)['equal'])
            b.write_bytes(data[:-4]+(19).to_bytes(4,'little'))
            r=HELP['compare_snapshots'](a,b)
            self.assertEqual((r['first_differing_word'],r['left_word'],r['right_word']),(65536,17,19))
            b.write_bytes(data[:-4]);r=HELP['compare_snapshots'](a,b)
            self.assertEqual(r['first_differing_word'],65536);self.assertIsNone(r['right_word'])

if __name__=='__main__':unittest.main()
