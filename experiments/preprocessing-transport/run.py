#!/usr/bin/env python3
"""Two-node host-memory transport gate, not distributed GPU arithmetic."""
import argparse
import datetime
import hashlib
import json
import os
import resource
from pathlib import Path
import subprocess

HERE=Path(__file__).resolve().parent
ROOT=HERE.parents[1]


def verify_rows(rows):
    hosts={r['rank']:r['host'] for r in rows if 'host'in r}
    assert set(hosts)=={0,1} and len(set(hosts.values()))==2
    for rank in (0,1):
        assert sum(r.get('rank')==rank and r.get('phase')=='workflow' and r.get('event')=='completed' for r in rows)==1
        for size in (1<<20,8<<20,32<<20):
            samples=[r for r in rows if r.get('rank')==rank and r.get('phase')=='timed-redistribution' and r['buffer_bytes_each']==size]
            assert len(samples)==3 and {r['completed'] for r in samples}=={2,3,4}
            assert all(r['exchange_seconds']>0 and r['peak_rss_KiB']<=1<<20 for r in samples)


def main():
    parser=argparse.ArgumentParser(description=__doc__)
    parser.add_argument('name',nargs='?')
    parser.add_argument('--check',action='store_true')
    args=parser.parse_args()
    assert 'MPI_Alltoall' in (HERE/'probe.cpp').read_text()
    if args.check:
        print('two ranks, two nodes, zero GPUs; 64MiB explicit buffers/rank,1GiB RSS guard/rank; no submission')
        return
    if not args.name or not args.name.replace('-','').isalnum():parser.error('fresh name required')
    if os.environ.get('SLURM_JOB_NUM_NODES')!='2':raise RuntimeError('two-node allocation required')
    # Bound every local artifact file, including subprocess stdout. Remote ranks
    # create no files; their stdout is forwarded to this bounded local log.
    resource.setrlimit(resource.RLIMIT_FSIZE,(32<<20,32<<20))
    out=HERE/'runs'/args.name
    out.mkdir(parents=True,exist_ok=False)
    cache=ROOT/'.cache/liftrs/preprocessing-transport'/args.name
    cache.mkdir(parents=True,exist_ok=False)
    record=dict(job_id=os.environ['SLURM_JOB_ID'],started_utc=datetime.datetime.now(datetime.timezone.utc).isoformat(),
        revision=subprocess.check_output(['git','rev-parse','HEAD'],text=True).strip(),status='running',commands=[],
        source_sha256={p.name:hashlib.sha256(p.read_bytes()).hexdigest() for p in (HERE/'probe.cpp',HERE/'run.py',HERE/'run.sbatch')},
        resources=dict(nodes=2,ranks=2,cpus_per_node=2,requested_host_GiB_per_node=4,gpus=0,
                       owned_buffers_MiB_per_rank=64,rss_guard_GiB_per_rank=1,controller_seconds=480,scheduler_seconds=600))
    def save(): (out/'results.json').write_text(json.dumps(record,indent=2)+'\n')
    def run(label,command,seconds=30):
        save()
        with (out/(label+'.log')).open('w') as log:
            result=subprocess.run(command,cwd=ROOT,stdout=log,stderr=subprocess.STDOUT,timeout=seconds)
        record['commands'].append(dict(label=label,argv=command,exit_code=result.returncode))
        save()
        if result.returncode:raise RuntimeError(label+' failed')
    save()
    try:
        run('hosts',['scontrol','show','hostnames',os.environ['SLURM_JOB_NODELIST']])
        run('mpi',['ompi_info','--parsable','--all'])
        run('build',['mpicxx','-std=c++17','-O3',str(HERE/'probe.cpp'),'-o',str(cache/'probe')],60)
        run('transport',['timeout','--signal=TERM','--kill-after=10s','300s','mpirun','--bind-to','core','--map-by','ppr:1:node','-np','2',str(cache/'probe')],320)
        rows=[json.loads(line[line.index('{'):]) for line in (out/'transport.log').read_text().splitlines() if '{' in line]
        verify_rows(rows)
        record['measurements']=rows
        record['status']='passed'
    except BaseException as error:
        record['status']='failed'
        record['error']=str(error)
        raise
    finally:
        record['evidence_sha256']={p.name:hashlib.sha256(p.read_bytes()).hexdigest() for p in out.glob('*.log')}
        record['binary_sha256']=hashlib.sha256((cache/'probe').read_bytes()).hexdigest() if (cache/'probe').exists() else None
        save()


if __name__=='__main__':main()
