#!/usr/bin/env python3
"""Record verified standalone s15 support while retaining MC/paper evidence.

Run analyze.py for first use and cached execution first. This does not promote
application/cryptographic integration or change historical planner snapshots.
"""
import json
from pathlib import Path

HERE = Path(__file__).resolve().parent
ROOT = HERE.parents[1]
first = json.loads((HERE/'initial/analysis.json').read_text())
cached = json.loads((HERE/'cached/analysis.json').read_text())
assert first['status'] == cached['status'] == 'passed'
assert cached['first_use_word_match'] and first['canonical_sha256'] == cached['canonical_sha256']
assert (first['k'], first['n']) == (216914047, 1 << 30)
p = ROOT/'pirconfig/configs_catalogue.json'
catalogue = json.loads(p.read_text())
entry = next(e for e in catalogue['configurations'] if e['id'] == 'binary-s15-m2-t5')
entry['code']['polynomial'] = 0x8003
entry['encoding_status'] = 'supported'
capacity = entry['capacity']
capacity.update(status='verified', k=first['k'], rate=first['k']/first['n'])
evidence = {'kind':'encoding', 'reference':'experiments/parallel-s15/report.md',
            'script':'experiments/parallel-s15/run.py', 'results':'experiments/parallel-s15/initial/results.json',
            'method':'Exact monomial enumeration and support discovery establish K=216914047,N=1073741824. Bounded reusable Go preparation with native transforms verifies P*G and retains support/P/G/spectrum. Three complete C++ encodings pass all systematic positions, 16 independently evaluated curve restrictions, repeated canonical hashes and packed readback; fresh cached execution matches. Preparation was manually paused/resumed and its raw elapsed time is interrupted. Standalone experimental native/minimal-file path only; default public discovery remains s<=13. No application-plan or cryptographic integration.'}
capacity['evidence'] = [e for e in capacity['evidence'] if e['kind'] != 'encoding'] + [evidence]
entry['encoder'] = {'implementation':'encoding/liftrs-shared/scaling_online.go: LoadScalingOnline; scaling_matrix_cpp.go: WithCPPMatrix/EncodeWithStats (complete native online path)',
                    'preprocessing':'Experimental exact discovery and bounded reusable preparation via experiments/parallel-s15/run.py; verified binary support/P/spectrum/G files retained outside Git with checksums. Public DiscoverOrbit/JSON basis limits remain s<=13. Application-plan execution is not integrated.',
                    'evidence':[evidence]}
p.write_text(json.dumps(catalogue, indent=2)+'\n')
print('Recorded exact s15 standalone support; sampling/paper evidence preserved')
