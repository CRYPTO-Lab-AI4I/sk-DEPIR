#!/usr/bin/env python3
"""Allocation-free streamed-constant and s17 width audit; no hardware execution."""
import hashlib
import json
import subprocess
from pathlib import Path

ROOT = Path(__file__).resolve().parents[2]
HERE = Path(__file__).resolve().parent
GIB = 1 << 30
TILE = 256 << 20
SOURCES = [
    'encoding/liftrs-cuda/matrix.cu', 'encoding/liftrs-cuda/bridge.go',
    'encoding/liftrs-shared/scaling_online.go',
    'encoding/liftrs-shared/orbit_arithmetic.go',
    'encoding/liftrs-shared/cpu_compact.go',
    'encoding/liftrs-shared/s15_firstuse_probe_test.go',
    'encoding/liftrs-cpp/c_api.cpp', 'encoding/liftrs-cpp/transform.hpp',
    'utils/gf2n.go', 'encoding/liftrs-shared/binary_field.go',
    'pirconfig/configs_catalogue.json', 'experiments/streamed-constants/run.py',
]
REFERENCES = {
    'gpu_reuse': 'experiments/cuda-buffer-reuse/runs/reuse-01/results.json',
    'gpu_compact': 'experiments/gpu-output-memory/runs/compact-01/results.json',
    'cpu_compact': 'experiments/cpu-output-memory/runs/compact-01/results.json',
}


def model(s, k):
    n = 1 << (2*s)
    assert s in (16, 17) and 0 < k <= n//2
    length = max(4, 1 << (2*k-2).bit_length())
    assert 2*length <= n
    ext = 4 if s == 16 else 8
    base = 4
    domain, p, message, output = ext*n, base*(k+1), base*k, base*n
    columns = 32*64*(1 << s)*ext
    staging = min(TILE, domain)
    code = p+domain
    return {
        's': s, 'n': n, 'k': k, 'prefix_n': length,
        'base_bits': s, 'extension_bits': 2*s,
        'index_bits': (n-1).bit_length(), 'count_bits': n.bit_length(),
        'base_slot_bytes': base, 'extension_slot_bytes': ext,
        'spectrum_bytes': domain, 'p_bytes': p, 'online_payload_bytes': code,
        'message_bytes': message, 'canonical_bytes': output,
        'packed_bytes': (s*n+7)//8, 'support_file_bytes': ext*k,
        'stream_upload_bytes_per_message': domain+message,
        'resident_code_initial_upload_bytes': code,
        'additional_code_upload_bytes_per_reused_message': domain+message,
        'prefix_p_unused_last_value_bytes': base,
        'spectrum_chunks': (domain+staging-1)//staging,
        # Fast mode borrows dead transpose scratch at pointwise multiplication.
        'gpu_fast_streamed_arrays_bytes': 2*domain,
        'gpu_one_arena_streamed_arrays_bytes': domain+staging,
        'gpu_resident_fast_arrays_bytes': 3*domain+p,
        'gpu_host_resident_plan_principal_bytes': code+message+output,
        'gpu_host_file_backed_principal_bytes': message+output+staging,
        'cpu_bounded_resident_principal_bytes': code+message+domain+columns,
        'cpu_bounded_file_backed_principal_bytes': message+domain+columns+staging,
        'cpu_bounded_columns_bytes': columns,
        'cpu_owning_arena_backing_bytes': domain,
        'file_backing_bytes': code,
        'gpu_one_arena_base_half_bytes': domain//2,
        'gpu_half_can_hold_uint32_output': output <= domain//2,
        'transfer_only_scenarios_seconds': {
            str(bandwidth): (domain+message)/(bandwidth*GIB)
            for bandwidth in (8, 16, 32)
        },
    }


