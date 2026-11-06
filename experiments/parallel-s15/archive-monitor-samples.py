#!/usr/bin/env python3
"""Losslessly archive monitor samples; retain raw copies outside Git."""
import gzip
import hashlib
import json
import shutil
from pathlib import Path
HERE=Path(__file__).resolve().parent
ROOT=HERE.parents[1]
manifest=HERE/'monitor-sample-archive.json'
entries=json.loads(manifest.read_text()) if manifest.exists() else {}
for folder in sorted(HERE.glob('guard-*')):
    if not folder.is_dir():continue
    for sample in sorted(folder.glob('*-native.txt')):
        raw=sample.read_bytes();digest=hashlib.sha256(raw).hexdigest()
        retained=ROOT/'.cache/liftrs/parallel-s15/monitor-native-samples'/folder.name/sample.name
        retained.parent.mkdir(parents=True,exist_ok=True);shutil.copyfile(sample,retained)
        packed=sample.with_suffix(sample.suffix+'.gz');packed.write_bytes(gzip.compress(raw,mtime=0))
        assert gzip.decompress(packed.read_bytes())==raw
        entries[str(sample.relative_to(HERE))]={'raw_sha256':digest,'gzip':str(packed.relative_to(HERE)),
            'gzip_sha256':hashlib.sha256(packed.read_bytes()).hexdigest(),'retained_raw':str(retained.relative_to(ROOT))}
        sample.unlink()
for entry in entries.values():
    assert hashlib.sha256(gzip.decompress((HERE/entry['gzip']).read_bytes())).hexdigest()==entry['raw_sha256']
manifest.write_text(json.dumps(entries,indent=2)+'\n')
print('Verified archived monitor samples:',len(entries))
