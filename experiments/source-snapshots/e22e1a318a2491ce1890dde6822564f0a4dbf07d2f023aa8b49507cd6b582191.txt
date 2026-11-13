#!/usr/bin/env python3
"""Reconcile direct GPU packed output, sink bounds and measured memory/I/O costs."""
import json
import runpy
import sys
from pathlib import Path
ROOT=Path(__file__).resolve().parents[2]
HERE=Path(__file__).resolve().parent
sys.path.insert(0,str(ROOT/'experiments'))
from source_evidence import verify_sources


def main():
    if len(sys.argv) not in (2,3) or (len(sys.argv)==3 and sys.argv[2]!='--recorded-cache') or Path(sys.argv[1]).name!=sys.argv[1]:raise SystemExit('Usage: analyze.py RUN_NAME [--recorded-cache]')
    recorded=len(sys.argv)==3;out=HERE/'runs'/sys.argv[1];r=json.loads((out/'results.json').read_text())
    if r['status']!='passed':print(json.dumps({'status':r['status'],'error':r.get('error')},indent=2));return
    sources=verify_sources(ROOT,r['source_sha256'],r['revision']);h=runpy.run_path(str(HERE/'run.py'));cache=Path(r['cache'])
    ref=Path(r['input']['reference']);h['validate_metadata'](json.loads((ref/'online-metadata.json').read_text()),(ref/'online.bin').stat().st_size)
    for name,digest in r['evidence_sha256'].items():assert h['sha'](out/name)==digest,name
    for name,digest in r['cache_sha256'].items():
        if recorded and name=='encoding/word-packed.bin':continue
        assert h['sha'](cache/name)==digest,name
    assert all(c['exit_code']==0 and c['guard'] is None for c in r['commands'])
    for key in ('old_reuse_library','old_packed_library','matrix_reference'):
        assert h['sha'](Path(r[key]['path']))==r[key]['sha256']
    matrix=[json.loads(line) for line in (out/'matrix-regression.log').read_text().splitlines()];assert len(matrix)==80 and all(row.get('match',True) for row in matrix)
    streamed=r['policy']['memory_mode']=='streamed';mode=r['policy']['output_mode']
    assert h['verify_small'](out,streamed)==r['small_checks']
    assert 'ERROR SUMMARY: 0 errors' in (out/'memcheck.log').read_text()
    if streamed:
        for name in ('boundaries','boundary-memcheck'):
            records=json.loads((out/name/'boundaries.json').read_text());assert [x['tile_elements'] for x in records]==[1,64,255,256] and all(x['matches']==3 for x in records)
        assert 'ERROR SUMMARY: 0 errors' in (out/'boundary-memcheck.log').read_text()
    if mode=='packed':
        for name in ('packed-tests','packed-memcheck'):
            records=json.loads((out/name/'packed-checks.json').read_text());assert len(records)==7 and all(x['matches']==3 and x['writer_errors'] and x['mixed_owning'] and x['closed'] for x in records)
            assert json.loads((out/name/'parity-check.json').read_text())=={'s':8,'chunks':16,'modes':2,'messages':4,'passed':True}
        assert 'ERROR SUMMARY: 0 errors' in (out/'packed-memcheck.log').read_text()
    summary=h['verify_packed'](out/'encoding',16,mode,streamed);assert summary==r['encoding']
    for name,selected in [('matched-baseline','owning'),('matched-candidate',mode)]:assert h['verify_packed'](out/name,14,selected,streamed)==r[name]
    for key in ('canonical_sha256','packed_sha256'):assert summary[key]==r['cpu_reference']['encoding'][key] and r['matched-baseline'][key]==r['matched-candidate'][key]
    meta=json.loads((cache/'encoding/word-metadata.json').read_text());assert meta['sha256']==summary['packed_sha256']==r['cache_sha256']['encoding/word-packed.bin'] and meta['canonical_sha256']==summary['canonical_sha256'] and meta['seed']==71
    assert meta['packed_bytes']==(cache/'encoding/word-packed.bin').stat().st_size==1<<33
    baseline=json.loads((ROOT/'experiments/cuda-buffer-reuse/runs/reuse-01/results.json').read_text());assert h['verify_memory'](r,baseline,streamed)==r['memory_comparison']
    progress=[json.loads(line) for line in (out/'progress.jsonl').read_text().splitlines()]
    phase_peaks={}
    for row in progress:
        phase=row['worker'].get('phase')
        if row['command']=='encoding' and phase:phase_peaks[phase]=max(phase_peaks.get(phase,0),row['current_rss_bytes'])
    result={'status':'passed','job_id':r['job_id'],'sources_verified':len(sources),'packed_payload_digest_recomputed':not recorded,'encoding':summary,'matched_s14':{key:r[key] for key in ('matched-baseline','matched-candidate')},'memory_comparison':r['memory_comparison'],'phase_sampled_rss_bytes':phase_peaks,'peak_sampled_rss_bytes':max(c['peak_rss_bytes'] for c in r['commands']),'tracked_file_cache_peak_bytes':max(sum(x['tracked_file_cache_bytes'].values()) for x in progress),'rss_plus_tracked_cache_peak_bytes':max(x['rss_bytes']+sum(x['tracked_file_cache_bytes'].values()) for x in progress),'limits':'Two complete packed files per mode, same-input repeat overwrites within new-cache budget. Fresh all-systematic/curves/canonical and packed hashes/readback through separately mapped packed artifact; verifier RSS/cache is separate. Full owning reference comparisons historical; matched s14 output pipelines include packing/writing/hash. No full s17/host-backed-transform/CPU-output claim.'}
    (out/'analysis.json').write_text(json.dumps(result,indent=2)+'\n');print(json.dumps(result,indent=2))

if __name__=='__main__':main()
