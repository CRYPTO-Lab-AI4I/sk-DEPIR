#!/usr/bin/env python3
"""Compare language-labelled kernels and an actual-size additive FFT on shared data."""
import argparse,hashlib,json,os,platform,shutil,statistics,subprocess,time
from datetime import datetime,timezone
from pathlib import Path
ROOT=Path(__file__).resolve().parents[2];HERE=Path(__file__).resolve().parent

def sha(path):
    h=hashlib.sha256()
    with path.open('rb') as file:
        for block in iter(lambda:file.read(1<<20),b''):h.update(block)
    return h.hexdigest()

def main():
    p=argparse.ArgumentParser(description=__doc__);p.add_argument('--go',default='go');p.add_argument('--cxx',default='clang++');p.add_argument('--smoke',action='store_true');p.add_argument('--run-name');args=p.parse_args()
    cache=ROOT/'.cache/liftrs/language-s13';cache.mkdir(parents=True,exist_ok=True)
    name=args.run_name or ('smoke' if args.smoke else 'comparison')
    if any(c not in 'abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789-_' for c in name):raise SystemExit('invalid run name')
    directory=HERE/name;directory.mkdir(exist_ok=False)
    sources=(sorted((ROOT/'encoding/liftrs-shared').rglob('*.go'))+list((ROOT/'encoding/liftrs-go').glob('*.go')))+sorted((ROOT/'encoding/liftrs-cpp').glob('*.*'))+[HERE/'run.py',ROOT/'go.mod',ROOT/'go.sum']
    result={'timestamp_utc':datetime.now(timezone.utc).isoformat(),'revision':subprocess.check_output(['git','rev-parse','HEAD'],cwd=ROOT,text=True).strip(),'source_sha256':{str(path.relative_to(ROOT)):sha(path) for path in sources},
        'environment':{'platform':platform.platform(),'cpu_count':os.cpu_count(),'go':subprocess.check_output([args.go,'version'],text=True).strip(),'cxx':subprocess.check_output([args.cxx,'--version'],text=True)},'cpp_flags':['-std=c++17','-O3','-g','-fno-omit-frame-pointer'],'cases':[]}
    def save():(directory/'results.json').write_text(json.dumps(result,indent=2)+'\n')
    save()
    if not (cache/'transform-input.bin').exists():
        env=dict(os.environ,PIR_LANGUAGE_FIXTURES=str(cache))
        subprocess.run([args.go,'test','./encoding/liftrs-shared','-run','^TestExportLanguageFixtures$','-count=1'],cwd=ROOT,env=env,check=True)
    result['fixture_sha256']={p.name:sha(p) for p in cache.glob('*.bin')}
    go_binary=cache/('go-bench-'+name+'.test');cpp_binary=cache/('cpp-bench-'+name)
    subprocess.run([args.go,'test','-c','-o',str(go_binary),'./encoding/liftrs-shared'],cwd=ROOT,check=True)
    with (directory/'compiler.log').open('w') as log:
        subprocess.run([args.cxx,*result['cpp_flags'],'-Rpass=loop-vectorize','-Rpass-missed=loop-vectorize',str(ROOT/'encoding/liftrs-cpp/bench.cpp'),'-o',str(cpp_binary)],cwd=ROOT,stdout=log,stderr=subprocess.STDOUT,check=True)
    result['binary_sha256']={str(x.relative_to(ROOT)):sha(x) for x in [go_binary,cpp_binary]};save()
    cases=[('kernel','base','fixed',1024),('kernel','extension','rotating',1024),('forward','extension','fixed',16)] if args.smoke else [( 'kernel',kind,pattern,size) for kind in ['base','extension'] for pattern in ['fixed','rotating'] for size in [1<<18,1<<22]]+[('forward','extension','fixed',1<<20),('forward','extension','fixed',1<<26)]
    for mode,kind,pattern,size in cases:
        label=f'{mode}-{kind}-{pattern}-{size}';record={'label':label,'mode':mode,'kind':kind,'pattern':pattern,'size':size,'implementations':{}};print(label,flush=True)
        for language in ['go','cpp']:
            data_dir=cache/name/label/language;data_dir.mkdir(parents=True,exist_ok=False)
            view_dir=directory/label/language;view_dir.mkdir(parents=True,exist_ok=True)
            log=view_dir/'process.log';start=time.monotonic()
            if language=='go':
                env=dict(os.environ,PIR_LANGUAGE_FIXTURES=str(cache),PIR_LANGUAGE_MODE=mode,PIR_LANGUAGE_CASE=kind,PIR_LANGUAGE_PATTERN=pattern,PIR_LANGUAGE_SIZE=str(size),PIR_LANGUAGE_OUTPUT=str(data_dir))
                with log.open('w') as output:subprocess.run([str(go_binary),'-test.run=^TestRunLanguageBench$','-test.count=1'],cwd=ROOT,env=env,stdout=output,stderr=subprocess.STDOUT,check=True,timeout=7200)
                for profile in data_dir.glob('*.pprof'):
                    shutil.copy2(profile,view_dir/profile.name)
                    with (view_dir/(profile.stem+'-flat.txt')).open('w') as output:subprocess.run([args.go,'tool','pprof','-top',str(go_binary),str(profile)],stdout=output,stderr=subprocess.STDOUT,check=True)
            else:
                command=[str(cpp_binary),str(cache),mode,kind,pattern,str(size),str(data_dir/'output.bin'),str(data_dir/'result.json')]
                with log.open('w') as output:
                    proc=subprocess.Popen(command,cwd=ROOT,stdout=subprocess.PIPE,stderr=output,text=True)
                    ready=proc.stdout.readline();output.write(ready);output.flush()
                    if ready.strip()!='READY':proc.kill();proc.wait();raise RuntimeError('native driver failed before timing')
                    sampler=None;sample_log=(view_dir/'sample.log').open('w')
                    if not args.smoke and shutil.which('sample'):
                        try:sampler=subprocess.Popen(['sample',str(proc.pid),'10','10','-mayDie','-file',str(view_dir/'cpu-sample.txt')],stdout=sample_log,stderr=subprocess.STDOUT)
                        except PermissionError as error:record['native_profile_exception']=repr(error)
                    try:proc.wait(timeout=7200)
                    except BaseException:proc.kill();proc.wait();raise
                    if sampler:
                        try:sampler.wait(timeout=15)
                        except subprocess.TimeoutExpired:sampler.terminate();sampler.wait()
                        record['native_sampler_returncode']=sampler.returncode
                    sample_log.close()
                    if proc.returncode:raise RuntimeError('native driver failed; see '+str(log))
            data=json.loads((data_dir/'result.json').read_text());data['sha256']=sha(data_dir/'output.bin');data['wall_seconds']=time.monotonic()-start;data['output_directory']=str(data_dir.relative_to(ROOT));data['median_ns']=statistics.median(data['elapsed_ns'])
            record['implementations'][language]=data
            print(f'  {language}: {data["median_ns"]/1e9:.4f}s',flush=True)
        a,b=record['implementations']['go'],record['implementations']['cpp'];assert a['sha256']==b['sha256'],label+' output differs';assert a['passes']==b['passes']
        record['cpp_speedup']=a['median_ns']/b['median_ns'];record['outputs_equal']=True;result['cases'].append(record);save()
    result['status']='completed';save()
if __name__=='__main__':main()
