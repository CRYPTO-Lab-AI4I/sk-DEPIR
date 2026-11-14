#!/usr/bin/env python3
"""Reconcile completed CPU modularity evidence and matched latency samples."""
import argparse
import hashlib
import json
from pathlib import Path
import re
import runpy
import statistics
import subprocess

HERE = Path(__file__).resolve().parent
VERIFY = runpy.run_path(str(HERE.parent/'s17-arithmetic/analyze.py'))['reconcile']
PATTERN = re.compile(r'^BenchmarkCPUEncodingModules/([a-z-]+)-\d+\s+\d+\s+([\d.]+) ns/op\s+(\d+) B/op\s+(\d+) allocs/op', re.MULTILINE)


def main():
    parser=argparse.ArgumentParser(description=__doc__)
    parser.add_argument('name')
    args=parser.parse_args()
    out=HERE/'runs'/args.name
    record=json.loads((out/'results.json').read_text())
    result=VERIFY(out)
    # Verify the baseline against Git, except the identical new benchmark file.
    benchmark='encoding/liftrs-shared/cpu_modules_test.go'
    identities=record['baseline_sha256']
    assert identities[benchmark]==record['source_sha256'][benchmark]
    names=[n for n in identities if n!=benchmark]
    requests=''.join(f"{record['baseline_revision']}:{n}\n" for n in names)
    data=subprocess.check_output(['git','cat-file','--batch'],input=requests.encode())
    position=0
    for name in names:
        end=data.index(b'\n',position);header=data[position:end].split()
        assert header[1]==b'blob',name
        size=int(header[2]);position=end+1
        assert hashlib.sha256(data[position:position+size]).hexdigest()==identities[name],name
        position+=size+1
    for label in ('modules','asan'):
        log=(out/(label+'.log')).read_text()
        for test in ('TestCPUEncodingModules','TestCPUWorkspaceLifetime','TestScalingWideCPPEncoding','TestScalingWideCPPLifetime'):
            assert '--- PASS: '+test+' ' in log,(label,test)
    modes={'resident','hybrid','wide','compact','file','disk','packed','load-resident','open-file'}
    samples={version:{mode:[] for mode in modes} for version in ('old','new')}
    for repeat in range(3):
        for version in ('old','new'):
            log=(out/f'timing-{repeat}-{version}.log').read_text()
            rows=PATTERN.findall(log);assert {r[0] for r in rows}==modes
            assert len(rows)==len(modes) and log.rstrip().endswith('PASS')
            for mode,ns,bytes_,allocs in rows:
                samples[version][mode].append({'ns':float(ns),'bytes':int(bytes_),'allocs':int(allocs)})
    comparisons={}
    for mode in sorted(modes):
        old=statistics.median(x['ns'] for x in samples['old'][mode])
        new=statistics.median(x['ns'] for x in samples['new'][mode])
        comparisons[mode]={'old_median_seconds':old/1e9,'new_median_seconds':new/1e9,
                           'new_over_old':new/old,'samples':{v:samples[v][mode] for v in samples},
                           'needs_review':new/old>1.2}
    for label in ('helpers-optimized','helpers-sanitized'):
        rows=[json.loads(l) for l in (out/(label+'.log')).read_text().splitlines() if l.startswith('{')]
        assert rows[-1]=={'helpers':True,'high_values':4096,'file_words':16389}
        progress=[r for r in rows if r.get('work')=='self-check-units']
        assert [r['completed'] for r in progress]==[1,2] and all(r['total']==2 for r in progress)
    result.update(baseline_files_verified=len(names),tests=['all CPU modes and lifetimes','Go/native regression suites','vet/race','Go ASan/C++ UBSan','direct helper/online/matrix/Taylor/wide oracles'],
                  native_progress_records=record['native_progress_records'],comparisons=comparisons,
                  scope='bounded CPU correctness and small s10 timing comparison; no fulls16 performance equivalence or s17 preparation/encoding')
    (out/'analysis.json').write_text(json.dumps(result,indent=2)+'\n')
    print(json.dumps({k:result[k] for k in ('status','job_id','source_files_verified','evidence_files_verified','cache_files_verified','baseline_files_verified')}))
    for mode,value in comparisons.items():print(mode,round(value['new_over_old'],3),'review' if value['needs_review'] else '')


if __name__=='__main__':main()
