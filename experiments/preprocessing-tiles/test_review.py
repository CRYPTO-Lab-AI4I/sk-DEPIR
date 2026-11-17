#!/usr/bin/env python3
"""Small deterministic checks for geometry and full-output hash admission."""
import hashlib,json,runpy,tempfile,unittest
from pathlib import Path
HERE=Path(__file__).resolve().parent
M=runpy.run_path(str(HERE/'model.py'))
VERIFY=runpy.run_path(str(HERE/'run.py'))['verify_outputs']

class ComponentTests(unittest.TestCase):
    def test_shape(self):
        self.assertEqual(M['check']()['split_device_arrays'],384<<30)
        for n,k in [(63,15),(64,1),(64,32)]:
            with self.assertRaises(ValueError):M['split_shape'](n,k)
    def test_hashes_and_completion(self):
        with tempfile.TemporaryDirectory() as tmp:
            root=Path(tmp);data=b'actual output';(root/'out.bin').write_bytes(data)
            expected=[dict(kind='product',index=0,output_sha256=hashlib.sha256(data).hexdigest())]
            row=dict(kind='product',index=0,output_file='out.bin',matches=True)
            log=root/'log';log.write_text(json.dumps(row)+'\n'+json.dumps(dict(kind='completed'))+'\n')
            self.assertEqual(len(VERIFY(log,root,expected)),1)
            (root/'out.bin').write_bytes(b'corrupt')
            with self.assertRaises(AssertionError):VERIFY(log,root,expected)
            (root/'out.bin').write_bytes(data);log.write_text(json.dumps(row)+'\n')
            with self.assertRaises(AssertionError):VERIFY(log,root,expected)

if __name__=='__main__':unittest.main()
