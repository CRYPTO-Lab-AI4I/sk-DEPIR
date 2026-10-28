#!/usr/bin/env python3
"""Reproduce the s=10 control using the pre-change a2e0352 source tree."""
import argparse,io,json,os,platform,subprocess,tarfile,tempfile,time
from pathlib import Path
ROOT=Path(__file__).resolve().parents[2]
HERE=Path(__file__).resolve().parent
p=argparse.ArgumentParser(description=__doc__);p.add_argument('--go',default='go');args=p.parse_args()
with tempfile.TemporaryDirectory(prefix='depir-mul-control-') as directory:
    checkout=Path(directory)/'repo';checkout.mkdir()
    archive=subprocess.check_output(['git','archive','a2e0352'],cwd=ROOT)
    with tarfile.open(fileobj=io.BytesIO(archive)) as tar:tar.extractall(checkout)
    binary=Path(directory)/'baseline'
    subprocess.run([args.go,'build','-o',str(binary),'./experiments/cyclic-scaling'],cwd=checkout,check=True)
    started=time.monotonic()
    with (HERE/'baseline-s10-events.jsonl').open('w') as output,(HERE/'baseline-s10-stderr.txt').open('w') as error:
        proc=subprocess.Popen([str(binary),'-s','10','-t','6','-repeats','3'],cwd=checkout,stdout=output,stderr=error)
        _,status,usage=os.wait4(proc.pid,0)
    (HERE/'baseline-s10-process.json').write_text(json.dumps({'revision':'a2e0352','command':'cyclic-scaling -s 10 -t 6 -repeats 3','exit_status':os.waitstatus_to_exitcode(status),'wall_seconds':time.monotonic()-started,'peak_rss_bytes':usage.ru_maxrss*(1 if platform.system()=='Darwin' else 1024)},indent=2)+'\n')
    if os.waitstatus_to_exitcode(status):raise SystemExit('baseline failed')
