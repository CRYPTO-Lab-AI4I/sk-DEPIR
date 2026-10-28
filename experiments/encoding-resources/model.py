#!/usr/bin/env python3
"""Static array/access accounting, not a memory or runtime benchmark.

Counts come from the current cached cyclic implementation. Future-width rows
use uint32 base and uint64 extension/index slots when 2*s>32. Candidate payloads
are scheduling scenarios: they exclude allocator/GC/capacity and OS overhead.
"""
import functools,hashlib,json,math,subprocess
from pathlib import Path
ROOT=Path(__file__).resolve().parents[2];HERE=Path(__file__).resolve().parent

def power2(n):return 1 << (n-1).bit_length()
@functools.lru_cache(None)
def tree_coefficients(k):
    """Sum logical polynomial lengths in the exact balanced binary product tree."""
    if k==1:return 2
    return k+1+tree_coefficients(k//2)+tree_coefficients(k-k//2)

def row(c):
    s=c['code']['s'];n=c['capacity']['n'];k=c['capacity']['k']
    b=4;e=4 if s<=16 else 8;d=e
    length=power2(2*k-1)
    assert n==1 << (2*s)
    assert length<=n # Catalogue point capacities remain below N/2.
    # Current 64-bit Go: []int support. Extension/cardinality widening is future work.
    resident=b*(n+1)+e*n+8*k
    prefix=resident+2*b*k+2*e*length
    assembly=resident+2*b*k+(e+b)*n
    lean=b*(k+1)+e*n # Proposed online artifact: P+spectrum, scalar K, no G/support.
    lean_prefix=lean+2*b*k+2*e*length
    lean_assembly=lean+2*b*k+(e+b)*n # Retain canonical output array; no free scatter.
    coeffs=tree_coefficients(k)
    mc=next((x['monte_carlo'] for x in c['capacity']['evidence'] if 'monte_carlo' in x),None)
    ranges=[length]
    if c['capacity']['status']!='verified' and mc:
        ranges=[power2(2*max(1,math.floor(n*mc['lower_rate']))-1),power2(2*math.ceil(n*mc['upper_rate'])-1)]
    packed=lambda count,bits:(count*bits+7)//8
    return {'s':s,'t':c['code']['t'],'n':n,'k':k,'capacity_status':c['capacity']['status'],
        'cached_transform_elements':3*length+2*n,'existing_uncached_transform_elements':3*length+3*n+3*(n//2),'candidate_single_domain_uncached_transform_elements':3*length+3*n,'extension_slot_bytes':e,'prefix_fft_elements':length,'prefix_fft_elements_interval':ranges,
        'input_packed_bytes':packed(k,s),'base_output_packed_bytes':packed(n,s),
        'discovery_array_scenario_bytes':(e+1)*n+20*k+4*(1<<s),'current_resident_array_bytes':resident,'current_principal_buffer_scenario_bytes':max(prefix,assembly),
        'current_prefix_scenario_bytes':prefix,'current_assembly_scenario_bytes':assembly,
        'dense_transform_setup_pair_bytes':e*(n+n//2+2),
        'product_tree_nodes':2*k-1,'product_tree_polynomial_elements':coeffs,
        'product_tree_polynomial_bytes':e*coeffs,'product_tree_header_bytes':72*(2*k-1),
        'tree_roots_support_payload_bytes':e*coeffs+72*(2*k-1)+e*k+8*k,
        'lean_resident_bytes':lean,'lean_prefix_scenario_bytes':lean_prefix,'lean_assembly_scenario_bytes':lean_assembly,'lean_disk_output_principal_bytes_excluding_tiles':max(lean_prefix,lean+2*b*k+e*n),'lean_principal_buffer_scenario_bytes':max(lean_prefix,lean_assembly),
        'aligned_full_binary_cache_bytes':b*(n+1)+e*n+d*k,
        'bitpacked_full_cache_bytes':packed(n+1,s)+packed(n,2*s)+packed(k,2*s),
        'aligned_online_cache_bytes':lean,'bitpacked_online_cache_bytes':packed(k+1,s)+packed(n,2*s),
        'bitpacked_without_spectrum_cache_bytes':packed(n+1,s),
        # One sweep means read+write of the vector once; p is UNMEASURED.
        'transform_io_scenarios':[{'sweeps_per_transform':p,'bytes':2*p*e*(3*length+2*n)} for p in [2,4,8]],
        # Pairwise point product after each forward pair: read 2 arrays, write 1.
        'pointwise_product_stream_bytes':3*e*(length+n),
        'routing_one_partition_level_bytes':e*n+2*n*(4+b)+b*n,
        'routing_record_bytes':8,'routing_wide_intermediate_extra_level_bytes':24*n,'routing_note':'one partition level to buckets spanning <=2^32 positions; uint32 local offsets plus uint32 base values. Earlier multi-level buckets may need uint64 offsets (12-byte pairs), adding 24*N bytes per read/write level.'}

def main():
    # Small independent closed-form checks catch broken tree/rounding accounting.
    assert tree_coefficients(1)==2 and tree_coefficients(2)==7 and tree_coefficients(3)==13
    for j in range(1,12):assert tree_coefficients(2**j)==2**j*(j+3)-1
    layout=json.loads((HERE/'layout.json').read_text())
    assert layout=={'host_int_bytes':8,'product_tree_node_bytes':72}
    configs=json.loads((ROOT/'pirconfig/configs_catalogue.json').read_text())['configurations']
    rows=[row(c) for c in configs if c['code']['s']>=13]
    assert rows[0]['prefix_fft_elements']==rows[0]['n']//2
    assert rows[-1]['prefix_fft_elements']==rows[-1]['n']
    assert rows[-2]['prefix_fft_elements_interval']==[rows[-2]['n']//2,rows[-2]['n']]
    sources=[ROOT/'pirconfig/configs_catalogue.json',HERE/'model.py',HERE/'layout.go',HERE/'layout.json',ROOT/'paper.pdf',ROOT/'experiments/native-encoding/s13/results.json',ROOT/'experiments/cyclic-s13/results.json',*sorted((ROOT/'encoding/liftrs-shared').rglob('*.go')),ROOT/'encoding/liftrs-cpp/transform.hpp',ROOT/'encoding/liftrs-cpp/c_api.cpp',ROOT/'encoding/liftrs-cpp/plan.go',ROOT/'experiments/native-encoding/main.go']
    native=json.loads((ROOT/'experiments/native-encoding/s13/results.json').read_text())['cases']['go-cpp']
    first=json.loads((ROOT/'experiments/cyclic-s13/results.json').read_text())
    result={'measured_anchors':{'s13_native_encoding_s':native['median_encoding_seconds'],'s13_loaded_process_peak_rss_bytes':native['process_peak_rss_bytes'],'s13_preparation_s':first['measurement']['preparation_s'],'s13_first_use_process_peak_rss_bytes':first['process_peak_rss_bytes']},'kind':'analytical_not_measured','revision':subprocess.check_output(['git','rev-parse','HEAD'],cwd=ROOT,text=True).strip(),
        'assumptions':{'base_slot_bytes':4,'host_int_bytes':8,'product_tree_node_bytes':72,'native_metadata_excluded':True,'heap_rss_prediction':False,'io_sweep_counts':'illustrative parameters, not derived pass counts'},
        'source_sha256':{str(p.relative_to(ROOT)):hashlib.sha256(p.read_bytes()).hexdigest() for p in sources},'rows':rows}
    (HERE/'results.json').write_text(json.dumps(result,indent=2)+'\n')
    def gib(x):return f'{x/2**30:,.2f}'
    table=['| s | K evidence | Current arrays resident (GiB) | Current principal-buffer scenario (GiB) | Lean scenario (GiB) | Online packed cache (GiB) | Packed codeword (GiB) |',
           '|---:|---|---:|---:|---:|---:|---:|']
    for r in rows:table.append('| '+str(r['s'])+' | '+r['capacity_status']+' | '+' | '.join(gib(r[k]) for k in ['current_resident_array_bytes','current_principal_buffer_scenario_bytes','lean_principal_buffer_scenario_bytes','bitpacked_online_cache_bytes','base_output_packed_bytes'])+' |')
    table+=['','Array scenarios exclude JSON, map validation, capacities, garbage retention, allocator and OS overhead; they are not RSS bounds. s>=14 is unimplemented; s>=16 needs size-accounting changes, s>=17 wider extension elements.','',
            '| s | Dense transform-setup pair (GiB) | Product-tree coefficients + node headers (GiB) | Full packed cache (GiB) | Transform I/O at p=4 (TiB) |',
            '|---:|---:|---:|---:|---:|']
    for r in rows:table.append(f'| {r["s"]} | {gib(r["dense_transform_setup_pair_bytes"])} | {gib(r["product_tree_polynomial_bytes"]+r["product_tree_header_bytes"])} | {gib(r["bitpacked_full_cache_bytes"])} | {r["transform_io_scenarios"][1]["bytes"]/2**40:,.3f} |')
    (HERE/'tables.md').write_text('\n'.join(table)+'\n')
    print('\n'.join(table))
if __name__=='__main__':main()