def checks():
    a, b = model(16, 918672421), model(17, 3870933765)
    assert a['online_payload_bytes'] == 20854558905-33
    assert a['gpu_fast_streamed_arrays_bytes'] == 32*GIB
    assert a['spectrum_chunks'] == 64 and b['spectrum_chunks'] == 512
    assert b['index_bits'] == 34 and b['count_bits'] == 35
    assert b['spectrum_bytes'] == 128*GIB and b['packed_bytes'] == 34*GIB
    assert not a['gpu_half_can_hold_uint32_output'] and b['gpu_half_can_hold_uint32_output']
    # Half-arena representation is sufficient after range checking at both widths.
    for row in (a, b):
        half_slot = row['extension_slot_bytes']//2
        assert (1 << row['base_bits'])-1 < 1 << (8*half_slot)
        assert 2*row['n']*half_slot == row['spectrum_bytes']
        assert row['count_bits'] == row['index_bits']+1
    # Extending uint32 values/count fields alone loses legal s17 coordinates.
    assert ((1 << 33) & 0xffffffff) != 1 << 33
    assert a['stream_upload_bytes_per_message'] == a['online_payload_bytes']-4
    assert b['gpu_fast_streamed_arrays_bytes'] > 143771*(1 << 20)


def main():
    checks()
    hashes = {}
    refs = {}
    for key, name in REFERENCES.items():
        data = (ROOT/name).read_bytes()
        hashes[name] = hashlib.sha256(data).hexdigest()
        refs[key] = json.loads(data)
        assert refs[key]['status'] == 'passed'
    catalogue = json.loads((ROOT/'pirconfig/configs_catalogue.json').read_text())
    # Find records without relying on the catalogue's outer metadata structure.
    def records(value):
        if isinstance(value, dict):
            if value.get('id') in ('binary-s16-m2-t5', 'binary-s17-m2-t5'):
                yield value
            for child in value.values():
                yield from records(child)
        elif isinstance(value, list):
            for child in value:
                yield from records(child)
    entries = {entry['code']['s']: entry for entry in records(catalogue)}
    rows = []
    for s in (16, 17):
        k = 918672421 if s == 16 else entries[s]['capacity']['k']
        row = model(s, k)
        row['k_status'] = 'verified-preparation' if s == 16 else entries[s]['capacity']['status']
        if s == 16:
            old = refs['gpu_reuse']['encoding']['encodings'][0]['cuda_stages']['owned_device_bytes']
            overhead = old-row['gpu_resident_fast_arrays_bytes']
            assert overhead > 0
            row['gpu_measured_reuse_overhead_bytes'] = overhead
            row['gpu_fast_streamed_projected_owned_bytes'] = old-row['online_payload_bytes']
            row['gpu_measured_compact_owned_bytes'] = refs['gpu_compact']['memory_comparison']['candidate_owned_bytes']
            assert row['gpu_fast_streamed_projected_owned_bytes'] < row['gpu_measured_compact_owned_bytes']
        rows.append(row)
    result = {
        'kind': 'allocation-free-model-not-benchmark',
        'revision': subprocess.check_output(['git','rev-parse','HEAD'],cwd=ROOT,text=True).strip(),
        'source_sha256': {name: hashlib.sha256((ROOT/name).read_bytes()).hexdigest() for name in SOURCES},
        'reference_sha256': hashes,
        'staging_bytes': TILE, 'models': rows,
        'limits': 's17 K is estimated; no s17 code/preprocessing or streamed kernel exists. '
                  'Exclude allocator/CUDA context/page-cache/pinned staging and new wider-plan overhead. '
                  'Bandwidth scenarios are hypothetical, not measured predictions. No budget admission.',
    }
    (HERE/'results.json').write_text(json.dumps(result,indent=2)+'\n')
    for row in rows:
        print('s%d K=%d (%s)' % (row['s'],row['k'],row['k_status']))
        for key in ('gpu_fast_streamed_arrays_bytes','gpu_one_arena_streamed_arrays_bytes',
                    'gpu_host_resident_plan_principal_bytes','gpu_host_file_backed_principal_bytes',
                    'cpu_bounded_resident_principal_bytes','cpu_bounded_file_backed_principal_bytes'):
            print('  %s: %.3f GiB' % (key,row[key]/GIB))


if __name__ == '__main__':
    main()
