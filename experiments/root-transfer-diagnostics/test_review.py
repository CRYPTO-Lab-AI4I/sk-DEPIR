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

    def test_wide_index_corruption(self):
        with tempfile.TemporaryDirectory()as path:
            a=Path(path)/'send';b=Path(path)/'receive'
            a.write_bytes((17).to_bytes(8,'little'));b.write_bytes((17|(1<<63)).to_bytes(8,'little'))
            r=HELP['compare_snapshots'](a,b,8)
            self.assertEqual((r['first_differing_word'],r['left_word'],r['right_word']),(0,17,17|(1<<63)))

    def test_support_proof_pair_and_coverage(self):
        proof='01'*64
        sender=dict(kind='wire-proof',phase='support',sequence=0,case=0,sample=0,rank=0,peer=1,direction='send',words=5,own_proof=proof,remote_proof=proof,matches=True)
        receiver=dict(sender,rank=1,peer=0,direction='receive')
        row=dict(kind='measurement',case=0,sample=0,rank=1,shard_count=5)
        HELP['review_proofs']([sender,receiver,row])
        with self.assertRaises(AssertionError):HELP['review_proofs']([receiver,row])
        with self.assertRaises(AssertionError):HELP['review_proofs']([sender,receiver,dict(row,shard_count=6)])

if __name__=='__main__':unittest.main()
