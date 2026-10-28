#!/usr/bin/env python3
"""Keep evidence that the measured native pass loop still calls the batch kernel."""
import hashlib,json,subprocess
from pathlib import Path
ROOT=Path(__file__).resolve().parents[2];HERE=Path(__file__).resolve().parent
r=json.loads((HERE/'comparison/results.json').read_text());name=next(n for n in r['binary_sha256'] if '/cpp-bench-' in n);binary=ROOT/name
assert hashlib.sha256(binary.read_bytes()).hexdigest()==r['binary_sha256'][name]
lines=subprocess.check_output(['/usr/bin/otool','-tvV',str(binary)],text=True).splitlines()
excerpt=[];found=False
for i,line in enumerate(lines):
 if '\tbl\t' in line and 'batch' in line:
  excerpt.extend(lines[max(0,i-6):i+5]);excerpt.append('')
  if any('b.ne' in row for row in lines[i+1:i+5]):found=True
assert found,'No retained batch call in a counted loop'
(HERE/'comparison/cpp-pass-loop.txt').write_text('$ otool -tvV '+name+'\nBinary SHA-256: '+r['binary_sha256'][name]+'\n\n'+'\n'.join(excerpt)+'\n')
print('Measured native binary retains batch calls inside the pass loop.')
