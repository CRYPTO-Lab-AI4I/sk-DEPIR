#!/usr/bin/env python3
"""Verify retained evidence and regenerate pprof views without rerunning encoding."""
import argparse,hashlib,json,re,subprocess
from pathlib import Path
ROOT=Path(__file__).resolve().parents[2];HERE=Path(__file__).resolve().parent

def sha(path):
    h=hashlib.sha256()
    with path.open('rb') as file:
        for block in iter(lambda:file.read(1<<20),b''):h.update(block)
    return h.hexdigest()

def rows(path):
    result={}
    for line in path.read_text().splitlines():
        match=re.match(r'^\s*(\S+)\s+([\d.]+)%\s+[\d.]+%\s+(\S+)\s+([\d.]+)%\s+(.+)$',line)
        if match:result[match[5]]={'flat':match[1],'flat_percent':float(match[2]),'cumulative':match[3],'cumulative_percent':float(match[4])}
    return result

def main():
    parser=argparse.ArgumentParser(description=__doc__);parser.add_argument('--go',default='go');args=parser.parse_args()
    all_profiles={}
    for name in ['setup','reuse-01']:
        directory=HERE/name;r=json.loads((directory/'results.json').read_text())
        for source,h in r['source_sha256'].items():
            path=directory/'run-snapshot.py' if name=='setup' and source=='experiments/encoder-profile-s13/run.py' else ROOT/source
            assert sha(path)==h,(name,source)
        binary=ROOT/r['profile_binary']['path'];assert sha(binary)==r['profile_binary']['sha256']
        for field in ['events','stderr','progress']:assert sha(directory/r[field])==r[field+'_sha256']
        for phase,event in r['phases'].items():
            # Only completed phases have valid profiles. The setup monitor failure
            # left a zero-byte encode_0 profile; preserve but do not analyze it.
            profile=directory/(phase+'.pprof');assert profile.stat().st_size>0
            texts={}
            for tag,options in [('flat',['-top']),('cum',['-top','-cum']),('tree',['-tree'])]:
                path=directory/(phase+'-'+tag+'.txt')
                with path.open('w') as output:subprocess.run([args.go,'tool','pprof',*options,str(binary),str(profile)],stdout=output,stderr=subprocess.STDOUT,check=True)
                texts[path.name]=sha(path)
            if phase in ['preparation','encode_0']:
                path=directory/(phase+'-source.txt')
                with path.open('w') as output:subprocess.run([args.go,'tool','pprof','-list','recurse|polyMul|constantProduct',str(binary),str(profile)],stdout=output,stderr=subprocess.STDOUT,check=True)
                texts[path.name]=sha(path)
            all_profiles[name+'/'+phase]={'profile':str(profile.relative_to(ROOT)),'sha256':sha(profile),'texts':texts,'cpu_rows':rows(directory/(phase+'-flat.txt'))}
    reused=json.loads((HERE/'reuse-01/results.json').read_text());cache=ROOT/reused['cache_directory']
    assert sha(cache/'cyclic.json')==reused['cache_metadata']['sha256']
    baseline=json.loads((ROOT/'experiments/cyclic-s13/results.json').read_text())
    for key in ['canonical_sha256','packed_sha256']:assert reused['verification'][key]==baseline['verification'][key]
    assert reused['baseline_match'] and reused['status']=='completed'
    setup=json.loads((HERE/'setup/results.json').read_text())
    lines=['# s=13 profiles and retained cache','','Profiled times include CPU sampling. Setup and encoding below come from separate runs after a monitor failure; there is no single combined process/RSS measurement.','','| Stage | Run | Measured time |','|---|---|---:|']
    for phase in ['discovery','preparation','cache_save','cache_load']:
        lines.append(f'| {phase} | setup | {setup["phases"][phase]["elapsed_ns"]/1e9:.3f} s |')
    for phase in ['cache_load','encode_0']:
        lines.append(f'| {phase} | reuse-01 | {reused["phases"][phase]["elapsed_ns"]/1e9:.3f} s |')
    lines+=['',f'Cache: {reused["cache_metadata"]["bytes"]:,} bytes ({reused["cache_metadata"]["bytes"]/2**20:.2f} MiB), retained under `{reused["cache_directory"]}`.',f'Reuse-only process peak RSS: {reused["process_peak_rss_bytes"]/2**30:.3f} GiB. Setup process peak RSS was not recovered after the monitor error.','', '## CPU sample shares','','Flat shares are self CPU samples. Cumulative shares include callees and overlap; neither is a fraction of all wall time.','','| Phase | Function | Flat | Cumulative |','|---|---|---:|---:|']
    for phase in ['setup/preparation','reuse-01/encode_0']:
        for function,row in all_profiles[phase]['cpu_rows'].items():
            if any(x in function for x in ['constantProduct.mul','BinaryField.Mul','additiveTransform).recurse','polyDiv','polyMulLow']):
                lines.append(f'| {phase} | `{function.split("liftrs.")[-1]}` | {row["flat_percent"]:.2f}% | {row["cumulative_percent"]:.2f}% |')
    (HERE/'analysis.json').write_text(json.dumps({'profiles':all_profiles,'cache_verified':True,'baseline_match':True},indent=2)+'\n')
    (HERE/'table.md').write_text('\n'.join(lines)+'\n');print('\n'.join(lines))
if __name__=='__main__':main()
