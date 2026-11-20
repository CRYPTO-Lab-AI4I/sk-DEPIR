#!/usr/bin/env python3
from pathlib import Path
import hashlib
import runpy
import struct
import unittest
from unittest import mock
HELP=runpy.run_path(str(Path(__file__).with_name('run.py')))

class Review(unittest.TestCase):
    def fixture(self):
        payload=struct.pack('<65537I',*range(65537));sha=hashlib.sha256(payload).hexdigest()
        wide=hashlib.sha256(b''.join(struct.pack('<Q',(1<<34)|j)for j in range(16384))).hexdigest()
        rows=[]
        for rank in range(4):
            rows.append(dict(kind='host',rank=rank,host='a'if rank<2 else'b',mode='direct',reclaim=False,staging_bytes=0,chunk_words=65536,payload_sha256=sha))
            rows.extend(dict(kind='progress',rank=rank,completed=j,total=256)for j in range(1,257))
            rows.append(dict(kind='completed',rank=rank,handoffs=256))
        for cycle in range(64):
            for case,(sender,receiver,width)in enumerate(((1,0,4),(2,0,4),(0,1,8),(0,2,8))):
                proof=(sha if width==4 else wide)*2
                for rank,peer,direction in ((sender,receiver,'send'),(receiver,sender,'receive')):
                    rows.append(dict(kind='proof',rank=rank,peer=peer,direction=direction,cycle=cycle,case=case,word_bytes=width,own_proof=proof,remote_proof=proof,matches=True,file=''))
        return payload,sha,rows
    def test_complete_and_corrupt_proof(self):
        payload,sha,rows=self.fixture()
        with mock.patch.dict(HELP['review'].__globals__,PAYLOAD_SHA=sha):
            self.assertEqual(HELP['review'](rows,payload,'direct',False)['proof_events'],512)
            rows[-1]['own_proof']='00'*64
            with self.assertRaises(AssertionError):HELP['review'](rows,payload,'direct',False)
    def test_missing_progress_or_wrong_mode(self):
        payload,sha,rows=self.fixture()
        with mock.patch.dict(HELP['review'].__globals__,PAYLOAD_SHA=sha):
            with self.assertRaises(AssertionError):HELP['review'](rows,payload,'staged',False)
            rows.pop(1)
            with self.assertRaises(AssertionError):HELP['review'](rows,payload,'direct',False)

    def test_transport_control_must_reach_every_worker(self):
        payload,sha,rows=self.fixture();setting={'OMPI_MCA_pml':'ob1','OMPI_MCA_btl':'self,tcp'}
        for row in rows:
            if row.get('kind')=='host':row['mpi_transport']=dict(setting)
        with mock.patch.dict(HELP['review'].__globals__,PAYLOAD_SHA=sha):
            HELP['review'](rows,payload,'direct',False,setting)
            rows[0]['mpi_transport']['OMPI_MCA_pml']='ucx'
            with self.assertRaises(AssertionError):HELP['review'](rows,payload,'direct',False,setting)

    def test_registration_cache_control_rejects_default_or_missing_flag(self):
        payload,sha,rows=self.fixture();setting={'UCX_RCACHE_ENABLE':'n'}
        for row in rows:
            if row.get('kind')=='host':row['mpi_transport']=dict(setting)
        with mock.patch.dict(HELP['review'].__globals__,PAYLOAD_SHA=sha):
            HELP['review'](rows,payload,'direct',False,setting)
            rows[0]['mpi_transport']['UCX_RCACHE_ENABLE']='try'
            with self.assertRaises(AssertionError):HELP['review'](rows,payload,'direct',False,setting)
            rows[0]['mpi_transport'].clear()
            with self.assertRaises(AssertionError):HELP['review'](rows,payload,'direct',False,setting)

    def test_pml_selection_is_not_component_availability(self):
        analyze=runpy.run_path(str(Path(__file__).with_name('analyze.py')))
        text='0: [node:123] select: initializing pml component ob1\n0: [node:123] select: component ucx selected\n0: [node:123] select: component ob1 not selected / finalized\n'
        selected=analyze['selected_pml'](text)
        self.assertEqual([(r['rank'],r['component'])for r in selected],[(0,'ucx')])

if __name__=='__main__':unittest.main()
