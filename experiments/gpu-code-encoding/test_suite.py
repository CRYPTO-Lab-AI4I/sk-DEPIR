#!/usr/bin/env python3
"""Check suite selectors and environment admission without allocating a GPU."""
import json
from pathlib import Path
import runpy
import tempfile
import unittest

HERE = Path(__file__).resolve().parent


class ClientSuiteTests(unittest.TestCase):
    def test_gpu_calls_are_enabled_and_bounded(self):
        suite = runpy.run_path(str(HERE/'suite.py'))['run_suite']
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            env, calls = {}, []

            def run(label, argv, cwd):
                calls.append((label, list(map(str, argv)), dict(env)))
                if label == 'code-memcheck':
                    (root/'code-memcheck.log').write_text(
                        'ERROR SUMMARY: 0 errors\nLEAK SUMMARY: 0 bytes\n')

            suite(dict(run=run, source=root, cache=root, out=root, env=env,
                       go='go', library=root/'plugin.so'))
            records = {label: (argv, current) for label, argv, current in calls}
            self.assertEqual(len(records), 7)
            for label in ('code-fixtures', 'code-wide-kernels', 'code-small'):
                self.assertEqual(records[label][1]['PIR_GPU_PREPARATION_LIBRARY'],
                                 str(root/'plugin.so'))
            self.assertEqual(records['code-wide-kernels'][1]['PIR_GPU_PREPARATION_CAP'],
                             str(64 << 20))
            self.assertEqual(records['code-wide-kernels'][1]['PIR_GPU_PREPARATION_DEVICES'],
                             '0,1,2,3')
            self.assertEqual(records['code-memcheck'][0][-1], '--small')
            self.assertNotIn('PIR_GPU_PREPARATION_LIBRARY', env)

    def test_review_rejects_missing_cuda_evidence(self):
        review = runpy.run_path(str(HERE/'analyze.py'))['review']
        with tempfile.TemporaryDirectory() as directory:
            out = Path(directory)
            (out/'code-small.log').write_text(
                '--- PASS: TestCUDACodeSmall (1s)\n--- PASS: TestCUDACodeResources (0s)\n')
            (out/'code-fixtures.log').write_text('--- PASS: TestCUDACodeNativeFixtures (1s)\n')
            (out/'code-wide-kernels.log').write_text(
                '--- PASS: TestGPUPreparationWide (1s)\n'
                '--- PASS: TestGPUPreparationPeerLifetime (1s)\n')
            for label, count in [('code-check', 60), ('code-memcheck', 12)]:
                rows = [dict(resident_code_cases=count, wide_positions=True),
                        dict(work='encoding-operations', completed=10, total=10,
                             event='completed')]
                text = ''.join(json.dumps(row)+'\n' for row in rows)
                if label == 'code-memcheck':
                    text += 'ERROR SUMMARY: 0 errors\nLEAK SUMMARY: 0 bytes\n'
                (out/(label+'.log')).write_text(text)
            result = {'tests': []}
            review(out, {}, result)
            self.assertEqual(result['client_native_cases'], 60)
            p = out/'code-memcheck.log'
            p.write_text(p.read_text().replace('ERROR SUMMARY: 0 errors',
                                              'ERROR SUMMARY: 1 errors'))
            with self.assertRaises(AssertionError):
                review(out, {}, {'tests': []})


if __name__ == '__main__':
    unittest.main()
