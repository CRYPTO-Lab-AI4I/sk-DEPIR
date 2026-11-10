#!/usr/bin/env python3
"""Current array/storage scenarios; estimates are not measured s16 RSS or latency."""
import hashlib
import json
from pathlib import Path
import subprocess

HERE = Path(__file__).resolve().parent
ROOT = HERE.parents[1]
GIB = 1 << 30


def scenario(s, k):
    n = 1 << (2*s)
    assert 1 < k <= n//2
    prefix = 1 << (2*k-2).bit_length()
    root = 1 << k.bit_length()
    return {
        's': s, 'k': k, 'n': n,
        'discovery_with_both_index_lists_and_capacity_slack_bytes': 8*n+24*k+n//8,
        'root_product_illustration_bytes': 4*n+8*k+12*(k+1)+8*root+36*32768,
        'identity_or_newton_arrays_bytes': 20*n+8*k+4,
        'newton_old_g_and_support_slack_bytes': 22*n+10*k+4,
        'prepared_plus_executor_bytes': 12*n+10*k+4,
        'online_arrays_bytes': 4*(k+1+n)+4*(max(2*prefix, n)+k+n)+4*k+4*n,
        'online_file_bytes': 4*(k+1+n)+(33 if s >= 16 else 25),
        'generator_file_bytes': 4*(n-k), 'support_file_bytes': 4*k,
        'packed_word_bytes': (n*s+7)//8,
        'code_and_two_packed_words_bytes': 8*n+4*k+(33 if s >= 16 else 25)+2*((n*s+7)//8),
    }


def main():
    catalogue = json.loads((ROOT/'pirconfig/configs_catalogue.json').read_text())
    config = next(c for c in catalogue['configurations'] if c['code']['s'] == 16)
    rows = [scenario(15, 216914047), scenario(16, config['capacity']['k']), scenario(16, (1<<32)//2)]
    assert abs(rows[1]['online_arrays_bytes']/GIB-74.3) < .1
    assert rows[1]['code_and_two_packed_words_bytes'] < 64*GIB
    paths = [HERE/'audit.py', HERE/'policy.json', ROOT/'pirconfig/configs_catalogue.json',
             *[ROOT/'encoding/liftrs-shared'/n for n in
               ('orbit.go', 'scaling_experimental.go', 'factor_product.go', 'orbit_polynomial.go',
                'orbit_product.go', 'scaling_prepare_cpp.go', 'scaling_matrix_cpp.go', 's15_firstuse_probe_test.go')]]
    result = {
        'kind': 'analytical_principal_array_scenarios_not_measured_s16_rss',
        'revision': subprocess.check_output(['git','rev-parse','HEAD'], cwd=ROOT, text=True).strip(),
        'source_sha256': {str(p.relative_to(ROOT)): hashlib.sha256(p.read_bytes()).hexdigest() for p in paths},
        'capacity_status': ['exact_s15', 'estimated_s16', 'admission_capacity_ceiling_not_a_capacity_prediction'],
        'rows': rows,
        'assumptions': ['8-byte Go int; 4-byte coefficients', '25% illustrative index-list capacity slack',
                        'bounded flat factor batches; reusable FFT pair plus N-element transpose',
                        'old reciprocal included in conservative scenario',
                        'excludes GC-retained unreachable allocations, allocator/runtime, OS and thread overhead'],
        'time_budget': {
            'native_encoding_scenario_seconds': 705.864*4*(32/30),
            'interpretation': 'domain/level extrapolation from s15, not a measured s16 time or worker-scaling prediction',
            'preparation': 'unmeasured on this host; historical s15 setup spans a manual pause and is not a clean baseline',
            'admission': '24h bounded exploratory allocation; initial 21h, cached 2h, runner deadline 23.5h; no automatic retry'},
    }
    (HERE/'audit.json').write_text(json.dumps(result, indent=2)+'\n')
    print(json.dumps({k: round(v/GIB,3) for k,v in rows[1].items() if k.endswith('_bytes')}, indent=2))


if __name__ == '__main__':
    main()
