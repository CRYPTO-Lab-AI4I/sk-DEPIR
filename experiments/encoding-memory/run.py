#!/usr/bin/env python3
"""Allocation-free source/measurement audit; no encoding or hardware benchmark."""
import argparse
import hashlib
import json
import subprocess
from pathlib import Path

ROOT = Path(__file__).resolve().parents[2]
HERE = Path(__file__).resolve().parent
GIB = 1 << 30
SOURCES = [
    'encoding/liftrs-cuda/matrix.cu',
    'encoding/liftrs-shared/cuda_encoder.go',
    'encoding/liftrs-shared/scaling_online.go',
    'encoding/liftrs-shared/scaling_experimental.go',
    'encoding/liftrs-shared/scaling_matrix_cpp.go',
    'encoding/liftrs-shared/orbit.go',
    'encoding/liftrs-shared/exponent_set.go',
    'encoding/liftrs-shared/orbit_product.go',
    'encoding/liftrs-shared/orbit_polynomial.go',
    'encoding/liftrs-shared/s15_firstuse_probe_test.go',
    'pirconfig/configs_catalogue.json',
]
BASELINES = {
    'gpu': 'experiments/cuda-s16-encoding/runs/encoding-01/results.json',
    'cpu': 'experiments/cached-s16-encoding/runs/encoding-01/results.json',
    'preparation': 'experiments/gpu-preparation/runs/actual-01/results.json',
    'checkpoints': 'experiments/gpu-preparation/runs/actual-01/worker/checkpoints.json',
    'gpu_phases': 'experiments/cuda-s16-encoding/runs/encoding-01/encoding/phases.json',
    'cpu_phases': 'experiments/cached-s16-encoding/runs/encoding-01/encoding/phases.json',
}


def read(name):
    return json.loads((ROOT / name).read_text())


