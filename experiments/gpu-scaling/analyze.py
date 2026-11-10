#!/usr/bin/env python3
"""Verify bounded GPU scaling evidence; do not forecast a full preparation run."""
import json, runpy, statistics, sys
from pathlib import Path
ROOT=Path(__file__).resolve().parents[2]
HERE=Path(__file__).resolve().parent
sys.path.insert(0,str(ROOT/'experiments'))
from source_evidence import verify_sources

def main():
    if len(sys.argv)!=2 or Path(sys.argv[1]).name!=sys.argv[1]:raise SystemExit('Usage: analyze.py RUN_NAME')
    out=HERE/'runs'/sys.argv[1]
    record=json.loads((out/'results.json').read_text());assert record['status']=='passed',record.get('error')
    sha=runpy.run_path(str(ROOT/'experiments/hpc-cached-s15/run.py'))['sha']
    sources=verify_sources(ROOT,record['source_sha256'],record['revision'])
    for path,digest in record['evidence_sha256'].items():assert sha(out/path)==digest,path
    cache=Path(record['cache'])
    for path,digest in record['cache_sha256'].items():assert sha(cache/path)==digest,path
    assert sha(Path(record['old_library']))==record['old_library_sha256']
    assert all(r['exit_code']==0 and not r['guard'] and r['peak_rss_bytes']<=12*2**30 for r in record['commands'])
    metadata=json.loads((out/'metadata.json').read_text())
    assert metadata['domain']==1<<32 and metadata['row_levels']==17
    assert metadata['device_before']==metadata['device_after']
    assert all(metadata[k] for k in ('full_plan_rejected_by_budget','malformed_rejected','underfunded_rejected'))
    assert 'ERROR SUMMARY: 0 errors' in (out/'memcheck.log').read_text()
    products=json.loads((out/'product-checks.json').read_text())
    assert products['cases_checked']==100 and products['old_transform_library']
    rows=json.loads((out/'measurements.json').read_text());assert [r['dimension'] for r in rows]==record['policy']['dimensions']
    summary=[]
    for r in rows:
        assert r['match'] and r['n']==1<<r['dimension'] and r['owned_device_bytes']<=record['policy']['device_cap_mib']*2**20
        for key in ('forward_sha256','product_sha256'):assert len(r[key])==64
        row={'dimension':r['dimension'],'n':r['n'],'owned_device_bytes':r['owned_device_bytes']}
        for name,key,cpu in [('forward','gpu_forward','cpu_forward_seconds'),('inverse','gpu_inverse','cpu_inverse_seconds'),('product','gpu_products','cpu_product_seconds')]:
            values=r[key];assert len(values)==2
            for stats in values:assert len(stats)==8 and min(stats)>=0 and sum(stats[:7])<=stats[7]+0.01
            median=statistics.median(stats[7] for stats in values)
            row[name]={'gpu_seconds':median,'cpu_seconds':r[cpu],'cpu_over_gpu':r[cpu]/median}
        summary.append(row)
    assert json.loads((out/'closed.json').read_text())['owned_device_bytes']==0
    result={'status':'passed','sources_verified':len(sources),'job_id':record['job_id'],'metadata':metadata,'measurements':summary,
            'sampled_process_peak_bytes':max(r['peak_rss_bytes'] for r in record['commands']),
            'cache_bytes':sum(p.stat().st_size for p in cache.rglob('*') if p.is_file()),
            'full_domain_array_payload_bytes':12*(1<<32),
            'limits':'One data-center GPU; profiled transfer-inclusive operations. CPU product includes Go pointwise multiplication. Metadata validated through2^32; execution measured at the recorded dimensions. No full s16 runtime forecast.'}
    (out/'analysis.json').write_text(json.dumps(result,indent=2)+'\n');print(json.dumps(result,indent=2))

if __name__=='__main__':main()
