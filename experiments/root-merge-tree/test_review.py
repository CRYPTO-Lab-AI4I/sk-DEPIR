#!/usr/bin/env python3
"""Check owned-tree evidence review with tiny files, no MPI/GPU work."""
import copy
import hashlib
from pathlib import Path
import runpy
import struct
import tempfile
import unittest
HELP=runpy.run_path(str(Path(__file__).with_name('run.py')))


def fixture(root):
    rows=[];leaf=[1,1];middle=[1,0,1];final=[1,0,0,0,1]
    def packed(values,width=4):return struct.pack('<'+('I'if width==4 else'Q')*len(values),*values)
    def digest(values):return hashlib.sha256(packed(values)).hexdigest()
    schedule={0:[(1,1,False),(2,2,False)],1:[(1,0,True)],2:[(1,3,False),(2,0,True)],3:[(1,2,True)]}
    for rank in range(4):
        dest=root/f'rank-{rank}';dest.mkdir()
        def save(name,p,width=4):(dest/name).write_bytes(packed(p,width))
        save('case-0-support.bin',[rank*257],8);save('case-0-local.bin',leaf)
        if rank==0:save('case-0-reference.bin',final);save('case-0-final.bin',final)
        rows.extend([dict(kind='host',rank=rank,host=f'node{rank//2}',local_rank=str(rank%2),visible_devices='0',chunk_words=1024),
            dict(kind='fixture',rank=rank,case=0,s=8,support_count=4,coordinator_support_count=4 if rank==0 else 0,
                reference_count=5 if rank==0 else 0,reference_sha256=digest(final)if rank==0 else hashlib.sha256(b'').hexdigest(),
                subtree_ranks=(4,1,2,1)[rank],max_elements=(8,4,8,4)[rank]),
            dict(kind='completed',rank=rank,cases=1)])
        for sample in (-1,2):
            current=leaf
            for stride,peer,sending in schedule[rank]:
                rows.append(dict(kind='tree',rank=rank,case=0,sample=sample,stride=stride,peer=peer,sending=sending,event='started',sha256=digest(current)))
                if sending:
                    name=f'case-0-stride-{stride}-send-{peer}.bin';payload=current
                else:
                    payload=leaf if stride==1 else middle
                    name=f'case-0-stride-{stride}-receive-{peer}.bin'
                    current=middle if stride==1 else final
                    save(f'case-0-merged-{stride}.bin',current)
                save(name,payload)
                rows.append(dict(kind='transfer',rank=rank,case=0,sample=sample,stride=stride,peer=peer,direction='send'if sending else'receive',
                    words=len(payload),sha256=digest(payload),file=name))
                rows.append(dict(kind='tree',rank=rank,case=0,sample=sample,stride=stride,peer=peer,sending=sending,event='completed',sha256=digest(current)))
            row=dict(kind='measurement',rank=rank,case=0,sample=sample,s=8,support_count=4,shard_count=1,ranks=4,
                owned_result_words=5 if rank==0 else 0,owned_device_bytes=1024,peak_rss_bytes=2048,validation_io=sample==2,matches=True)
            row.update({f:5.0 if f=='seconds'else 1.0 for f in HELP['FIELDS']})
            rows.append(row)
    return rows


class Review(unittest.TestCase):
    def test_complete_owned_state_chain(self):
        with tempfile.TemporaryDirectory()as path:
            root=Path(path);self.assertEqual(len(HELP['review'](fixture(root),root,small=True)),8)
    def test_reject_missing_rank_replication_wrong_owner_or_digest(self):
        with tempfile.TemporaryDirectory()as path:
            root=Path(path);rows=fixture(root)
            for kind in ('missing','replication','ownership','digest','cap','topology','invented_chunks'):
                changed=copy.deepcopy(rows)
                if kind=='missing':changed.pop()
                elif kind=='replication':next(r for r in changed if r.get('kind')=='fixture'and r['rank']==1)['reference_count']=5
                elif kind=='ownership':next(r for r in changed if r.get('kind')=='measurement'and r['rank']==1)['owned_result_words']=5
                elif kind=='digest':next(r for r in changed if r.get('kind')=='transfer')['sha256']='wrong'
                elif kind=='cap':next(r for r in changed if r.get('kind')=='measurement')['owned_device_bytes']=(64<<20)+1
                elif kind=='topology':next(r for r in changed if r.get('kind')=='host'and r['rank']==1)['visible_devices']='1'
                else:changed.append(dict(kind='mpi-chunk',event='completed',rank=0,peer=1,total=1025,completed=1025,unit='words',phase='root-transfer',sending=True))
                with self.assertRaises(AssertionError):HELP['review'](changed,root,small=True)
    def test_reject_retained_intermediate_or_final_corruption(self):
        with tempfile.TemporaryDirectory()as path:
            root=Path(path);rows=fixture(root)
            for name in ('rank-0/case-0-merged-1.bin','rank-2/case-0-stride-2-send-0.bin','rank-0/case-0-final.bin'):
                file=root/name;original=file.read_bytes();file.write_bytes(b'\x02'+original[1:])
                with self.assertRaises(AssertionError):HELP['review'](rows,root,small=True)
                file.write_bytes(original)


if __name__=='__main__':unittest.main()