def payload(s, k):
    assert 2 <= s <= 17
    n = 1 << (2 * s)
    assert 0 < k <= n // 2
    length = max(4, 1 << (2 * k - 2).bit_length())
    assert 2 * length == n, 'This study assumes the catalogue s16/s17 prefix shape'
    base = 4
    ext = 4 if s <= 16 else 8
    domain = ext * n
    p = base * (k + 1)
    message = base * k
    word = base * n
    plan = p + domain
    # Models count owned array payload, not RSS, allocator slack or CUDA context.
    return {
        's': s, 'n': n, 'k': k, 'prefix_n': length,
        'extension_slot_bytes': ext, 'base_slot_bytes': base,
        'wide_domain_bytes': domain, 'p_bytes': p, 'message_bytes': message,
        'canonical_word_bytes': word, 'packed_word_bytes': (s * n + 7) // 8,
        'support_in_memory_bytes': 8 * k,
        'support_file_bytes': (4 if s <= 16 else 8) * k,
        'online_plan_bytes': plan,
        'prepared_minimum_retained_bytes': base * (n + 1) + domain + 8 * k,
        'identity_result_bytes': word,
        'cpu_online_principal_bytes': plan + 2 * message + 2 * domain + word,
        'gpu_online_principal_bytes': plan + 3 * domain,
        'gpu_host_plan_retained_bytes': plan + message + word,
        'gpu_host_plan_released_peak_bytes': max(plan, message + word),
        'gpu_two_workspace_bytes': plan + max(3 * ext * length, 2 * domain),
        # Proposed schedules: not implementations, limits or admission requests.
        'gpu_two_workspace_streamed_code_bytes': max(3 * ext * length + p, 2 * domain + (2 << 30)),
        'gpu_one_workspace_streamed_code_bytes': domain + (2 << 30),
        'gpu_compact34_current_layout_bytes': ((4 * n * 34 + 7) // 8 + p) if s == 17 else None,
        'streamed_code_extra_per_reused_encode_bytes': plan,
        # Two disk work arrays, saved prefix operand, canonical routing file,
        # spectrum/P, and final packed file; excludes support/G/preparation files.
        'disk_backed_example_storage_bytes': 2 * domain + ext * length + word + plan + (s * n + 7) // 8,
        'disk_backed_example_scratch_bytes': 2 * domain + ext * length + word,
        # Illustrative p complete read/write passes per transform (3L+2N).
        # Not a bound for every algorithm; Taylor dependencies may require more.
        'tile_transfer_scenarios_bytes': {
            str(passes): 2 * passes * ext * (3 * length + 2 * n)
            for passes in (1, 4, 16)
        },
    }


def checks():
    m = payload(16, 918672421)
    assert m['prefix_n'] == 1 << 31
    assert m['online_plan_bytes'] == 20854558905 - 33
    assert m['packed_word_bytes'] == 8 * GIB
    assert m['gpu_online_principal_bytes'] - m['gpu_two_workspace_bytes'] == 16 * GIB
    assert m['cpu_online_principal_bytes'] == 16 * m['n'] + 12 * m['k'] + 4
    w = payload(17, 3870933765)
    assert w['wide_domain_bytes'] == 128 * GIB and w['packed_word_bytes'] == 34 * GIB
    assert w['tile_transfer_scenarios_bytes']['1'] == 896 * GIB
    assert w['gpu_two_workspace_bytes'] > 140 * GIB
    assert w['gpu_compact34_current_layout_bytes'] > 140 * GIB
    assert w['gpu_one_workspace_streamed_code_bytes'] > 96 * GIB
    try:
        payload(17, 1)
    except AssertionError:
        pass
    else:
        raise AssertionError('accepted incompatible prefix shape')


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--check', action='store_true', help='check arithmetic without writing results')
    args = parser.parse_args()
    checks()
    if args.check:
        print('Memory accounting and schedule arithmetic checks pass')
        return
    gpu, cpu, prep = (read(BASELINES[key]) for key in ('gpu', 'cpu', 'preparation'))
    assert all(r['status'] == 'passed' for r in (gpu, cpu, prep))
    assert gpu['encoding']['canonical_sha256'] == cpu['encoding']['canonical_sha256']
    assert gpu['encoding']['packed_sha256'] == cpu['encoding']['packed_sha256']
    k = gpu['encoding']['k']
    catalogue = read('pirconfig/configs_catalogue.json')
    # Catalogue is a JSON object with a configurations array.
    def configs(value):
        if isinstance(value, list):
            return value
        return next(v for v in value.values() if isinstance(v, list) and v and isinstance(v[0], dict) and 'id' in v[0])
    c17 = next(c for c in configs(catalogue) if c['id'] == 'binary-s17-m2-t5')
    m16, m17 = payload(16, k), payload(17, c17['capacity']['k'])
    assert c17['capacity']['status'] == 'estimated'
    peak = lambda r: max(c['peak_rss_bytes'] for c in r['commands'])
    phases_gpu = {p['phase']: p for p in read(BASELINES['gpu_phases'])}
    phases_cpu = {p['phase']: p for p in read(BASELINES['cpu_phases'])}
    assert phases_gpu['load-online']['plan_bytes'] == m16['online_plan_bytes']
    assert phases_cpu['load-online']['plan_bytes'] == m16['online_plan_bytes']
    owned = gpu['encoding']['encodings'][0]['cuda_stages']['owned_device_bytes']
    assert 0 < owned - m16['gpu_online_principal_bytes'] < 2 << 20
    assert 0 < peak(cpu) - m16['cpu_online_principal_bytes'] < GIB
    assert 0 < peak(gpu) - m16['gpu_host_plan_retained_bytes'] < GIB
    checkpoints = read(BASELINES['checkpoints'])
    stage = gpu['encoding']['encodings'][1]['cuda_stages']
    h2d = m16['message_bytes'] / (stage['host_to_device_ns'] / 1e9)
    d2h = m16['canonical_word_bytes'] / (stage['device_to_host_ns'] / 1e9)
    # A diagnostic sensitivity calculation, not a streamed bandwidth measurement.
    sensitivity = {}
    for passes, traffic in m17['tile_transfer_scenarios_bytes'].items():
        sensitivity[passes] = {
            'bytes': traffic,
            'copy_seconds_at_s16_directional_rates': traffic / 2 / h2d + traffic / 2 / d2h,
            'disk_seconds_at_hypothetical_1_GiB_per_second_aggregate': traffic / GIB,
        }
    paths = SOURCES + list(BASELINES.values()) + [str(Path(__file__).relative_to(ROOT))]
    result = {
        'kind': 'analytical_models_with_retained_measured_baselines',
        'revision': subprocess.check_output(['git', 'rev-parse', 'HEAD'], cwd=ROOT, text=True).strip(),
        'command': 'python3 -B experiments/encoding-memory/run.py',
        'source_sha256': {name: hashlib.sha256((ROOT / name).read_bytes()).hexdigest() for name in paths},
        'models': {'s16_exact_k': m16, 's17_estimated_k': m17},
        'measured_s16': {
            'gpu_job': gpu['job_id'], 'cpu_job': cpu['job_id'], 'preparation_job': prep['job_id'],
            'gpu_owned_device_bytes': owned,
            'gpu_array_model_residual_bytes': owned - m16['gpu_online_principal_bytes'],
            'gpu_host_peak_rss_bytes': peak(gpu), 'cpu_peak_rss_bytes': peak(cpu),
            'preparation_peak_rss_bytes': peak(prep),
            'preparation_max_live_heap_after_gc_bytes': max(p['live_heap_after_gc'] for p in checkpoints),
            'preparation_complete_live_heap_after_gc_bytes': checkpoints[-1]['live_heap_after_gc'],
            'gpu_encode_seconds': [p['seconds'] for p in gpu['encoding']['encodings']],
            'h2d_warmed_bytes_per_second': h2d, 'd2h_warmed_bytes_per_second': d2h,
        },
        's17_transfer_sensitivity_not_runtime_predictions': sensitivity,
        'limits': 'No encoding, new hardware measurement, GPU allocation or s17 arithmetic implementation. Array payloads exclude allocator/context/pinning overhead; proposed schedules need dependency/correctness/latency validation. Preprocessing peak and storage are not bounded by retained-result sizes. Streaming PCIe/disk sensitivity is not a runtime prediction. No new job or memory-time selection.',
    }
    (HERE / 'results.json').write_text(json.dumps(result, indent=2) + '\n')
    print(json.dumps({'revision': result['revision'], 'status': 'models_reconciled',
                      'source_files': len(paths), 's17_model_GiB': {key: round(value/GIB, 3) for key, value in m17.items() if key.endswith('_bytes') and isinstance(value, int)}}, indent=2))


if __name__ == '__main__':
    main()
