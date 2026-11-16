#!/usr/bin/env python3
"""Explicit dense-layout capacity models, not s18 implementation/admission."""
import argparse
import json

GIB = 1 << 30


def model(s, k):
    n = 1 << (2*s)
    if s not in (17, 18) or not 0 < k <= n//2:
        raise ValueError('bounded model requires s17/18 and exact0<K<=N/2')
    files = dict(support=8*k, root=4*(k+1), code=53+4*(k+1)+8*n,
                 generator=4*(n-k))
    return dict(s=s, k=k, n=n, file_bytes=files,
                canonical_retained_bytes=sum(files.values()),
                discovery_grid_bytes=8*n,
                discovery_principal_host_bytes=16*k+8*n+2*(n//8),
                final_retained_host_bytes=8*k+12*n+4,
                identity_result_bytes=4*n,
                device_product_arrays_bytes=24*n,
                device_capacity_floor_140_4gib=(24*n+int(140.4*GIB)-1)//int(140.4*GIB),
                device_capacity_floor_112gib_cap=(24*n+112*GIB-1)//(112*GIB),
                power_of_two_device_count_for_112gib_cap=1 << (((24*n+112*GIB-1)//(112*GIB))-1).bit_length(),
                two_each_phase_snapshot_upper_bytes=2*((n+7)//8+4*(k+1)+4*(n-k)+3*32768*4+64*4),
                measured_peak_host_bytes=None, runtime_seconds=None,
                limits=['Principal live-array/file formulas; not peak RSS, driver or filesystem use.',
                        'P/G products, Newton temporaries, GC, staging and cache/headroom are additional.',
                        'Multiple-node GPU and distributed host implementation absent.',
                        'No new material memory-time tradeoff selected or measured.'])


def write_model(path, s, k):
    result = model(s, k)
    path.write_text(json.dumps(result, indent=2)+'\n')
    return result


if __name__ == '__main__':
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('k', type=int)
    args = parser.parse_args()
    print(json.dumps(model(18, args.k), indent=2))
