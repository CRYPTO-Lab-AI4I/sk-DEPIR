#!/usr/bin/env python3
"""Record a timestamped hardware/power observation without changing settings."""
import json
import subprocess
import sys
from datetime import datetime, timezone
from pathlib import Path

commands = [['/usr/sbin/sysctl','-n','machdep.cpu.brand_string','hw.memsize','hw.logicalcpu'],
            ['/usr/bin/pmset','-g','batt'], ['/usr/bin/pmset','-g','therm']]
result = {'timestamp_utc': datetime.now(timezone.utc).isoformat(), 'observations': []}
for command in commands:
    run = subprocess.run(command,capture_output=True,text=True,timeout=15)
    result['observations'].append({'command': command, 'exit_code': run.returncode,
                                   'stdout': run.stdout, 'stderr': run.stderr})
Path(sys.argv[1]).write_text(json.dumps(result,indent=2)+'\n')
print(json.dumps(result,indent=2))
