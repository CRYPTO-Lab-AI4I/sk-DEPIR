#!/usr/bin/env python3
from pathlib import Path
import hashlib
import runpy
import tempfile
import unittest
from unittest import mock
HELP=runpy.run_path(str(Path(__file__).with_name('run.py')))
SHAPES=((1,3),(2,1),(3,2),(3,5))
SETTING={'UCX_RCACHE_ENABLE':'n'}


def fixture(cache):
    rows=[dict(kind='host',rank=r,host='a'if r<2 else'b',mode='direct',chunk_words=1024,staging_bytes=0,mpi_transport=dict(SETTING))for r in range(4)]
    for op in range(1,9):
        width=4 if op<=4 else 8;nr,nc=SHAPES[(op-1)%4];tile=16384 if (op-1)%4==3 else 7
        mask=(1<<(8*width))-1;data=[(j*0x9e3779b97f4a7c15^0xa5a5a5a5a5a5a5a5)&mask for j in range(nr*nc)]
        transposed=[data[r*nc+c]for c in range(nc)for r in range(nr)]
        for rank in range(4):
            _,ir=HELP['band'](nr,rank);first,oc=HELP['band'](nc,rank)
            values=transposed[first*nr:(first+oc)*nr];raw=b''.join(v.to_bytes(width,'little')for v in values)
            rows.append(dict(kind='measurement',operation=op,rank=rank,rows=nr,columns=nc,word_bytes=width,matches=True,input_words=ir*nc,output_words=oc*nr,
                scratch_words=min(tile,max(ir*HELP['band'](nc,0)[1],oc*HELP['band'](nr,0)[1])),seconds=0.01,output_sha256=hashlib.sha256(raw).hexdigest(),peak_rss_bytes=32<<20))
            if op in(4,8):(cache/f'rank-{rank}-op-{op}-output.bin').write_bytes(raw)
        for source in range(4):
            first_row,count_row=HELP['band'](nr,source)
            for dest in range(4):
                if source==dest:continue
                first_col,count_col=HELP['band'](nc,dest)
                block=[data[r*nc+c]for c in range(first_col,first_col+count_col)for r in range(first_row,first_row+count_row)]
                for offset in range(0,len(block),tile):
                    values=block[offset:offset+tile];proof=hashlib.sha256(b''.join(v.to_bytes(width,'little')for v in values)).hexdigest()*2
                    for rank,peer,direction in ((source,dest,'send'),(dest,source,'receive')):
                        rows.append(dict(kind='tile-proof',operation=op,rank=rank,peer=peer,direction=direction,round=(dest-source)%4,offset=offset,words=len(values),word_bytes=width,matches=True,file='',own_proof=proof,remote_proof=proof))
    for rank in range(4):
        rows.append(dict(kind='completed',rank=rank,operations=8))
        for lane in('root32','support64'):rows.append(dict(kind='legacy-check',rank=rank,lane=lane,matches=True))
    return rows


class Review(unittest.TestCase):
    def test_complete_uneven_and_empty_shards(self):
        with tempfile.TemporaryDirectory()as tmp,mock.patch.dict(HELP['review'].__globals__,SHAPES=SHAPES):
            cache=Path(tmp);rows=fixture(cache);self.assertEqual(len(HELP['review'](rows,cache,'direct',SETTING)['operations']),8)
    def test_missing_packet_or_bad_saved_bytes_fail(self):
        with tempfile.TemporaryDirectory()as tmp,mock.patch.dict(HELP['review'].__globals__,SHAPES=SHAPES):
            cache=Path(tmp);rows=fixture(cache);bad=list(rows);bad.pop(next(j for j,r in enumerate(bad)if r.get('kind')=='tile-proof'))
            with self.assertRaises(AssertionError):HELP['review'](bad,cache,'direct',SETTING)
            (cache/'rank-0-op-8-output.bin').write_bytes(b'bad width')
            with self.assertRaises(AssertionError):HELP['review'](rows,cache,'direct',SETTING)


if __name__=='__main__':unittest.main()
