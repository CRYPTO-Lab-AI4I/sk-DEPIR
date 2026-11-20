#!/usr/bin/env python3
"""Bounded rank-owned transpose prerequisite for collective preprocessing FFTs."""
from pathlib import Path
import hashlib
import json
import runpy
import sys
import shutil
HERE=Path(__file__).resolve().parent
CONTROL=runpy.run_path(str(HERE.parent/'ucx-registration-cache/run.py'))
BASE=CONTROL['HOST']['BASE']
SHAPES=((1,3),(7,11),(64,128),(1024,256))
MASK=(1<<64)-1


def band(size,rank,ranks=4):
    q,r=divmod(size,ranks);return q*rank+min(rank,r),q+(rank<r)


def review(rows,cache,mode,settings):
    hosts=[r for r in rows if r.get('kind')=='host']
    assert len(hosts)==4 and {r['rank']for r in hosts}==set(range(4))
    placement={r['rank']:r['host']for r in hosts}
    assert placement[0]==placement[1]and placement[2]==placement[3]and placement[0]!=placement[2]
    assert all(r['mode']==mode and r['chunk_words']==1024 and r['staging_bytes']==(8192 if mode=='staged'else 0)and
        all(r['mpi_transport'].get(k)==v for k,v in settings.items())for r in hosts)
    measures=[r for r in rows if r.get('kind')=='measurement'];assert len(measures)==32
    proofs=[r for r in rows if r.get('kind')=='tile-proof']
    expected_proofs=0;summary=[]
    for op in range(1,9):
        width=4 if op<=4 else 8;shape=SHAPES[(op-1)%4];nr,nc=shape;tile=16384 if (op-1)%4==3 else 7
        selected=[r for r in measures if r['operation']==op]
        assert len(selected)==4 and {r['rank']for r in selected}==set(range(4))
        expected_events={}
        for source in range(4):
            _,source_rows=band(nr,source)
            for dest in range(4):
                if source==dest:continue
                _,dest_columns=band(nc,dest);words=source_rows*dest_columns;round_id=(dest-source)%4
                for offset in range(0,words,tile):
                    count=min(tile,words-offset)
                    expected_events[(source,dest,'send',round_id,offset)]=count
                    expected_events[(dest,source,'receive',round_id,offset)]=count
        actual={};packet_rows={}
        for row in proofs:
            if row['operation']!=op:continue
            key=(row['rank'],row['peer'],row['direction'],row['round'],row['offset'])
            assert key not in actual and row['word_bytes']==width and row['matches']and not row['file']
            own,remote=bytes.fromhex(row['own_proof']),bytes.fromhex(row['remote_proof'])
            assert len(own)==len(remote)==64 and own[:32]==own[32:]==remote[:32]==remote[32:]
            actual[key]=row['words'];packet_rows[key]=row
        assert actual==expected_events;expected_proofs+=len(expected_events)
        for (rank,peer,direction,round_id,offset),row in packet_rows.items():
            if direction!='send':continue
            other=packet_rows[(peer,rank,'receive',round_id,offset)]
            assert row['own_proof']==other['remote_proof']and row['remote_proof']==other['own_proof']
            first_row,source_rows=band(nr,rank);first_column,_=band(nc,peer)
            digest=hashlib.sha256()
            for index in range(offset,offset+row['words']):
                column=first_column+index//source_rows;orig_row=first_row+index%source_rows
                value=((orig_row*nc+column)*0x9e3779b97f4a7c15^0xa5a5a5a5a5a5a5a5)&((1<<(8*width))-1)
                digest.update(value.to_bytes(width,'little'))
            assert row['own_proof']==digest.hexdigest()*2
        for row in selected:
            rank=row['rank'];_,own_rows=band(nr,rank);first,own_columns=band(nc,rank)
            assert row['rows']==nr and row['columns']==nc and row['word_bytes']==width and row['matches']
            assert row['input_words']==own_rows*nc and row['output_words']==own_columns*nr
            assert row['scratch_words']==min(tile,max(own_rows*band(nc,0)[1],own_columns*band(nr,0)[1]))
            assert row['seconds']>0 and row['seconds']==selected[0]['seconds']
            # Independent global-coordinate oracle, limited to this owner's output.
            digest=hashlib.sha256();data=bytearray()
            for col in range(first,first+own_columns):
                for orig_row in range(nr):
                    value=((orig_row*nc+col)*0x9e3779b97f4a7c15^0xa5a5a5a5a5a5a5a5)&((1<<(8*width))-1)
                    packed=value.to_bytes(width,'little');digest.update(packed)
                    if op in(4,8):data.extend(packed)
            assert digest.hexdigest()==row['output_sha256']
            if op in(4,8):assert (cache/f'rank-{rank}-op-{op}-output.bin').read_bytes()==data
        summary.append(dict(operation=op,rows=nr,columns=nc,word_bytes=width,seconds=selected[0]['seconds']))
    assert len(proofs)==expected_proofs
    for rank in range(4):
        assert sum(r.get('kind')=='completed'and r['rank']==rank and r['operations']==8 for r in rows)==1
        for lane in ('root32','support64'):
            assert sum(r.get('kind')=='legacy-check'and r['rank']==rank and r['lane']==lane and r['matches']for r in rows)==1
    assert all(r.get('peak_rss_bytes',0)<=1<<30 for r in rows)
    return dict(mode=mode,operations=summary,proof_events=len(proofs),peak_rss_bytes_per_rank=max(r.get('peak_rss_bytes',0)for r in rows),hosts=placement)


