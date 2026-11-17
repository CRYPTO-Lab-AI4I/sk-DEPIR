#!/usr/bin/env python3
"""No allocations: policies and completion assertions, including negative cases."""
import copy
import runpy
from pathlib import Path
import unittest

HERE = Path(__file__).resolve().parent
RUN = runpy.run_path(str(HERE/'run.py'))
ANALYZE = runpy.run_path(str(HERE/'analyze.py'))


class EncodingAdmissionTests(unittest.TestCase):
    def test_limits_and_verified_manifest(self):
        manifest = RUN['validate']()
        self.assertEqual(manifest['header']['N'], 1 << 34)
        self.assertEqual(RUN['policy']('cpu')['gpus'], 0)
        self.assertEqual(RUN['policy']('gpu')['gpus'], 4)
        self.assertEqual(sum(RUN['policy'](b)['retained_cap_bytes'] for b in ('cpu', 'gpu')), 128 << 30)

    def test_gpu_host_cache_headroom(self):
        p = RUN['policy']('gpu')
        self.assertEqual(p['host_request_bytes'], 640 << 30)
        self.assertEqual(p['cgroup_guard_bytes'], 608 << 30)
        code_bytes = RUN['validate']()['files']['online.bin']['bytes']
        # Allow the full process guard AND both complete files charged as cache.
        self.assertGreater(p['cgroup_guard_bytes'] - p['rss_guard_bytes'] - code_bytes - (34 << 30), 100 << 30)
        launcher = (HERE/'gpu.sbatch').read_text()
        self.assertIn('#SBATCH --mem=640G', launcher)

    def test_cpu_reference_rejects_different_outputs(self):
        reference = dict(status='passed', backend='cpu', code_manifest={'identity': 'same'},
                         phases={'phases': {p: {'sha256': p} for p in ('verify-canonical', 'pack-output')}})
        gpu = copy.deepcopy(reference)
        RUN['compare_cpu_reference'](gpu, reference)
        for phase in ('verify-canonical', 'pack-output'):
            bad = copy.deepcopy(gpu)
            bad['phases']['phases'][phase]['sha256'] = 'wrong'
            with self.assertRaises(AssertionError):
                RUN['compare_cpu_reference'](bad, reference)

    def test_reject_incomplete_results(self):
        phases = {name: dict(seconds=1, profiled=True) for name in ANALYZE['PHASES']}
        phases['verify-systematic']['symbols'] = 3876566203
        phases['verify-curves']['curves'] = 16
        phases['verify-canonical'].update(canonical_bytes=(1 << 34)*4, sha256='a'*64)
        phases['pack-output'].update(bytes=34 << 30, sha256='b'*64)
        phases['readback-packed']['full_digest'] = 'b'*64
        phases['encoding']['native'] = dict(timing_source='synchronized-host-wall')
        self.assertEqual(ANALYZE['inspect_phases'](phases, 'gpu')['packed_sha256'], 'b'*64)
        for phase, key, value in [('verify-systematic', 'symbols', 1),
                                   ('pack-output', 'bytes', 1),
                                   ('readback-packed', 'full_digest', 'wrong')]:
            damaged = copy.deepcopy(phases)
            damaged[phase][key] = value
            with self.assertRaises(AssertionError):
                ANALYZE['inspect_phases'](damaged, 'cpu')


if __name__ == '__main__':
    unittest.main()
