#!/usr/bin/env python3
"""Reconcile completed buffer reuse validation, matched s10 and full s16 memory."""
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
    assert h['sha'](Path(r['old_reuse_library']['path']))==r['old_reuse_library']['sha256']
    matrix=[json.loads(line) for line in (out/'matrix-regression.log').read_text().splitlines()]
    assert len(matrix)==80 and all(row.get('match',True) for row in matrix)
    assert h['sha'](Path(r['matrix_reference']['path']))==r['matrix_reference']['sha256']
    summary=h['verify_phases'](out/'encoding',16)
    assert h['verify_small'](out)==r['small_checks']
    baseline=json.loads((ROOT/'experiments/cuda-s16-encoding/runs/encoding-01/results.json').read_text())
    assert h['verify_memory'](r,baseline)==r['memory_comparison']
    assert 'ERROR SUMMARY: 0 errors' in (out/'memcheck.log').read_text()
    cpu=r['cpu_reference']['encoding']
    assert summary['canonical_sha256']==cpu['canonical_sha256'] and summary['packed_sha256']==cpu['packed_sha256']
    assert r['cache_sha256']['encoding/word-packed.bin']==summary['packed_sha256']
    check_packed_metadata(summary,json.loads((cache/'encoding/word-metadata.json').read_text()),(cache/'encoding/word-packed.bin').stat().st_size)
    r_summary={'status':'passed','job_id':r['job_id'],'sources_verified':len(sources),'packed_payload_digest_recomputed':not recorded_cache,'encoding':summary,'peak_sampled_rss_bytes':max(c['peak_rss_bytes'] for c in r['commands']),
               'historical_cpu_encode_seconds':cpu['phase_seconds']['encode-0'],'memory_comparison':r['memory_comparison'],'matched_s10':{k:r[k] for k in ('matched-baseline','matched-reuse')},
               'limits':('Packed payload digest is the completed controller result, not a fresh bulk rehash. ' if recorded_cache else '')+'Two profiled same-message encodings with a reused GPU session, full retained CPU digest agreement, all systematic positions, sampled independent curves and packed readback. Full-size CPU/default-GPU timing is historical; s10 default/reuse timings share an allocation. Not cryptographic/application execution or TRL5 acceptance.'}
    (out/'analysis.json').write_text(json.dumps(r_summary,indent=2)+'\n');print(json.dumps(r_summary,indent=2))

def check_packed_metadata(summary,meta,size):
    assert meta['packed_bytes']==size==(summary['n']*summary['s']+7)//8
    assert meta['sha256']==summary['packed_sha256'] and meta['canonical_sha256']==summary['canonical_sha256'] and meta['seed']==71

if __name__=='__main__':main()
