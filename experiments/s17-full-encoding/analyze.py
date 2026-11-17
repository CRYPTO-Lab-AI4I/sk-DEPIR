#!/usr/bin/env python3
"""Reconcile a completed full encoding; compare CPU/GPU hashes when both passed."""
import argparse
import json
from pathlib import Path
import runpy

HERE = Path(__file__).resolve().parent
PHASES = {'load-code', 'database', 'encoder-setup', 'encoding', 'verify-systematic',
          'verify-curves', 'verify-canonical', 'pack-output', 'readback-packed'}


def inspect_phases(phases, backend):
    assert set(phases) == PHASES
    assert all(p['seconds'] >= 0 and p['profiled'] is True for p in phases.values())
    assert phases['verify-systematic']['symbols'] == 3876566203
    assert phases['verify-curves']['curves'] == 16
    assert phases['verify-canonical']['canonical_bytes'] == (1 << 34)*4
    assert phases['pack-output']['bytes'] == 34 << 30
    assert phases['readback-packed']['full_digest'] == phases['pack-output']['sha256']
    if backend == 'gpu':
        assert phases['encoding']['native']['timing_source'] == 'synchronized-host-wall'
    return dict(canonical_sha256=phases['verify-canonical']['sha256'],
                packed_sha256=phases['pack-output']['sha256'])


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('name')
    parser.add_argument('backend', choices=('cpu', 'gpu'))
    parser.add_argument('--forensic', action='store_true', help='seal review of a failed attempt; never mark it passed')
    args = parser.parse_args()
    out = HERE/'runs'/args.name/args.backend
    record = json.loads((out/'results.json').read_text())
    result = runpy.run_path(str(HERE.parent/'s17-arithmetic/analyze.py'))['reconcile'](out, require_success=not args.forensic)
    if args.forensic:
        assert record['status'] != 'passed', 'use normal successful validation'
        result.update(scope='forensic reconciliation of failed attempt; no verified full encoding', tests=['source/evidence/cache seals only'], failure=record.get('error'))
        (out/'failure-analysis.json').write_text(json.dumps(result, indent=2)+'\n')
        print(json.dumps(result, indent=2))
        return
    assert record['backend'] == args.backend
    assert record['probe_sha256'] == record['source_sha256']['experiments/s17-full-encoding/probe.go.txt']
    phases = record['phases']['phases']
    digests = inspect_phases(phases, args.backend)
    cache = Path(record['cache'])
    assert json.loads((cache/'phases.json').read_text()) == record['phases']
    assert (cache/'word-packed.bin').stat().st_size == 34 << 30
    assert record['code_manifest']['files']['online.bin']['sha256'] == phases['load-code']['source_sha256']
    rows = []
    with (out/'encoding-pipeline.log').open() as stream:
        for line in stream:
            if line.startswith('{'):
                row = json.loads(line)
                if row.get('phase') == 'workflow':
                    rows.append(row)
    assert any(row.get('event') == 'completed' and row.get('canonical_sha256') ==
               digests['canonical_sha256'] and row.get('packed_sha256') == digests['packed_sha256']
               for row in rows)
    other = 'gpu' if args.backend == 'cpu' else 'cpu'
    peer = out.parent/other/'results.json'
    comparison = 'other result not available/passed yet'
    if peer.exists():
        other_result = json.loads(peer.read_text())
        if other_result['status'] == 'passed':
            assert inspect_phases(other_result['phases']['phases'], other) == digests
            assert other_result['code_manifest'] == record['code_manifest']
            comparison = 'CPU/GPU complete canonical and packed hashes match'
    result.update(backend=args.backend, digests=digests, comparison=comparison,
                  tests=['full code hash/load', 'all3876566203systematic symbols', '16 independent scalar curves', 'all-value range/canonical SHA256', '34GiB tight packing/full disk digest/1027 positions'],
                  phases=phases, scope='full s17 standalone mathematical encoding with profiled timing; producer packed readback, no extra34GiB rehash in review')
    (out/'analysis.json').write_text(json.dumps(result, indent=2)+'\n')
    print(json.dumps(result, indent=2))


if __name__ == '__main__':
    main()
