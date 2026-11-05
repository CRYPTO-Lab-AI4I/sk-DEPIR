#!/usr/bin/env python3
"""Observe benchmark power/phase changes without modifying system settings."""
import json
import os
import subprocess
import sys
import time
from datetime import datetime,timezone
from pathlib import Path

pid=int(sys.argv[1]);out=Path(sys.argv[2])
with (out/'power-observations.jsonl').open('a') as stream:
    while True:
        try: os.kill(pid,0)
        except ProcessLookupError: break
        run=subprocess.run(['/usr/bin/pmset','-g','batt'],capture_output=True,text=True,timeout=10)
        try: phase=json.loads((out/'state.json').read_text())['phase']
        except (OSError,ValueError,KeyError): phase='unknown'
        item={'timestamp_utc':datetime.now(timezone.utc).isoformat(),'phase':phase,'command':['/usr/bin/pmset','-g','batt'],'exit_code':run.returncode,'stdout':run.stdout,'stderr':run.stderr}
        stream.write(json.dumps(item)+'\n');stream.flush()
        print(item['timestamp_utc'],phase,run.stdout.splitlines()[0] if run.stdout else run.stderr,flush=True)
        time.sleep(15)
