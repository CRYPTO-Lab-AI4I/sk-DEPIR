#!/usr/bin/env python3
from pathlib import Path
import runpy
import tempfile
import json
import unittest
from unittest import mock
HELP=runpy.run_path(str(Path(__file__).with_name('run.py')))


def fixture():
    rows=[]
    for rank in (0,1):
        rows.extend([dict(kind='host',rank=rank,host='node'+str(rank)),dict(rank=rank,phase='workflow',event='completed')])
        for size in HELP['SIZES']:
            for sample in (-1,0,1,2):
                row=dict(kind='measurement',rank=rank,bytes=size,sample=sample,matches=True,owned_device_bytes=64<<20,pinned_host_bytes=2*size,peak_rss_bytes=512<<20)
                row.update({name:0.01 for name in ('host_seconds','producer_seconds','d2h_seconds','exchange_seconds','h2d_seconds','consumer_seconds','pipeline_seconds')})
                rows.append(row)
    return rows


class Review(unittest.TestCase):
    def test_complete_and_remote_byte_count(self):
        rows=fixture();HELP['verify'](rows)
        summary=runpy.run_path(str(Path(__file__).with_name('analyze.py')))['summarize'](rows)
        self.assertEqual(summary[-1]['remote_MiB_per_node'],16)
        self.assertEqual(summary[-1]['remote_payload_GiBs_per_node'],1.5625)
    def test_reject_missing_rank_coverage_bad_caps_or_timing(self):
        rows=fixture();rows.pop()
        with self.assertRaises(AssertionError):HELP['verify'](rows)
        rows=fixture();rows[2]['owned_device_bytes']=129<<20
        with self.assertRaises(AssertionError):HELP['verify'](rows)
        rows=fixture();rows[2]['exchange_seconds']=0.02
        with self.assertRaises(AssertionError):HELP['verify'](rows)
        rows=fixture();rows[14]['host']='node0'
        with self.assertRaises(AssertionError):HELP['verify'](rows)
    def test_optional_command_summary_does_not_invent_RSS(self):
        shared=runpy.run_path(str(Path(__file__).resolve().parent.parent/'s17-arithmetic/analyze.py'))['reconcile']
        with tempfile.TemporaryDirectory()as temporary, mock.patch('subprocess.check_output',return_value=b''):
            out=Path(temporary)
            record=dict(status='passed',job_id='test',revision='test',cache=temporary,
                source_sha256={},cache_sha256={},evidence_sha256={},commands=[dict(label='transfer',seconds=1,exit_code=0)])
            (out/'results.json').write_text(json.dumps(record))
            with self.assertRaises(KeyError):shared(out)
            result=shared(out,command_fields=('label','seconds','exit_code'))
            self.assertEqual(result['commands'],record['commands'])
            self.assertNotIn('peak_rss_bytes',result['commands'][0])
            for bad in (dict(exit_code=1),dict(guard='RSS')):
                record['commands'][0].update(bad)
                (out/'results.json').write_text(json.dumps(record))
                with self.assertRaises(AssertionError):
                    shared(out,command_fields=('label','seconds','exit_code'))
                record['commands'][0]=dict(label='transfer',seconds=1,exit_code=0)
    def test_snapshot_is_git_blob_not_ignored_worktree(self):
        blob=b'package fixture\n';config=b'{}'
        tree=('100644 blob '+'a'*40+' '+str(len(blob))+'\tcode.go\n'
             +'100644 blob '+'b'*40+' 2\tpirconfig/config.json\n'
             +'100644 blob '+'c'*40+' 37479757\texperiments/x/artifacts/paper.json\n'
             +'100644 blob '+'d'*40+' 40\tdocs/no.md\n'
             +'100644 blob '+'e'*40+' 40\texperiments/x/runs/evidence.json\n')
        response=b'a'*40+b' blob '+str(len(blob)).encode()+b'\n'+blob+b'\n'+b'b'*40+b' blob 2\n'+config+b'\n'
        with tempfile.TemporaryDirectory()as temporary, mock.patch('subprocess.check_output',side_effect=[tree,response])as commands:
            root=Path(temporary);hashes=HELP['snapshot']('revision',root)
            self.assertEqual(set(hashes),{'code.go','pirconfig/config.json'})
            self.assertEqual((root/'code.go').read_bytes(),blob)
            self.assertEqual((root/'pirconfig/config.json').read_bytes(),config)
            self.assertFalse((root/'docs/no.md').exists())
            self.assertNotIn(b'paper.json',commands.call_args.kwargs['input'])
    def test_snapshot_size_limits_fail_before_content_read(self):
        for entries in ([('oversized.go',(32<<20)+1)],[(str(i)+'.go',30<<20)for i in range(5)]):
            tree=''.join('100644 blob '+'a'*40+' '+str(size)+'\t'+name+'\n' for name,size in entries)
            with tempfile.TemporaryDirectory()as temporary,mock.patch('subprocess.check_output',return_value=tree)as commands:
                with self.assertRaises(ValueError):HELP['snapshot']('revision',Path(temporary))
                self.assertEqual(commands.call_count,1)


if __name__=='__main__':unittest.main()
