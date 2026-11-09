#!/usr/bin/env python3
"""Verify sealed baseline evidence and summarize phase/resource observations."""
import json
from pathlib import Path
import re
import runpy
import sys

HERE = Path(__file__).resolve().parent
ROOT = HERE.parents[1]
sys.path.insert(0, str(ROOT / 'experiments'))
from source_evidence import verify_sources


def main():
    if len(sys.argv) != 2:
        raise SystemExit('Usage: analyze.py RUN_NAME')
    out = HERE / 'runs' / sys.argv[1]
    record = json.loads((out / 'results.json').read_text())
    assert record['status'] == 'passed'
    module = runpy.run_path(str(HERE / 'run.py'))
    sha = module['sha']
    for name, expected in record['evidence_sha256'].items():
        assert sha(out / name) == expected, name
    sources = verify_sources(ROOT, record['source_sha256'], record['revision'])
    durable = Path(record['durable_cache'])
    assert sha(durable / 'probe') == record['binary_sha256']
    assert sha(durable / 'word-packed.bin') == record['retained_output_sha256']
    reference = ROOT / '.cache/liftrs/parallel-s15/s15-m2-t5/word-metadata.json'
    phases = json.loads((out / 'phases.json').read_text())
    module['verify'](phases, json.loads(reference.read_text()))
    encode = next(p for p in phases if p['phase'] == 'encode-0')
    native_ns = encode['native_after']['steady_ns'] - encode['native_before']['steady_ns']
    stats = encode['native_stages']
    total = sum(stats[k] for k in ('buffer_ns', 'pointwise_ns', 'assembly_ns',
                                 'taylor_ns', 'columns_transposes_ns', 'rows_ns'))
    assert total <= native_ns + 1000000, 'native stages exceed enclosing clock'
    log = (out / 'cached-encoding.log').read_text()
    rss_kib = int(re.search(r'Maximum resident set size \(kbytes\): (\d+)', log)[1])
    result = {'status': 'passed', 'job_id': record['job_id'],
              'source_files_verified': len(sources), 'sources_resolved': sources,
              'phase_seconds': {p['phase']: p['elapsed_ns'] / 1e9 for p in phases},
              'probe_peak_rss_gib': rss_kib / 1024**2,
              'native_enclosing_seconds': native_ns / 1e9,
              'native_stage_seconds': total / 1e9,
              'stage_and_verify_seconds': record['stage_and_verify_seconds'],
              'output_retention_seconds': record['output_retention_seconds'],
              'validation': record['validation']}
    (out / 'analysis.json').write_text(json.dumps(result, indent=2) + '\n')
    print(json.dumps({k: v for k, v in result.items() if k != 'sources_resolved'}, indent=2))


if __name__ == '__main__':
    main()
