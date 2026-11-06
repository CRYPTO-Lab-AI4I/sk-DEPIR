#!/usr/bin/env python3
"""Verify preparation evidence and summarize diagnostic memory checkpoints."""
import gzip
import hashlib
import json
import shutil
import sys
from pathlib import Path

ROOT=Path(__file__).resolve().parents[2]
HERE=Path(__file__).resolve().parent
sys.path.insert(0,str(ROOT/'experiments'))
from source_evidence import verify_sources

def sha(path):
    h=hashlib.sha256()
    with path.open('rb') as f:
        for data in iter(lambda:f.read(1<<20),b''):h.update(data)
    return h.hexdigest()

def analyze(name):
    out=HERE/name;record=json.loads((out/'results.json').read_text())
    assert record['status'] in ['passed','not-started','stopped'],record['status']
    result={'status':record['status'],'sources_resolved':verify_sources(ROOT,record['source_sha256'],record['revision'])}
    if 'binary' in record:assert sha(ROOT/record['binary']['path'])==record['binary']['sha256']
    if record['status']=='not-started':
        result['reason']=record.get('admission_failure','not admitted')
        (out/'analysis.json').write_text(json.dumps(result,indent=2)+'\n');return
    archive={}
    for name,expected in record['evidence_sha256'].items():
        path=out/name
        if name.endswith('-native.txt'):
            packed=out/(name+'.gz')
            original=(ROOT/record['binary']['path']).parent/'native-samples'/out.name/name
            if path.exists():
                raw=path.read_bytes();assert hashlib.sha256(raw).hexdigest()==expected
                original.parent.mkdir(parents=True,exist_ok=True);shutil.copyfile(path,original)
                packed.write_bytes(gzip.compress(raw,mtime=0));path.unlink()
            assert hashlib.sha256(gzip.decompress(packed.read_bytes())).hexdigest()==expected
            archive[name]={'gzip':packed.name,'gzip_sha256':sha(packed),'original_sha256':expected,'retained_original':str(original.relative_to(ROOT))}
        else:assert sha(path)==expected,name
    for artifact in record.get('input_artifacts',{}).values():
        assert sha(ROOT/artifact['path'])==artifact['sha256']
    if record['status']=='stopped':
        result.update(partial_only=True,reason=record.get('stop_reason'),wall_s=record['wall_s'],peak_process_rss_gib=record['process_peak_rss_bytes']/2**30,native_sample_archive=archive)
        (out/'analysis.json').write_text(json.dumps(result,indent=2)+'\n')
        print(json.dumps({k:v for k,v in result.items() if k not in ['sources_resolved','native_sample_archive']},indent=2));return
    phases=record['phases'];prep=next(p for p in phases if p['phase']=='preparation')
    verify=next(p for p in phases if p['phase']=='verify-constants');assert verify['all_constants_match']
    b,a=prep['native_before'],prep['native_after']
    clocks={key:a[key]-b[key] for key in b}
    points=json.loads((out/'checkpoints.json').read_text())
    assert points and points[-1]['stage']=='complete' and points[-1]['workspace_bytes']==0
    assert any(p['stage'].startswith('active-') and p['workspace_bytes']>0 for p in points)
    samples=[json.loads(line) for line in (out/'heap-samples.jsonl').read_text().splitlines()]
    if not record['smoke']:
        assert prep['k']==50976325 and prep['n']==1<<28
        assert verify['online_sha256']=='88d5ff39c5df8bdd4aa0f85fc19833ae673df95365b4fd862bb83936e5a0adf8'
        assert verify['generator_sha256']=='4b9461c0dc7f7fb54389f8217f9ab6e72977e9613a99d10a835f6bb0c22cc7da'
    result.update(native_sample_archive=archive,preparation_s=clocks['steady_ns']/1e9,
                  go_preparation_s=prep['go_elapsed_ns']/1e9,process_cpu_s=clocks['process_cpu_ns']/1e9,
                  clock_deltas_ns=clocks,peak_process_rss_gib=record['process_peak_rss_bytes']/2**30,
                  peak_checkpoint_live_heap_gib=max(p['live_heap_after_gc'] for p in points)/2**30,
                  peak_observed_allocated_heap_gib=max([p['heap_before_gc'] for p in points]+[p['heap_alloc'] for p in samples])/2**30,
                  allocation_traffic_gib=prep['allocation_traffic_bytes']/2**30,
                  checkpoint_gc_s=sum(p['gc_ns'] for p in points)/1e9,
                  constant_checks=verify)
    (out/'analysis.json').write_text(json.dumps(result,indent=2)+'\n')
    print(json.dumps({k:v for k,v in result.items() if k not in ['sources_resolved','native_sample_archive']},indent=2))

if __name__=='__main__':analyze(sys.argv[1] if len(sys.argv)>1 else 's14')
