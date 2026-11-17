#!/usr/bin/env python3
"""Reproduce workflow payload/geometry scenarios; never infer measured runtime."""
import json
import hashlib
import math
from pathlib import Path
import runpy

HERE = Path(__file__).resolve().parent
ADMISSION = runpy.run_path(str(HERE.parent / 's18-preprocessing-admission/memory.py'))['model']
SPLIT = runpy.run_path(str(HERE.parent / 'preprocessing-tiles/model.py'))['split_shape']
GIB = 1 << 30


def padded(n):
    return max(4, 1 << (n - 1).bit_length())


def split_prefix(a, b, limit, block):
    """Mirror prefix overlap-add geometry, not arithmetic or a supported backend."""
    a, b = min(a, limit), min(b, limit)
    pairs = []
    for i in range(0, a, block):
        for j in range(0, b, block):
            if i + j >= limit:
                continue
            ac, bc = min(block, a - i), min(block, b - j)
            pairs.append((padded(ac + bc - 1), min(limit - i - j, ac + bc - 1)))
    return dict(calls=len(pairs), max_domain=max(x[0] for x in pairs),
                max_temporary_bytes=4 * max(x[1] for x in pairs))


def newton_trace(n, k):
    """Match polyInverseRecover and polyMulLow lengths, including backing arrays."""
    target = n - k
    old = 1
    rows = []
    while old < target:
        output = min(2 * old, target)
        square = 2 * old - 1
        a, b = min(k + 1, output), min(square, output)
        domain = padded(a + b - 1)
        # P/root alias counted once. A truncated square view retains full backing.
        square_stage = 8 * k + 4 * (k + 1 + old + square)
        product_stage = 8 * k + 4 * (k + 1 + square + output)
        principal = max(square_stage, product_stage)
        retention = 8 * k + 4 * (k + 1 + old + square + output)
        row = dict(old=old, square_backing=square, output=output, a=a, b=b,
                   resident_domain=domain, host_principal_bytes=principal,
                   square_stage_host_bytes=square_stage, product_stage_host_bytes=product_stage,
                   old_prefix_retention_envelope_bytes=retention)
        if domain > n // 4:
            row['split_prefix'] = split_prefix(a, b, output, n // 8)
        rows.append(row)
        old = output
    return rows


def build():
    s, k = 18, 16300262095
    n = 1 << (2 * s)
    base = ADMISSION(s, k)
    split = SPLIT(n, k)
    newton = newton_trace(n, k)
    large = [r for r in newton if 'split_prefix' in r]
    peak = max(r['host_principal_bytes'] for r in newton)
    # Proposed evenly sharded arrays; no central gather or replicated full support.
    host = dict(discovery=base['discovery_principal_host_bytes'],
                root_merge=8*k + 8*(k+1),
                newton=peak, identity=8*k + 8*n + 4,
                final_streamed_export=4*n + 4 + (1 << 20))
    host_split_newton = max(max(r['square_stage_host_bytes'], r['product_stage_host_bytes'] +
        r.get('split_prefix', {}).get('max_temporary_bytes', 0)) for r in newton)
    retention_newton = max(r['old_prefix_retention_envelope_bytes'] +
        r.get('split_prefix', {}).get('max_temporary_bytes', 0) for r in newton)
    sources = dict(
        full_s17='experiments/gpu-s17-full-preprocessing/runs/resume-02/results.json',
        full_s17_analysis='experiments/gpu-s17-full-preprocessing/runs/resume-02/analysis.json',
        bounded_components='experiments/preprocessing-tiles/runs/components-01/components.json',
        exact_s18='experiments/s18-preprocessing-admission/runs/counts-01/memory-model.json')
    root = Path(__file__).resolve().parents[2]
    full = json.loads((root / sources['full_s17']).read_text())
    components = json.loads((root / sources['bounded_components']).read_text())
    assert full['status'] == 'passed'
    assert full['job_id'] == '75832'
    assert len(components['fixtures']) == 12
    # Source-backed s17 root phase reported by the completed full-run review.
    phases = json.loads((root / sources['full_s17_analysis']).read_text())['phase_seconds']
    root_s17 = phases['root-product'][0]
    k17 = 3876566203
    scale = k / k17
    root_scenarios = dict(
        baseline_seconds=root_s17,
        linear_K_work_seconds=root_s17*scale,
        K_log2K_squared_work_seconds=root_s17*scale*(math.log2(k)/math.log2(k17))**2,
        caveat='Same effective root throughput scenarios only, not route forecasts or confidence limits; no fourfold GPU scaling assumed.')
    # Hypothetical balanced redistributions: fraction crossing four node boundaries.
    # Two passes is a proposed schedule; six mirrors current local-layout call count.
    network = []
    for passes in (2, 6):
        transform_bytes = passes * (8*n) * 3 // 4
        product_bytes = 3 * transform_bytes
        network.append(dict(redistributions_per_transform=passes,
            cross_node_bytes_per_full_transform=transform_bytes,
            cross_node_bytes_per_full_product=product_bytes,
            optimistic_seconds_by_per_node_egress_GiBs={str(rate): product_bytes/4/GIB/rate for rate in (3,12,25)},
            excludes='Taylor source exchanges, roots, host uploads/downloads, synchronization/latency, contention and persistence; no bandwidth measured.'))
    audit_paths = list(sources.values()) + [str(HERE.relative_to(root) / 'model.py'),
        'experiments/s18-preprocessing-admission/memory.py', 'experiments/preprocessing-tiles/model.py',
        'encoding/liftrs-shared/orbit_polynomial.go', 'encoding/liftrs-shared/orbit_product.go',
        'encoding/liftrs-shared/preprocessing_core.go', 'encoding/liftrs-cuda/preparation.cu']
    hashes = {name:hashlib.sha256((root/name).read_bytes()).hexdigest() for name in audit_paths}
    return dict(s=s,k=k,n=n,sources=sources,input_source_sha256=hashes,
        measured_s17_phase_seconds=phases,
        routes=dict(hybrid=dict(gpus=4,nodes=1,
            max_product_array_bytes=split['split_device_arrays'],
            max_product_array_bytes_per_gpu=split['split_device_arrays']//4,
            host_stage_principal_bytes=host,
            split_newton_host_principal_bytes=host_split_newton,
            old_prefix_retention_envelope_bytes=retention_newton,
            candidate_early_support_release_newton_bytes=host_split_newton-8*k,
            missing=['tiled base-grid discovery','tiled final spectrum and canonical export integration',
                     'prefix-aware split products in Newton/backend dispatch','host lifetime/headroom validation']),
          distributed=dict(gpus=16,nodes=4,
            max_product_array_bytes=24*n,max_product_array_bytes_per_gpu=24*n//16,
            sharded_host_stage_principal_bytes_per_node={name:(value+3)//4 for name,value in host.items()},
            missing=['cross-node transform transport/schedule','sharded discovery/support and polynomial state',
                     'parallel root work and deterministic merge','distributed atomic checkpoints/publication'])),
        newton=newton,oversized_newton_count=len(large),
        final_identity=split,
        storage=dict(canonical_bytes=base['canonical_retained_bytes'],
            two_each_phase_snapshot_upper_bytes=base['two_each_phase_snapshot_upper_bytes'],
            canonical_plus_snapshots_bytes=base['canonical_retained_bytes']+base['two_each_phase_snapshot_upper_bytes'],
            excludes='logs/profiles, filesystem overhead, failed attempts and simultaneous replacement of an existing complete stage'),
        network_scenarios=network,root_work_scenarios=root_scenarios,
        complete_route_runtime_seconds=None,
        scope='Source-derived principal payloads and hypothetical work/communication scenarios; not execution support, peak RAM, measured network, runtime forecast or budget admission.')


def check(result):
    n = result['n']
    assert result['oversized_newton_count'] == 3
    assert [r['resident_domain'] for r in result['newton'] if 'split_prefix' in r] == [n//2,n,n]
    assert all(r['split_prefix']['max_domain'] <= n//4 for r in result['newton'] if 'split_prefix' in r)
    assert result['newton'][-1]['split_prefix']['calls'] == 13
    assert result['final_identity']['subproducts'] == 14
    for route in result['routes'].values():
        assert route['max_product_array_bytes_per_gpu'] == 96*GIB
    assert result['routes']['hybrid']['old_prefix_retention_envelope_bytes'] > 800*GIB
    assert result['routes']['hybrid']['split_newton_host_principal_bytes'] < 700*GIB
    assert result['complete_route_runtime_seconds'] is None


if __name__ == '__main__':
    result = build()
    check(result)
    print(json.dumps(result, indent=2)+'\n', end='')
