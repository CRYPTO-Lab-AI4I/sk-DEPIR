#!/usr/bin/env python3
"""Reconcile a completed two-node probe and derive remote payload throughput."""
import argparse
import hashlib
import json
from pathlib import Path
import runpy
import statistics
import subprocess

HERE=Path(__file__).resolve().parent
ROOT=HERE.parents[1]


def sha(path):return hashlib.sha256(path.read_bytes()).hexdigest()


def summarize(rows):
    runpy.run_path(str(HERE/'run.py'))['verify_rows'](rows)
    sizes=[]
    for size in (1<<20,8<<20,32<<20):
        first=[r for r in rows if r.get('phase')=='timed-redistribution' and r['rank']==0 and r['buffer_bytes_each']==size]
        second=[r for r in rows if r.get('phase')=='timed-redistribution' and r['rank']==1 and r['buffer_bytes_each']==size]
        times={r['completed']:r['exchange_seconds'] for r in first}
        assert times=={r['completed']:r['exchange_seconds'] for r in second}
        latency=list(times.values())
        # Each rank keeps half its own data and sends the other half remotely.
        median=statistics.median(latency)
        sizes.append(dict(input_MiB_per_rank=size/(1<<20),remote_MiB_per_rank=size/2/(1<<20),
            samples_seconds=latency,median_seconds=median,
            remote_payload_GiBs_per_node=(size/2)/(1<<30)/median,
            bidirectional_aggregate_payload_GiBs=size/(1<<30)/median))
    return dict(samples=sizes,hosts={str(r['rank']):r['host'] for r in rows if 'host'in r},
                peak_RSS_MiB_by_rank={str(rank):max(r.get('peak_rss_KiB',0) for r in rows if r['rank']==rank)/1024 for rank in (0,1)})


def main():
    parser=argparse.ArgumentParser(description=__doc__);parser.add_argument('name');args=parser.parse_args()
    out=HERE/'runs'/args.name
    record=json.loads((out/'results.json').read_text())
    assert record['status']=='passed' and all(c['exit_code']==0 for c in record['commands'])
    for name,digest in record['source_sha256'].items():
        source=subprocess.check_output(['git','show',record['revision']+':experiments/preprocessing-transport/'+name],cwd=ROOT)
        assert hashlib.sha256(source).hexdigest()==digest,name
    for name,digest in record['evidence_sha256'].items():assert sha(out/name)==digest,name
    binary=ROOT/'.cache/liftrs/preprocessing-transport'/args.name/'probe'
    assert sha(binary)==record['binary_sha256']
    rows=[json.loads(line[line.index('{'):]) for line in (out/'transport.log').read_text().splitlines() if '{' in line]
    assert rows==record['measurements']
    result=summarize(rows)
    result.update(status='passed',job_id=record['job_id'],revision=record['revision'],
        source_files_verified=len(record['source_sha256']),evidence_files_verified=len(record['evidence_sha256']),binary_verified=True,
        scope='two-node host-memory redistribution only; no GPU transform, root execution, four-node bandwidth or storage benchmark')
    (out/'analysis.json').write_text(json.dumps(result,indent=2)+'\n')
    print(json.dumps(result,indent=2))


if __name__=='__main__':main()
