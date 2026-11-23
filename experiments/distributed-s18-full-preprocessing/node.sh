#!/bin/bash
# Runs once per node before numerical work; shared paths must work everywhere.
set -euo pipefail
binary="$1"
library="$2"
config="$3"
profile_dir="$4"
test -x "$binary"
test -r "$library"
test -r "$config"
test -d "$profile_dir"
probe=$(mktemp "$profile_dir/.node-preflight-XXXXXXXX")
trap 'rm -f "$probe"' EXIT
ldd "$binary" > "$probe"
if grep -q 'not found' "$probe"; then cat "$probe"; exit 1; fi
ldd "$library" > "$probe"
if grep -q 'not found' "$probe"; then cat "$probe"; exit 1; fi
"$binary" --config "$config" --check-config
python3 - <<'PY'
from pathlib import Path
import os,json,shutil
available=int(next(line.split()[1] for line in Path('/proc/meminfo').read_text().splitlines() if line.startswith('MemAvailable:')))*1024
assert available >= 832*(1<<30), 'less than832GiB node RAM available before allocation'
assert shutil.disk_usage('/tmp').free >= 8*(1<<30), 'insufficient node-local scratch'
print(json.dumps(dict(node=os.uname().nodename,available_host_bytes=available,cuda_visible_devices=os.environ.get('CUDA_VISIBLE_DEVICES'))),flush=True)
PY
