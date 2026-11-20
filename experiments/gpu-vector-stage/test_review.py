#!/usr/bin/env python3
"""Reviewer rejects missing routes and byte corruption, not just success flags."""
from pathlib import Path
import runpy
import tempfile
import unittest
HELP=runpy.run_path(str(Path(__file__).with_name('run.py')))


class ReviewTest(unittest.TestCase):
    def fixture(self,directory):
        rows=[]
        for s,dims in ((4,(4,5,8)),(17,(7,12,34))):
            for dim in dims:
                n=1<<dim;longest=1<<((dim+1)//2)
                for columns in (False,True):
                    length=1<<(dim-dim//2 if columns else dim//2);jobs=n//length;count=1 if dim==34 else min(3,jobs)
                    for first in (0,jobs-count):
                        label=f's{s}-d{dim}-'+('columns-'if columns else'rows-')+str(first)
                        width=4 if s==4 else 8;data=bytes(length*count*width)
                        for suffix in ('expected','forward','input','inverse'):(directory/(label+'-'+suffix+'.bin')).write_bytes(data)
                        rows.extend([dict(kind='measurement',case=label,s=s,dimension=dim,columns=columns,first=first,count=count,length=length,word_bytes=width,capacity_words=min(n,3*longest),owned_device_bytes=1,peak_rss_bytes=1,setup_seconds=1,forward_kernel_wall_seconds=1,inverse_kernel_wall_seconds=1,matches=True),dict(phase=label,event='completed',completed=4,total=4)])
        rows.append(dict(kind='completed',cases=24));return rows

    def test_outputs_and_missing_case(self):
        with tempfile.TemporaryDirectory()as tmp:
            directory=Path(tmp);rows=self.fixture(directory)
            self.assertEqual(len(HELP['review'](rows,directory)),24)
            with self.assertRaises(AssertionError):HELP['review'](rows[1:],directory)

    def test_corruption(self):
        with tempfile.TemporaryDirectory()as tmp:
            directory=Path(tmp);rows=self.fixture(directory)
            p=next(directory.glob('*-forward.bin'));data=bytearray(p.read_bytes());data[0]=1;p.write_bytes(data)
            with self.assertRaises(AssertionError):HELP['review'](rows,directory)


if __name__=='__main__':unittest.main()
