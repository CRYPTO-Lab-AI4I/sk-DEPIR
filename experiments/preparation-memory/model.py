#!/usr/bin/env python3
"""Source-based preparation allocation scenarios, explicitly not RSS bounds."""
import hashlib
import json
import math
import subprocess
from pathlib import Path

ROOT=Path(__file__).resolve().parents[2]
HERE=Path(__file__).resolve().parent
GIB=2**30

def pow2(n):
    return 1 << (n-1).bit_length()


def scenario(s,k,status,map_bpe):
    n=1<<(2*s);q=1<<s
    # At most q-1 support exponents are fixed by Frobenius. All others pair.
    factors=(k+min(k,q-1)+1)//2
    headers=24*math.ceil(1.25*factors)
    leaves=16*factors  # 12-byte quadratics occupy the 16-byte class on this host.
    support=8*k
    # For K<N/2, the first overlap-add block of P*G needs an N-point FFT.
    assert 1<k<n//2
    identity=4*n+support+4*(k+1)+4*(n-k)+4*n+8*n
    # Last Newton step uses previous g length N/2 and square backing N-1.
    # polyMulLow trims its view, but the full square backing allocation remains.
    newton=4*n+support+4*(k+1)+4*(n-1)+4*(n-k)+8*n
    prefix=pow2(2*k-1)
    # Root illustration: all leaf references plus two children, output and FFT pair.
    root_fft=pow2(k+1)
    root=4*n+support+headers+leaves+4*(2*k+3)+8*root_fft
    return {'s':s,'k':k,'n':n,'capacity_status':status,
            'support_len_bytes':support,'support_capacity_slack_scenario_bytes':2*k,
            'transpose_bytes':4*n,'membership_bitset_bytes':(n+7)//8,
            'validation_map_estimate_bytes':[math.ceil(k*x) for x in map_bpe],
            'factor_count_ceiling':factors,'factor_headers_125_percent_capacity_bytes':headers,
            'factor_leaf_allocation_scenario_bytes':leaves,'factors_total_scenario_bytes':headers+leaves,
            'root_top_multiply_scenario_bytes':root,
            'identity_array_scenario_bytes':identity,
            'newton_array_scenario_bytes':newton,
            'newton_with_old_g_retained_scenario_bytes':newton+2*n,
            'identity_with_support_capacity_slack_bytes':identity+2*k,
            'prepared_arrays_bytes':12*n+support+4,
            'minimal_online_file_bytes':4*(k+1+n)+25,
            'support_p_g_spectrum_files_bytes':4*k+4*(k+1+n)+25+4*(n-k),
            'packed_word_bytes':(n*s+7)//8,
            'online_prefix_elements':prefix,
            'online_complete_array_payload_bytes':4*(k+1+n)+4*(max(2*prefix,n)+k+n)+4*k+4*n}


