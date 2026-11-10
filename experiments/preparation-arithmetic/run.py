"""Bounded, isolated CPU comparison; never touches the retained s16 cache."""
import argparse
import hashlib
import json
import os
from pathlib import Path
import runpy
import shutil
import signal
import subprocess
import time

ROOT = Path(__file__).resolve().parents[2]
HELP = runpy.run_path(str(ROOT/'experiments/hpc-cached-s15/run.py'))


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument('name')
    parser.add_argument('--profile-only', action='store_true')
    args = parser.parse_args()
    if not args.name.replace('-', '').isalnum(): raise ValueError('invalid name')
    job = os.environ['SLURM_JOB_ID']
    out = ROOT/'experiments/preparation-arithmetic/runs'/args.name
    out.mkdir(parents=True, exist_ok=False)
    scratch = Path('/tmp')/('sk-depir-preparation-'+job)
    scratch.mkdir(exist_ok=False)
    cache = ROOT/'.cache/liftrs/preparation-arithmetic'/args.name
    cache.mkdir(parents=True, exist_ok=False)
    env = dict(os.environ, GOTOOLCHAIN='local', GOPROXY='off', GOSUMDB='off',
               GOPATH=str(ROOT/'.cache/liftrs/toolchains/gopath'), GOCACHE=str(scratch/'go-cache'),
               TMPDIR=str(scratch), CGO_ENABLED='1', CXX='g++', GOMAXPROCS='4', GOFLAGS='-p=4')
    go = str(ROOT/'.cache/liftrs/toolchains/go1.27.3-linux-amd64/go/bin/go')
    record = {'revision': subprocess.check_output(['git','rev-parse','HEAD'],cwd=ROOT,text=True).strip(),
              'job':job, 'cache':str(cache), 'commands':[], 'status':'running'}
    started = time.monotonic()

    def run(label, command, cwd=ROOT, extra=None):
        begin = time.monotonic(); peak = 0
        with (out/(label+'.log')).open('w') as log:
            p = subprocess.Popen(command,cwd=cwd,env=dict(env,**(extra or {})),stdout=log,stderr=subprocess.STDOUT,start_new_session=True)
            try:
                while p.poll() is None:
                    peak = max(peak, HELP['group_rss'](p.pid))
                    if peak > 12*2**30 or time.monotonic()-started > 1500:
                        raise RuntimeError('12 GiB RSS / 25 minute experiment guard')
                    time.sleep(1)
            finally:
                if p.poll() is None:
                    os.killpg(p.pid,signal.SIGTERM)
                    try: p.wait(timeout=20)
                    except subprocess.TimeoutExpired: os.killpg(p.pid,signal.SIGKILL); p.wait()
        record['commands'].append({'label':label,'argv':list(map(str,command)), 'seconds':time.monotonic()-begin,'peak_rss_bytes':peak,'exit_code':p.returncode})
        if p.returncode: raise RuntimeError(label+' failed')

    try:
        for label, command in [('cpu',['lscpu']),('compiler',['g++','--version']),('go',[go,'version']),('job',['scontrol','show','job',job])]: run(label,command)
        variants = ['baseline'] if args.profile_only else ['baseline','optimized']
        for variant in variants:
            source = scratch/variant
            shutil.copytree(ROOT,source,ignore=shutil.ignore_patterns('.git','.cache','runs','__pycache__','slurm-*.log'))
            if variant == 'baseline':
                for name, digest in json.loads((ROOT/'experiments/preparation-arithmetic/baseline.json').read_text()).items():
                    (source/name).write_bytes((ROOT/'experiments/source-snapshots'/(digest+'.txt')).read_bytes())
            hashes = {str(p.relative_to(source)):hashlib.sha256(p.read_bytes()).hexdigest() for p in source.rglob('*.go')}
            record[variant+'_sources'] = hashes
            for name, digest in hashes.items():
                dest = ROOT/'experiments/source-snapshots'/(digest+'.txt')
                if not dest.exists(): dest.write_bytes((source/name).read_bytes())
            binary = cache/(variant+'.test')
            run(variant+'-build',[go,'test','-tags','liftrs_cpp','-c','-o',str(binary),'./encoding/liftrs-shared'],source)
            record[variant+'_binary_sha256'] = hashlib.sha256(binary.read_bytes()).hexdigest()
            leaf = out/(variant+'-leaves'); leaf.mkdir()
            run(variant+'-leaves',[str(binary),'-test.run=^TestPreparationLeafProfile$','-test.v'],extra={'PIR_PREPARATION_OUT':str(leaf),'PIR_PREPARATION_LOOKUP':'1' if variant=='optimized' else '0'})
            run(variant+'-pprof',[go,'tool','pprof','-top','-nodecount=20',str(binary),str(leaf/'leaves.pprof')])
            if not args.profile_only:
                run(variant+'-tests',[go,'test','-tags','liftrs_cpp','./encoding/liftrs-shared','./encoding/liftrs-cpp'],source)
                for repeat in range(3):
                    dest = out/(variant+'-'+str(repeat)); dest.mkdir()
                    constants = cache/(variant+'-'+str(repeat))
                    run(variant+'-'+str(repeat),[str(binary),'-test.run=^TestPreparationBenchmark$','-test.v','-test.timeout=8m'],extra={'PIR_PREPARATION_OUT':str(dest),'PIR_PREPARATION_CACHE':str(constants)})
        record['status'] = 'passed'
    except Exception as error:
        record['status']='failed'; record['error']=repr(error)
        raise
    finally:
        record['evidence_sha256']={str(p.relative_to(out)):hashlib.sha256(p.read_bytes()).hexdigest() for p in out.rglob('*') if p.is_file() and p.name!='results.json'}
        (out/'results.json').write_text(json.dumps(record,indent=2)+'\n')


if __name__ == '__main__': main()
