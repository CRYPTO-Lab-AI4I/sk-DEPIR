#!/usr/bin/env python3
"""Verify retained integration evidence and regenerate the measured summary table."""
import hashlib,json,statistics,sys
from pathlib import Path
ROOT=Path(__file__).resolve().parents[2]
HERE=Path(__file__).resolve().parent
sys.path.insert(0,str(HERE.parent))
from source_evidence import verify_sources

def sha(p):
    h=hashlib.sha256()
    with p.open('rb') as stream:
        for block in iter(lambda:stream.read(1<<20),b''):h.update(block)
    return h.hexdigest()

def main():
    run=HERE/(sys.argv[1] if len(sys.argv)>1 else 's13')
    data=json.loads((run/'results.json').read_text())
    assert data['status']=='completed'
    verify_sources(ROOT,data['source_sha256'],data['revision'])
    assert sha(ROOT/data['binary']['path'])==data['binary']['sha256']
    assert sha(Path(data['cache_directory'])/'cyclic.json')==data['cache_metadata']['sha256']
    rows=['| Backend | Cache load (s) | Native init (s) | Encoding calls (s) | Median (s) | Whole-process peak RSS (GiB) |','|---|---:|---:|---|---:|---:|']
    for label,case in data['cases'].items():
        assert case['status']=='completed'
        for name,digest in case['evidence_sha256'].items():assert sha(run/label/name)==digest,name
        times=case['encoding_seconds']
        assert len(times)==data['policy']['repeats']
        assert statistics.median(times)==case['median_encoding_seconds']
        verification=case['verification']
        assert verification['systematic_positions_checked']==data['cache_metadata']['k']
        assert verification['curve_checks']==16
        if label=='go-cpp':
            for e in case['events']:
                if e['event']=='encode_stats':
                    s=e['native'];assert s['native_calls']>0 and s['transformed_elements']>0 and s['interface_message_copy_bytes']==0
                    assert s['native_call_ns']>=s['native_kernel_ns']>0
            if data['policy']['native_sample']:assert case['native_sample_returncode']==0
        phases=case['phases']
        rows.append(f'| {label} | {phases["cache_load"]["elapsed_ns"]/1e9:.3f} | {phases.get("native_init",{}).get("elapsed_ns",0)/1e9:.6f} | '+', '.join(f'{x:.3f}' for x in times)+f' | {statistics.median(times):.3f} | {case["process_peak_rss_bytes"]/2**30:.3f} |')
    if 'go' in data['cases'] and 'go-cpp' in data['cases']:
        for key in ['canonical_sha256','packed_sha256']:
            assert data['cases']['go']['verification'][key]==data['cases']['go-cpp']['verification'][key]
    (run/'table.md').write_text('\n'.join(rows)+'\n')
    print('\n'.join(rows))
    if 'speedup' in data:print(f'Full-encoding median speedup: {data["speedup"]:.4f}x')
if __name__=='__main__':main()
