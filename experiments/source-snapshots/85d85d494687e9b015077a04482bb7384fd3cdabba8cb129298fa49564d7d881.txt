#!/usr/bin/env python3
"""Verify CPU scratch/output reuse memory and output evidence without re-encoding."""
import json,runpy,sys
from pathlib import Path
ROOT=Path(__file__).resolve().parents[2]
HERE=Path(__file__).resolve().parent
sys.path.insert(0,str(ROOT/'experiments'))
from source_evidence import verify_sources

def main():
    if len(sys.argv) not in (2,3) or (len(sys.argv)==3 and sys.argv[2]!='--recorded-cache') or Path(sys.argv[1]).name!=sys.argv[1]:raise SystemExit('Usage: analyze.py RUN_NAME [--recorded-cache]')
    recorded_cache=len(sys.argv)==3
    out=HERE/'runs'/sys.argv[1];r=json.loads((out/'results.json').read_text())
    if r['status']!='passed':print(json.dumps({'status':r['status'],'error':r.get('error')},indent=2));return
    sources=verify_sources(ROOT,r['source_sha256'],r['revision']);h=runpy.run_path(str(HERE/'run.py'));cache=Path(r['cache'])
    reference=Path(r['input']['reference']);h['validate_metadata'](json.loads((reference/'online-metadata.json').read_text()),(reference/'online.bin').stat().st_size)
    assert r['input']['sha256']==h['EXPECTED']
    for path,digest in r['evidence_sha256'].items():assert h['sha'](out/path)==digest,path
    for path,digest in r['cache_sha256'].items():
        if recorded_cache and path=='encoding/word-packed.bin':continue
        assert h['sha'](cache/path)==digest,path
    assert all(c['exit_code']==0 and c['guard'] is None for c in r['commands'])
    summary=h['verify_phases'](out/'encoding',16,True)
    assert r['cache_sha256']['encoding/word-packed.bin']==summary['packed_sha256']
    check_packed_metadata(summary,json.loads((cache/'encoding/word-metadata.json').read_text()),(cache/'encoding/word-packed.bin').stat().st_size)
    for key in ('canonical_sha256','packed_sha256'):assert summary[key]==r['cpu_reference']['encoding'][key]
    for mode in ('matched-baseline','matched-reuse'):assert h['verify_phases'](out/mode,10,mode=='matched-reuse')==r[mode]
    assert r['matched-baseline']['canonical_sha256']==r['matched-reuse']['canonical_sha256']
    native=next(p for p in json.loads((out/'encoding/phases.json').read_text()) if p['phase']=='encode-0')['native_stages']
    assert native['calls']==5 and native['boundary_calls']==1
    r_summary={'status':'passed','job_id':r['job_id'],'sources_verified':len(sources),'packed_payload_digest_recomputed':not recorded_cache,'native_stage_seconds':{key:value/1e9 for key,value in native.items() if key.endswith('_ns')},'encoding':summary,'peak_sampled_rss_bytes':max(c['peak_rss_bytes'] for c in r['commands']),
               'principal_memory_bytes':r['principal_memory_bytes'],'matched_s10':{mode:r[mode] for mode in ('matched-baseline','matched-reuse')},'limits':('Packed payload digest is the completed controller result, not a fresh bulk rehash. ' if recorded_cache else '')+'Two profiled cached mathematical encodings, owning-output tests, ASAN/UBSAN and matched s10 buffer schedules, all systematic positions, sampled independent curves and packed readback. Full s16 canonical/packed digests match retained CPU/GPU reference; timing comparison is historical. No cryptographic/application execution or TRL5 acceptance.'}
    (out/'analysis.json').write_text(json.dumps(r_summary,indent=2)+'\n');print(json.dumps(r_summary,indent=2))

def check_packed_metadata(summary,meta,size):
    assert meta['packed_bytes']==size==(summary['n']*summary['s']+7)//8
    assert meta['sha256']==summary['packed_sha256'] and meta['canonical_sha256']==summary['canonical_sha256'] and meta['seed']==71

if __name__=='__main__':main()
