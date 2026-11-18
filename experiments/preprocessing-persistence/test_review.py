#!/usr/bin/env python3
"""Reviewer contract regressions without large files or scheduler/GPU work."""
import copy
import hashlib
import json
from pathlib import Path
import runpy
import tempfile
import unittest
HELP=runpy.run_path(str(Path(__file__).with_name('run.py')))


def fixture(root):
    rows=[];size=128<<10
    modes=[(0,1),(4<<20,1),(4<<20,2),(0,2),(0,4),(4<<20,4)];order=modes+list(reversed(modes))
    for trial,(buffer,writers)in enumerate(order):
        files=[]
        for shard in range(4):
            data=bytes([shard])*size;record=dict(bytes=size,sha256=hashlib.sha256(data).hexdigest())
            stage=f'io-trial-{trial:02d}-file-{shard}';dest=root/stage;dest.mkdir()
            (dest/'manifest.json').write_text(json.dumps(dict(stage=stage,files={'payload.bin':record})))
            if trial==11:(dest/'payload.bin').write_bytes(data)
            chunk=min(size,buffer or 65536)
            stats=dict(bytes=size,calls=size//chunk,sizes={str(chunk):size//chunk},file_write_seconds=.1,
                produce_flush_seconds=.3,flush_seconds=.1,sync_seconds=.1,close_seconds=.1)
            files.append(dict(shard=shard,stage=stage,writer=shard%writers,pid=1,record=record,stats=stats,publish_seconds=1,readback_seconds=1))
        rows.append(dict(kind='measurement',trial=trial,repeat=trial//6,buffer_bytes=buffer,writers=writers,codec_chunk_bytes=65536,
            file_bytes=size,total_bytes=4*size,payloads_retained=trial==11,peak_rss_bytes=1024,
            publication_seconds=1,readback_seconds=1,publish_readback_seconds=2,files=files,matches=True))
    return rows+[dict(kind='completed',trials=12)]


class Review(unittest.TestCase):
    def test_fixed_modes_and_final_evidence(self):
        with tempfile.TemporaryDirectory()as path:
            root=Path(path);rows=fixture(root);self.assertEqual(len(HELP['review'](rows,root,small=True)),12)
    def test_reject_coverage_owner_counts_or_limits(self):
        with tempfile.TemporaryDirectory()as path:
            root=Path(path);rows=fixture(root)
            for change in ('missing','owner','calls','RSS','bytes','order'):
                bad=copy.deepcopy(rows)
                if change=='missing':bad.pop()
                elif change=='owner':bad[0]['files'][1]['writer']=1
                elif change=='calls':bad[0]['files'][0]['stats']['calls']+=1
                elif change=='RSS':bad[0]['peak_rss_bytes']=(2<<30)+1
                elif change=='bytes':bad[0]['total_bytes']+=1
                else:bad[0],bad[1]=bad[1],bad[0]
                with self.assertRaises(AssertionError):HELP['review'](bad,root,small=True)
    def test_reject_final_corruption_or_changed_manifest(self):
        with tempfile.TemporaryDirectory()as path:
            root=Path(path);rows=fixture(root)
            payload=root/'io-trial-11-file-0/payload.bin';data=payload.read_bytes();payload.write_bytes(b'x'+data[1:])
            with self.assertRaises(AssertionError):HELP['review'](rows,root,small=True)
            payload.write_bytes(data)
            manifest=root/'io-trial-00-file-0/manifest.json';m=json.loads(manifest.read_text());m['files']['payload.bin']['bytes']+=1;manifest.write_text(json.dumps(m))
            with self.assertRaises(AssertionError):HELP['review'](rows,root,small=True)


if __name__=='__main__':unittest.main()
