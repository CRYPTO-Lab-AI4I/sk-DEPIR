#!/usr/bin/env python3
"""Reconcile the22 completed table cells, selecting replacements before hashing."""
import ast
import datetime
import json
import os
from pathlib import Path
import runpy
import signal
import subprocess
import time

HERE=Path(__file__).resolve().parent
ROOT=HERE.parents[1]
HELP=runpy.run_path(str(ROOT/'experiments/hpc-cached-s15/run.py'))
GUARD=runpy.run_path(str(ROOT/'experiments/hpc-s16/run.py'))

def selected_cases():
    """A replacement supersedes its cancelled predecessor before any file reads."""
    chosen={};manifests={}
    for batch in ('fill-01','fill-02'):
        manifest=HERE/'runs'/batch/'submissions.json'
        manifests[str(manifest.relative_to(ROOT))]=HELP['sha'](manifest)
        chosen.update({case:(batch,item) for case,item in json.loads(manifest.read_text())['cases'].items()})
    return chosen,manifests

def main():
    import argparse
    parser=argparse.ArgumentParser(description=__doc__)
    parser.add_argument('name',nargs='?');parser.add_argument('--check',action='store_true')
    args=parser.parse_args()
    if args.check:
        ast.parse(Path(__file__).read_text());assert 600<900<1200
        chosen,_=selected_cases();assert len(chosen)==22
        assert {case for case,(batch,_)in chosen.items()if batch=='fill-02'}=={'cpu-disk-s15','cpu-disk-s16','cpu-packed-s16'}
        for case,(batch,item)in chosen.items():
            record=json.loads((HERE/'runs'/batch/case/'results.json').read_text())
            assert record['status']=='passed' and record['job_id']==str(item['job_id'])
        print('reconciliation syntax/deadlines and22 replacement-selected completed records checked; hashes not reread');return
    if not args.name or not args.name.replace('-','').isalnum():parser.error('fresh run name required')
    out=HERE/'reviews'/args.name;out.mkdir(parents=True,exist_ok=False)
    sources=['experiments/encoding-table/'+n for n in ('reconcile.py','reconcile.sbatch','analyze.py','run.py')]+['experiments/source_evidence.py','experiments/hpc-cached-s15/run.py','experiments/hpc-s16/run.py']
    record={'status':'running','job_id':os.environ['SLURM_JOB_ID'],'revision':subprocess.check_output(['git','rev-parse','HEAD'],cwd=ROOT,text=True).strip(),
            'source_sha256':{n:HELP['sha'](ROOT/n) for n in sources},'started_utc':datetime.datetime.now(datetime.timezone.utc).isoformat(),
            'policy':{'cpus':4,'host_request_bytes':8<<30,'rss_guard_bytes':6<<30,'output_guard_bytes':2<<30,'worker_seconds':600,'controller_seconds':900,'slurm_seconds':1200},'commands':[],'rows':[]}
    chosen={};manifests={};start=time.monotonic();cancelled=[]
    signal.signal(signal.SIGTERM,lambda *_:cancelled.append(True))
    def save():(out/'results.json').write_text(json.dumps(record,indent=2)+'\n')
    def event(case,kind,**fields):
        row=dict(utc=datetime.datetime.now(datetime.timezone.utc).isoformat(),case=case,event=kind,completed=len(record['rows']),total=len(chosen),unit='cases',elapsed_seconds=time.monotonic()-start,**fields)
        with(out/'progress.jsonl').open('a')as stream:stream.write(json.dumps(row)+'\n')
        print(json.dumps(row),flush=True)
    save()
    try:
        chosen,manifests=selected_cases()
        assert len(chosen)==22
        record['submission_sha256']=manifests;digests={}
        for case,(batch,item) in chosen.items():
            for name,want in record['source_sha256'].items():assert HELP['sha'](ROOT/name)==want,'analysis source changed: '+name
            path=HERE/'runs'/batch/case;r=json.loads((path/'results.json').read_text())
            assert r['status']=='passed' and r['job_id']==str(item['job_id']),(batch,case)
            command=['python3','-B',str(HERE/'analyze.py'),batch,'--case',case];began=time.monotonic();last=began;peak=0;reason=None
            event(case,'started',batch=batch)
            with(out/(case+'.log')).open('w')as log:
                process=subprocess.Popen(command,cwd=ROOT,stdout=log,stderr=subprocess.STDOUT,start_new_session=True)
                try:
                    while process.poll()is None:
                        peak=max(peak,HELP['group_rss'](process.pid))
                        if peak>6<<30:reason='RSS guard'
                        if GUARD['tree_bytes'](out)>2<<30:reason='output guard'
                        if time.monotonic()-began>600 or time.monotonic()-start>900:reason='deadline'
                        if cancelled:reason='termination signal'
                        if time.monotonic()-last>=30:event(case,'heartbeat',command_seconds=time.monotonic()-began,peak_rss_bytes=peak);last=time.monotonic()
                        if reason:GUARD['stop_group'](process);break
                        time.sleep(1)
                finally:
                    if process.poll()is None:GUARD['stop_group'](process)
            record['commands'].append({'case':case,'argv':command,'seconds':time.monotonic()-began,'peak_rss_bytes':peak,'exit_code':process.returncode,'guard':reason});save()
            if reason or process.returncode:raise RuntimeError(case+': '+str(reason or process.returncode))
            analysis=path/'analysis.json';row=json.loads(analysis.read_text())['rows'][0]
            phases=[p for name in r['phase_files'] for p in json.loads((path/name).read_text())]
            digest=next(p['canonical_sha256']for p in phases if'systematic_checked'in p)
            if r['s']in digests:assert digest==digests[r['s']],('cross-mode canonical hash',case)
            digests[r['s']]=digest
            row.update(batch=batch,canonical_sha256=digest,analysis_sha256=HELP['sha'](analysis));record['rows'].append(row);save();event(case,'completed')
        record['status']='passed';record['canonical_sha256_by_s']=digests
    except BaseException as error:
        record['status']='failed';record['error']=repr(error);raise
    finally:
        record['elapsed_seconds']=time.monotonic()-start
        record['evidence_sha256']={str(p.relative_to(out)):HELP['sha'](p)for p in out.rglob('*')if p.is_file()and p.name!='results.json'}
        save()

if __name__=='__main__':main()