def suite(context):
    settings=CONTROL['configure'](context);run=context['run'];cache=context['cache'];out=context['out'];env=context['env'];go=context['go']
    context['resources'].update(ranks=4,ranks_per_node=2,cpus_per_rank=1,gpus=0,gpu_model=None,owned_cap_MiB_per_card=0,pinned_cap_MiB_per_rank=0,rss_cap_GiB_per_rank=1,max_shard_words=1<<18,tile_words=16384,mpi_chunk_words=1024)
    context['save']();env.update(CGO_ENABLED='1',CC='mpicc',CXX='mpicxx',GOMAXPROCS='1',GOMEMLIMIT=str(512<<20),PIR_RANK_LOG_FILES='1')
    binary=cache/'transpose.test'
    run('build',[go,'test','-c','-tags','liftrs_mpi','-o',binary,'./encoding/liftrs-shared'],120)
    run('local-tests',[binary,'-test.run=^TestRankTranspose|^TestTransposeDescriptor','-test.v','-test.timeout=30s'],45)
    run('local-race',[go,'test','-race','./encoding/liftrs-shared','-run','^TestRankTranspose','-count=1','-timeout=45s'],90)
    wrapper=cache/'rank.sh';shutil.copyfile(context['source']/'experiments/distributed-root/rank.sh',wrapper)
    state=dict(status='running',arms=[])
    def save():(out/'transpose-progress.json').write_text(json.dumps(state,indent=2)+'\n')
    save()
    for mode in ('staged','direct'):
        state['current_arm']=mode;save();directory=cache/mode;directory.mkdir();env.update(PIR_TRANSPOSE_OUTPUT=str(directory),PIR_TRANSPOSE_BUFFER=mode)
        try:
            run(mode,['timeout','--signal=TERM','--kill-after=10s','200s','srun','--mpi=pmix_v3','--nodes=2','--ntasks=4','--ntasks-per-node=2','--cpus-per-task=1','--distribution=block:block','--cpu-bind=cores','--chdir='+str(cache),'--label','--kill-on-bad-exit=1','bash',wrapper,out/mode,binary,'-test.run=^TestDistributedTransposeMPI$','-test.v','-test.timeout=180s'],220)
            state['arms'].append(review(BASE['parse_rank_logs'](out/(mode+'.log')),directory,mode,settings));save()
        except BaseException as error:state.update(status='failed',error=str(error));save();raise
    run('profile',[go,'tool','pprof','-symbolize=none','-top',binary,out/'staged-0.pprof'],90)
    state['status']='passed';save();return dict(distributed_transpose=state)


if __name__=='__main__':
    if '--check'in sys.argv:print('2 HPC nodes / 4 ranks / 4 CPU / 8 GiB host / zero GPUs; bounded32/64bit shards/tiles; explicit UCX cache control; same15min scheduler envelope; no distributed FFT claim')
    else:BASE['main'](here=HERE,extra_suite=suite,require_gpu=False)