def main():
    replica=json.loads((HERE/'replicas-final/results.json').read_text())
    assert replica['status']=='passed'
    maps=[c for c in replica['cases'] if c['kind']=='map']
    ratios=sorted(c['retained_delta_bytes']/c['count'] for c in maps)
    factor=next(c for c in replica['cases'] if c['kind']=='factors')
    assert factor['int_bytes']==8 and factor['slice_header_bytes']==24
    bitset=next(c for c in replica['cases'] if c['kind']=='bitset')
    assert abs(bitset['retained_delta_bytes']-(bitset['domain']+7)//8)<1<<20
    configs=json.loads((ROOT/'pirconfig/configs_catalogue.json').read_text())['configurations']
    rows=[]
    for s in [14,15]:
        config=next(c for c in configs if c['code']['s']==s)
        rows.append(scenario(s,config['capacity']['k'],config['capacity']['status'],ratios))
    c15=next(c for c in configs if c['code']['s']==15)
    mc=next(e['monte_carlo'] for e in c15['capacity']['evidence'] if 'monte_carlo' in e)
    bounds=[math.floor(c15['capacity']['n']*mc['lower_rate']),math.ceil(c15['capacity']['n']*mc['upper_rate'])]
    interval=[scenario(15,k,'monte_carlo_interval_endpoint',ratios) for k in bounds]
    measured=json.loads((ROOT/'experiments/parallel-s14/initial/results.json').read_text())
    phases=json.loads((ROOT/'experiments/parallel-s14/initial/phases.json').read_text())
    prep=next(p for p in phases if p['phase']=='preparation')
    sources=[HERE/'model.py',HERE/'replicas-final/results.json',ROOT/'pirconfig/configs_catalogue.json',
             ROOT/'experiments/parallel-s14/initial/results.json',ROOT/'experiments/parallel-s14/initial/phases.json',
             ROOT/'experiments/parallel-s14/initial/preparation-flat.txt',
             *[ROOT/'encoding/liftrs-shared'/p for p in ['orbit.go','orbit_product.go','orbit_polynomial.go','scaling_experimental.go','scaling_prepare_cpp.go','scaling_matrix_cpp.go','parallel_firstuse_probe_test.go']]]
    output={'kind':'analytical_payload_scenarios_not_measured_s15_rss','revision':subprocess.check_output(['git','rev-parse','HEAD'],cwd=ROOT,text=True).strip(),
            'source_sha256':{str(p.relative_to(ROOT)):hashlib.sha256(p.read_bytes()).hexdigest() for p in sources},
            'assumptions':{'int_bytes':8,'coefficient_bytes':4,'slice_header_bytes':24,'factor_capacity_scenario':1.25,'factor_leaf_bytes':16,
                           'excludes':['GC-retained unreachable objects','allocator/runtime/native metadata','thread stacks','OS pages and other processes'],
                           'map_projection':'Empirical B/entry range at two small map densities; not an exact capacity formula or bound.',
                           'coefficient_lifetimes':'Logically required arrays; compiler liveness and physical RSS may differ. Old reciprocal g shown separately.'},
            'measured_s14':{'whole_process_peak_rss_bytes':measured['process_peak_rss_bytes'],'preparation_allocation_traffic_bytes':prep['allocated_bytes'],'preparation_end_heap_bytes':prep['heap_after_bytes'],'preparation_end_go_sys_bytes':prep['go_sys_bytes'],'preparation_s':prep['elapsed_ns']/1e9},
            'measured_map_bytes_per_entry':ratios,'rows':rows,'s15_capacity_interval':bounds,'s15_interval_rows':interval}
    # Independent arithmetic checks: polynomial identity envelope and byte layout.
    for row in rows+interval:
        assert row['identity_array_scenario_bytes']==20*row['n']+8*row['k']+4
        assert row['newton_array_scenario_bytes']==20*row['n']+8*row['k']
        assert row['online_prefix_elements']==row['n']//2
    (HERE/'results.json').write_text(json.dumps(output,indent=2)+'\n')
    keys=[('Transpose scratch','transpose_bytes'),('Support, length only','support_len_bytes'),('Factor list and leaf scenario','factors_total_scenario_bytes'),('Top root product illustration','root_top_multiply_scenario_bytes'),('P*G identity array scenario','identity_array_scenario_bytes'),('With support capacity slack','identity_with_support_capacity_slack_bytes'),('Last Newton, old g retained','newton_with_old_g_retained_scenario_bytes'),('Cached online array payload','online_complete_array_payload_bytes'),('Code artifacts on disk','support_p_g_spectrum_files_bytes')]
    table=['| Quantity (GiB) | s=14, exact K | s=15, estimated K |','|---|---:|---:|']
    for label,key in keys:table.append('| '+label+' | '+' | '.join(f'{r[key]/GIB:.3f}' for r in rows)+' |')
    (HERE/'tables.md').write_text('\n'.join(table)+'\n')
    print('\n'.join(table))
    print('s15 map estimate GiB',*[round(x/GIB,3) for x in rows[1]['validation_map_estimate_bytes']])
    print('s15 coefficient interval GiB',*[round(r['identity_with_support_capacity_slack_bytes']/GIB,3) for r in interval])

if __name__=='__main__':
    main()
