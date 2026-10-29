#!/usr/bin/env python3
"""Read-only RSS samples after preparation release, without repeating encoding."""
import argparse,hashlib,json,subprocess,time
from datetime import datetime,timezone
from pathlib import Path
p=argparse.ArgumentParser(description=__doc__);p.add_argument('--pid',type=int,required=True);p.add_argument('--run',type=Path,required=True);p.add_argument('--out',type=Path,required=True);a=p.parse_args();a.out.mkdir(exist_ok=False)
manifest={'script_sha256':hashlib.sha256(Path(__file__).read_bytes()).hexdigest(),'pid':a.pid,'run':str(a.run),'method':'ps rss in KiB every second; sampled RSS, not an OS peak counter; begins during first encoding after explicit preparation release','samples':0,'peaks_by_phase_bytes':{}}
start=time.monotonic()
with (a.out/'samples.jsonl').open('w') as f:
 while time.monotonic()-start<600:
  r=subprocess.run(['/bin/ps','-o','rss=','-p',str(a.pid)],capture_output=True,text=True)
  if r.returncode or not r.stdout.strip():break
  phase=json.loads((a.run/'state.json').read_text())['phase'];rss=int(r.stdout.strip())*1024
  f.write(json.dumps({'timestamp_utc':datetime.now(timezone.utc).isoformat(),'phase':phase,'rss_bytes':rss})+'\n');f.flush();manifest['samples']+=1;manifest['peaks_by_phase_bytes'][phase]=max(manifest['peaks_by_phase_bytes'].get(phase,0),rss)
  time.sleep(1)
manifest['elapsed_s']=time.monotonic()-start;manifest['samples_sha256']=hashlib.sha256((a.out/'samples.jsonl').read_bytes()).hexdigest();(a.out/'results.json').write_text(json.dumps(manifest,indent=2)+'\n');print(json.dumps(manifest['peaks_by_phase_bytes']))
