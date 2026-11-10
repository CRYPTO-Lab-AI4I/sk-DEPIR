#!/usr/bin/env python3
"""Record external cancellation without rewriting the interrupted controller log."""
import datetime
import hashlib
import json
from pathlib import Path
import subprocess
import sys

HERE = Path(__file__).resolve().parent
ROOT = HERE.parents[1]


def main():
    if len(sys.argv) != 2 or Path(sys.argv[1]).name != sys.argv[1]:
        raise SystemExit('Usage: record-stop.py RUN_NAME')
    out = HERE/'runs'/sys.argv[1]
    record = json.loads((out/'results.json').read_text())
    accounting = subprocess.check_output(['sacct','-j',record['job_id'],
        '--format=JobID,State,Elapsed,ExitCode','--noheader','--parsable2'],text=True)
    states = [line.split('|')[1] for line in accounting.splitlines() if line]
    if not states or any(s not in ('FAILED','CANCELLED','COMPLETED','TIMEOUT','OUT_OF_MEMORY') for s in states):
        raise SystemExit('Job not conclusively terminal: '+accounting)
    cache = Path(record['cache'])
    meta = json.loads((cache/'basis-metadata.json').read_text())
    h = hashlib.sha256()
    with (cache/'support.bin').open('rb') as f:
        for block in iter(lambda:f.read(1<<20),b''): h.update(block)
    assert h.hexdigest() == meta['sha256']
    assert (cache/'support.bin').stat().st_size == meta['bytes'] == 4*meta['k']
    stop = {'status':'stopped','timestamp_utc':datetime.datetime.now(datetime.timezone.utc).isoformat(),
            'job_id':record['job_id'],'reason':'User-authorized stop after measured progress exceeded the job budget.',
            'accounting':accounting,'basis':meta,'support_checksum_verified':True,
            'last_worker_state':json.loads((out/'initial/state.json').read_text()),
            'preparation_profile_bytes':(out/'initial/preparation.pprof').stat().st_size,
            'limits':'Original results.json was not finalized; batch-shell termination prevented controller cleanup. Preparation CPU profile is empty. Completed phase profiles and discovery survive; no full preparation or encoding result.'}
    with (out/'interruption.json').open('x') as f: json.dump(stop,f,indent=2); f.write('\n')
    print(json.dumps(stop,indent=2))


if __name__ == '__main__': main()
