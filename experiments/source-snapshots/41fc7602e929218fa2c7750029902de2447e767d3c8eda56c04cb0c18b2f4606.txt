"""Check resource guards and large-domain arithmetic without large allocations."""
import json
from pathlib import Path
import runpy
import subprocess
import sys
import tempfile
import unittest

HERE=Path(__file__).resolve().parent
RUN=runpy.run_path(str(HERE/'run.py'))
AUDIT=runpy.run_path(str(HERE/'audit.py'))
GIB=1<<30


class AdmissionChecks(unittest.TestCase):
    def test_each_guard(self):
        policy=json.loads((HERE/'policy.json').read_text())
        check=RUN['guard_reason']
        normal=[100*GIB,400*GIB,50*GIB,20*GIB,20,100,policy]
        self.assertIsNone(check(*normal))
        for index,value,expected in [(0,225*GIB,'RSS guard'),(1,31*GIB,'node available-memory guard'),
                                     (2,65*GIB,'cache guard'),(3,129*GIB,'scratch guard'),(4,101,'command deadline')]:
            bad=normal[:];bad[index]=value
            self.assertEqual(check(*bad),expected)

    def test_array_and_file_counts(self):
        row=AUDIT['scenario'](16,919440829)
        self.assertEqual(row['n'],1<<32)
        self.assertEqual(row['online_file_bytes'],4*((1<<32)+919440830)+33)
        self.assertEqual(row['packed_word_bytes'],8*GIB)
        self.assertLess(row['code_and_two_packed_words_bytes'],64*GIB)
        upper=AUDIT['scenario'](16,(1<<32)//2)
        self.assertLess(upper['newton_old_g_and_support_slack_bytes'],109*GIB)
        self.assertLess(upper['online_arrays_bytes'],105*GIB)
        self.assertLess(upper['code_and_two_packed_words_bytes'],64*GIB)

    def test_symlinks_do_not_double_count_cache(self):
        with tempfile.TemporaryDirectory() as temporary:
            directory=Path(temporary)
            (directory/'data').write_bytes(b'fixture')
            (directory/'alias').symlink_to(directory/'data')
            self.assertEqual(RUN['tree_bytes'](directory),7)

    def test_stop_group_reaps_worker(self):
        child=subprocess.Popen([sys.executable,'-c','import time; time.sleep(60)'],start_new_session=True)
        RUN['stop_group'](child)
        self.assertIsNotNone(child.returncode)


if __name__=='__main__':unittest.main()
