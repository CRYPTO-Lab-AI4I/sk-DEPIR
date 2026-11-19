#!/usr/bin/env python3
from pathlib import Path
from unittest import mock
import copy
import hashlib
import json
import runpy
import tempfile
import unittest
HELP=runpy.run_path(str(Path(__file__).with_name('run.py')))


class Review(unittest.TestCase):
    def test_actual_mode_and_scratch_are_required(self):
        rows=[dict(kind='host',rank=r,host='a'if r<2 else'b',buffer_mode='staged',staging_bytes=524288)for r in range(4)]
        self.assertEqual(HELP['review_hosts'](rows,'staged')['3'],'b')
        with self.assertRaises(AssertionError):HELP['review_hosts'](rows,'direct')
        rows[3]['staging_bytes']=8192
        with self.assertRaises(AssertionError):HELP['review_hosts'](rows,'staged')

    def test_medians_exclude_oracle_artifact_and_other_rank(self):
        rows=[dict(kind='measurement',case=0,rank=0,s=8,support_count=7,sample=sample,
            **{field:value for field in HELP['TREE']['FIELDS']})for sample,value in [(-1,1000),(0,2),(1,4),(2,2000)]]
        rows.append(dict(rows[1],rank=1,seconds=3000))
        result=HELP['summarize_cases'](rows)[0]
        self.assertEqual(result['timed_samples'],2);self.assertEqual(result['median_seconds']['seconds'],3)

    def test_comparison_rejects_output_or_placement_changes(self):
        arms=[dict(name=name,mode=mode,status='passed',final_sha256={'0':'same'},normal=dict(hosts={'0':'node-a'}),
            cases=[dict(case=c,median_seconds=dict(seconds=value))for c in range(3)])
            for (name,mode),value in zip(HELP['ARMS'],[2,1,2,4])]
        ratios=HELP['compare'](arms)
        self.assertEqual(ratios[0]['staged_over_direct_pair_a'],2)
        self.assertEqual(ratios[0]['staged_over_direct_pair_b'],2)
        bad=copy.deepcopy(arms);bad[2]['final_sha256']['0']='different'
        with self.assertRaises(AssertionError):HELP['compare'](bad)
        bad=copy.deepcopy(arms);bad[3]['normal']['hosts']['0']='another-node'
        with self.assertRaises(AssertionError):HELP['compare'](bad)

    def test_next_arm_clears_small_fixture_state_and_uses_new_paths(self):
        class Stop(Exception):pass
        with tempfile.TemporaryDirectory()as tmp:
            path=Path(tmp);env=dict(PIR_ROOT_TREE_SMALL='1',PIR_ROOT_TREE_MPI_BUFFER='direct')
            def run(label,argv,timeout):
                self.assertEqual(label,'direct-a-normal')
                self.assertNotIn('PIR_ROOT_TREE_SMALL',env)
                self.assertEqual(env['PIR_ROOT_TREE_CHUNK_WORDS'],'65536')
                self.assertEqual(env['PIR_ROOT_TREE_OUTPUT'],str(path/'direct-a-normal'))
                self.assertIn('-root-mpi-buffer=direct',argv)
                self.assertIn(path/'direct-a-normal',argv)
                raise Stop()
            with self.assertRaises(Stop):HELP['TREE']['execute'](dict(env=env,cache=path,out=path,go='go',run=run),('binary','library','wrapper'),prefix='direct-a-')

    def test_failed_arm_is_saved_and_not_skipped(self):
        with tempfile.TemporaryDirectory()as tmp:
            path=Path(tmp);binary=path/'binary';library=path/'library';binary.write_bytes(b'b');library.write_bytes(b'l')
            context=dict(env={},cache=path,out=path,resources={})
            with mock.patch.dict(HELP['TREE'],prepare=lambda c:(binary,library,path/'wrapper'),execute=mock.Mock(side_effect=RuntimeError('bad transfer'))):
                with self.assertRaisesRegex(RuntimeError,'bad transfer'):HELP['suite'](context)
                self.assertEqual(HELP['TREE']['execute'].call_count,1)
            state=json.loads((path/'comparison-progress.json').read_text())
            self.assertEqual(state['status'],'failed');self.assertEqual(state['current_arm'],'staged-a');self.assertEqual(state['arms'],[])


    def test_paired_corruption_capture(self):
        analyze=runpy.run_path(str(Path(__file__).with_name('analyze.py')))
        with tempfile.TemporaryDirectory()as tmp:
            cache=Path(tmp);rows=[]
            payloads=[(1).to_bytes(4,'little')+(17).to_bytes(4,'little'),(1).to_bytes(4,'little')+bytes(4)]
            hashes=[hashlib.sha256(p).hexdigest()for p in payloads]
            for rank,direction in enumerate(('send','receive')):
                folder=cache/'direct-a-normal'/f'rank-{rank}';folder.mkdir(parents=True)
                (folder/'capture.bin').write_bytes(payloads[rank])
                rows.append(dict(kind='transport-failure',matches=False,direction=direction,rank=rank,peer=1-rank,host='same-node',
                    case=2,sample=5,phase='root',sequence=1,word_bytes=4,words=2,file='capture.bin',sha256=hashes[rank],
                    own_proof=hashes[rank]*2,remote_proof=hashes[1-rank]*2))
            result=analyze['review_captures'](rows,cache,'direct-a');pair=result['paired_snapshots'][0]
            self.assertTrue(pair['same_host']);self.assertEqual(pair['changed_words_received_as_zero'],1)
            self.assertEqual(pair['comparison']['first_differing_word'],1)
            self.assertTrue(all(r['capture_matches_logged_sha']for r in result['captures']))
            with self.assertRaises(AssertionError):analyze['review_captures'](rows[:1],cache,'direct-a')

if __name__=='__main__':unittest.main()
