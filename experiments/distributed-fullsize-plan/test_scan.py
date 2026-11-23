#!/usr/bin/env python3
"""Focused archive-boundary tests and independent retained-s10 owner comparison."""
import hashlib
import json
from pathlib import Path
import struct
import subprocess
import sys
import tempfile

HERE = Path(__file__).resolve().parent
ROOT = HERE.parents[1]
binary = str(Path(sys.argv[1]).resolve())
with tempfile.TemporaryDirectory() as directory:
    directory = Path(directory)
    source, result = directory/'support.bin', directory/'result.json'
    def check(values, *, count=None, sha=None, suffix=b'', passes=True):
        data = b''.join(struct.pack('<Q', v) for v in values)+suffix
        source.write_bytes(data)
        result.unlink(missing_ok=True)
        command = [binary, '8', '8', str(len(values) if count is None else count),
                   hashlib.sha256(data).hexdigest() if sha is None else sha,
                   str(source), str(result)]
        completed = subprocess.run(command, capture_output=True, text=True, timeout=5)
        assert (completed.returncode == 0) == passes, completed.stderr
        assert result.exists() == passes
        if passes:
            rows = json.loads(result.read_text())
            for ranks, counts in rows['owners'].items():
                expected = [0]*int(ranks)
                for v in values:
                    expected[min(v, v*256 % 65535) % int(ranks)] += 1
                assert counts == expected
    check([0, 1, 257, 65534])
    check([0, 1], sha='0'*64, passes=False)
    check([0, 1], suffix=b'\x00', passes=False)
    check([1, 1], passes=False)
    check([2, 1], passes=False)
    check([65535], passes=False)
    check([1, 2], count=1, passes=False)
    check([1], count=2, passes=False)
    native = ROOT/'.cache/liftrs/gpu-products/products-02/cpu-0/support.bin'
    subprocess.run([binary, '10', '4', '146641',
                    'da214fb884c9f73a552088db49ed105fd470f9766f0dcf2b90d5b3596ff50966',
                    str(native), str(result)], check=True, capture_output=True, timeout=5)
    rows = json.loads(result.read_text())
    reference = json.loads((HERE/'support-s10.json').read_text())
    assert rows['owners']['16'] == reference['actual_owners16']
    assert all(sum(counts) == 146641 for counts in rows['owners'].values())
print(json.dumps({'status': 'passed', 'synthetic_cases': 8, 'native_s10_comparison': True}))
