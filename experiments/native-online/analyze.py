#!/usr/bin/env python3
"""Verify measured sources/results and losslessly archive large native samples."""
import gzip
import hashlib
import json
import shutil
import statistics
import sys
from pathlib import Path

ROOT = Path(__file__).resolve().parents[2]
HERE = Path(__file__).resolve().parent
sys.path.insert(0, str(ROOT / 'experiments'))
from source_evidence import verify_sources


def digest(data):
    return hashlib.sha256(data).hexdigest()


def file_digest(path):
    value = hashlib.sha256()
    with path.open('rb') as stream:
        for block in iter(lambda: stream.read(1 << 20), b''):
            value.update(block)
    return value.hexdigest()


def main():
    summaries = {}
    for out in sorted(HERE.iterdir()):
        if not out.is_dir() or not (out / 'results.json').exists():
            continue
        record = json.loads((out / 'results.json').read_text())
        assert record['status'] == 'passed', out
        resolved = verify_sources(ROOT, record['source_sha256'], record['revision'])
        binary = ROOT / record['binary']['path']
        assert digest(binary.read_bytes()) == record['binary']['sha256']
        archive = {}
        for name, expected in record['evidence_sha256'].items():
            path = out / name
            if name.endswith('-native.txt'):
                # Preserve original bytes locally and commit a much smaller,
                # lossless gzip file. The original results manifest stays intact.
                compressed = out / (name + '.gz')
                if path.exists():
                    raw = path.read_bytes()
                    assert digest(raw) == expected
                    local = binary.parent / 'native-samples' / out.name / name
                    local.parent.mkdir(parents=True, exist_ok=True)
                    shutil.copyfile(path, local)
                    compressed.write_bytes(gzip.compress(raw, mtime=0))
                    path.unlink()
                assert digest(gzip.decompress(compressed.read_bytes())) == expected
                archive[name] = {'gzip': compressed.name, 'sha256': digest(compressed.read_bytes()),
                                 'original_sha256': expected, 'retained_raw': str((binary.parent / 'native-samples' / out.name / name).relative_to(ROOT))}
            else:
                assert digest(path.read_bytes()) == expected, path
        if 'input_cache' in record:
            cache = ROOT / record['input_cache']['path']
            assert cache.stat().st_size == record['input_cache']['bytes']
            assert file_digest(cache) == record['input_cache']['sha256']
        phases = record['phases']
        encodes = [x for x in phases if x['phase'].startswith('encode-')]
        verifies = [x for x in phases if x['phase'].startswith('verify-')]
        assert encodes and len(encodes) == len(verifies)
        expected_calls = 1 if record['backend'] == 'native' else 5
        for case in encodes:
            assert case['native_stages']['calls'] == 5
            assert case['native_stages']['boundary_calls'] == expected_calls
        assert all(x['full_hash_match'] for x in verifies)
        hashes = {x['canonical_sha256'] for x in verifies}
        assert len(hashes) == 1
        if not record['smoke']:
            assert hashes == {'b828d4dda538e95b78d914dce52fbfa7fc29a139e4781b0dad83f061f3e26d6c'}
        summary = {'backend': record['backend'], 'encoding_s': [x['elapsed_ns'] / 1e9 for x in encodes],
                   'median_s': statistics.median(x['elapsed_ns'] / 1e9 for x in encodes),
                   'native_stage_sum_s': [sum(v for key,v in x['native_stages'].items() if key.endswith('_ns'))/1e9 for x in encodes],
                   'timers_consistent': [sum(v for key,v in x['native_stages'].items() if key.endswith('_ns')) <= x['elapsed_ns'] * 1.01 for x in encodes],
                   'peak_rss_gib': record['process_peak_rss_bytes'] / 2**30,
                   'canonical_sha256': next(iter(hashes)),
                   'mean_stage_s': {key: statistics.mean(x['native_stages'][key] / 1e9 for x in encodes)
                                    for key in ['taylor_ns', 'columns_transposes_ns', 'rows_ns', 'buffer_ns', 'pointwise_ns', 'assembly_ns']}}
        summaries[out.name] = summary
        (out / 'analysis.json').write_text(json.dumps({'summary': summary, 'sources_resolved': resolved, 'native_sample_archive': archive}, indent=2) + '\n')
    if 's14-native-awake' in summaries and 's14-hybrid-awake' in summaries:
        native, hybrid = summaries['s14-native-awake'], summaries['s14-hybrid-awake']
        consistent = all(native['timers_consistent']) and all(hybrid['timers_consistent'])
        summaries['comparison'] = {'timers_consistent': consistent, 'descriptive_median_ratio': hybrid['median_s'] / native['median_s'],
                                   'median_time_reduction_percent': 100 * (1 - native['median_s'] / hybrid['median_s'])}
    verification = {}
    for path in sorted(HERE.glob('*/verification.json')):
        evidence = json.loads(path.read_text())
        assert evidence['status'] == 'passed', path
        resolved = verify_sources(ROOT, evidence.get('source_sha256', {}), evidence.get('revision', 'HEAD'))
        verification[path.parent.name] = {'status': evidence['status'], 'sources_resolved': resolved}
    (HERE / 'verification-analysis.json').write_text(json.dumps(verification, indent=2) + '\n')
    (HERE / 'summary.json').write_text(json.dumps(summaries, indent=2) + '\n')
    print(json.dumps(summaries, indent=2))


if __name__ == '__main__':
    main()
