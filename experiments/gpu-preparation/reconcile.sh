#!/bin/bash
# Lightweight completed-run review; large cached payloads are not rehashed.
set -euo pipefail
cd "$(dirname "$0")/../.."
python3 - "$@" <<'PY'
from pathlib import Path
import json,hashlib,runpy,sys
if len(sys.argv)!=2 or Path(sys.argv[1]).name!=sys.argv[1]:raise SystemExit('Usage: reconcile.sh RUN_NAME')
root=Path.cwd();out=root/'experiments/gpu-preparation/runs'/sys.argv[1];r=json.loads((out/'results.json').read_text());cache=Path(r['cache'])
sources=runpy.run_path(str(root/'experiments/source_evidence.py'))['verify_sources'](root,r['source_sha256'],r['revision'])
def sha(p):
 h=hashlib.sha256()
 with p.open('rb') as f:
  for block in iter(lambda:f.read(1<<20),b''):h.update(block)
 return h.hexdigest()
assert r['status']=='passed' and not r.get('error') and not r['validation']
for name,digest in r['evidence_sha256'].items():assert sha(out/name)==digest,name
assert all(c['exit_code']==0 and c['guard'] is None for c in r['commands'])
for name in ('libgpu-matrix.so','preparation.test','provenance.json','root-metadata.json','online-metadata.json'):assert sha(cache/name)==r['cache_sha256'][name],name
meta=json.loads((cache/'online-metadata.json').read_text());pmeta=json.loads((cache/'root-metadata.json').read_text());assert meta['identity_verified']
assert meta['s']==16 and meta['k']==918672421 and meta['n']==1<<32
for name,digest,size in [('root.bin',pmeta['sha256'],pmeta['bytes']),('online.bin',meta['sha256'],meta['bytes']),('g.bin',meta['generator']['sha256'],meta['generator']['bytes'])]:
 assert r['cache_sha256'][name]==digest and (cache/name).stat().st_size==size
summary=json.loads((out/'worker/gpu-summary.json').read_text());assert summary['owned_after_close']==0 and summary['preparation_only']
assert r['support_sha256']==pmeta['support_sha256']=='021446aa0d2764f4020e989ee54f3a99c63c7d9a407f5babbcb333919a2854ad'
result={'status':'passed','scope':'Sources, sealed output evidence, small cache files, metadata and artifact sizes verified; large artifact digests are recorded controller results, not freshly recomputed.','script_sha256':sha(root/'experiments/gpu-preparation/reconcile.sh'),'sources_verified':len(sources),'sealed_evidence_files_verified':len(r['evidence_sha256']),'preparation_identity_verified':True,'peak_sampled_rss_bytes':max(c['peak_rss_bytes'] for c in r['commands']),'retained_cache_bytes':sum(p.stat().st_size for p in cache.iterdir() if p.is_file() and not p.is_symlink()),'gpu_calls':summary['calls'],'gpu_products':summary['product_calls'],'gpu_call_seconds':summary['stage_seconds'][7]}
(out/'reconciliation.json').write_text(json.dumps(result,indent=2)+'\n');print(json.dumps(result,indent=2))
PY
