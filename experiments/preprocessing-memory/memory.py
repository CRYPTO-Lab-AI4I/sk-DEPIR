#!/usr/bin/env python3
"""Final-stage principal arrays only; not an admission or peak-RSS estimate."""
import json
import runpy
from pathlib import Path
BASE=runpy.run_path(str(Path(__file__).resolve().parents[1]/'s18-preprocessing-admission/memory.py'))['model']

def model(s,k,tile_bytes=1<<20):
    if not 8<=tile_bytes<=64<<20:
        raise ValueError('prototype tile cap outside8bytes..64MiB')
    base=BASE(s,k);n=base['n']
    resident=base['final_retained_host_bytes']
    # Root checkpoints alias P; counting them again would invent an extra4K.
    return dict(s=s,k=k,n=n,resident_final_array_bytes=resident,
                spectrum_array_bytes=8*n,stream_tile_bytes=tile_bytes,
                streamed_final_array_bytes=resident-8*n+tile_bytes,
                saved_final_array_bytes=8*n-tile_bytes,
                caller_support_bytes=8*k,
                support_release_candidate_final_bytes=4*n+4+tile_bytes,
                identity_stage_principal_bytes=8*k+8*n+4,
                discovery_principal_bytes=base['discovery_principal_host_bytes'],
                device_product_array_bytes=base['device_product_arrays_bytes'],
                additional_transforms=0,additional_canonical_file_bytes=0,
                runtime_penalty_seconds=None,measured_peak_rss_bytes=None,
                scope='Source-derived final-stage model. Caller support release not implemented; discovery/products unchanged. Runtime, cache and earlier temporaries additional.')

def check():
    for s,k in [(17,3876566203),(18,16300262095)]:
        r=model(s,k)
        assert r['resident_final_array_bytes']-r['streamed_final_array_bytes']==8*r['n']-(1<<20)
        assert r['streamed_final_array_bytes']-8*k==r['support_release_candidate_final_bytes']
        assert r['additional_transforms']==r['additional_canonical_file_bytes']==0
    assert model(18,16300262095)['spectrum_array_bytes']==512<<30
    try:model(18,16300262095,7)
    except ValueError:pass
    else:raise AssertionError('invalid tile admitted')

if __name__=='__main__':
    check()
    print(json.dumps([model(17,3876566203),model(18,16300262095)],indent=2))
