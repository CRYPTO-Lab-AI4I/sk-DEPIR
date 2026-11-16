#!/usr/bin/env python3
"""Reconcile completed scalar validation evidence without scheduler/process calls."""
import argparse
import hashlib
import json
from pathlib import Path
import subprocess


def reconcile(out, require_success=True, source_overrides=None):
    """Verify sealed hashes; forensic failed-run review must opt out of success."""
    record=json.loads((out/'results.json').read_text())
    if require_success: assert record['status']=='passed'
    # Evidence logs can be gigabytes even when code/cache metadata is small.
    # Keep verification byte-identical without retaining a whole file in RAM.
    def sha(path):
        digest=hashlib.sha256()
        with Path(path).open('rb')as stream:
            for block in iter(lambda:stream.read(1<<20),b''):digest.update(block)
        return digest.hexdigest()
    for name,digest in record['evidence_sha256'].items():assert sha(out/name)==digest,name
    for name,digest in record['cache_sha256'].items():assert sha(Path(record['cache'])/name)==digest,name
    # Explicit review-only snapshots handle a documented dirty job-start tree.
    # Every override must still match the producer's original sealed digest.
    overrides=source_overrides or {}
    assert set(overrides)<=set(record['source_sha256'])
    for name,path in overrides.items():assert sha(path)==record['source_sha256'][name],name
    names=[name for name in record['source_sha256']if name not in overrides]
    requests=''.join(f"{record['revision']}:{name}\n" for name in names)
    data=subprocess.check_output(['git','cat-file','--batch'],input=requests.encode())
    position=0
    for name in names:
        end=data.index(b'\n',position)
        header=data[position:end].split()
        assert header[1]==b'blob',name
        size=int(header[2]);position=end+1
        assert hashlib.sha256(data[position:position+size]).hexdigest()==record['source_sha256'][name],name
        position+=size+1
    if require_success: assert all(c['exit_code']==0 and not c['guard'] for c in record['commands'])
    result={'status':record['status'],'job_id':record['job_id'],'revision':record['revision'],
            'source_files_verified':len(record['source_sha256']),
            'source_override_files_verified':sorted(overrides),'evidence_files_verified':len(record['evidence_sha256']),
            'cache_files_verified':len(record['cache_sha256']),
            'tests':['TestWideOrbitArithmetic','TestTableFreeFields'],
            'commands':[{k:c[k] for k in ('label','seconds','peak_rss_bytes','exit_code')} for c in record['commands']],
            'scope':'scalar correctness; no transform/preprocessing/encoding capability at s17'}
    return result


def main():
    parser=argparse.ArgumentParser(description=__doc__)
    parser.add_argument('name')
    args=parser.parse_args()
    out=Path(__file__).resolve().parent/'runs'/args.name
    result=reconcile(out)
    (out/'analysis.json').write_text(json.dumps(result,indent=2)+'\n')
    print('passed:',result['source_files_verified'],'source files; evidence/cache hashes reconciled')


if __name__=='__main__':main()
