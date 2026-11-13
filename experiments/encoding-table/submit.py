#!/usr/bin/env python3
"""Submit independent cases in two serial resource chains without monitoring."""
import argparse
import datetime
import json
from pathlib import Path
import runpy
import subprocess

HERE = Path(__file__).resolve().parent
ROOT = HERE.parents[1]
RUN = runpy.run_path(str(HERE / 'run.py'))


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('batch')
    parser.add_argument('--dry-run', action='store_true')
    parser.add_argument('--extended-limits', action='store_true')
    parser.add_argument('--scheduling', choices=['serial', 'split', 'independent'], default='serial', help='split runs packed output independently of a serial disk queue')
    parser.add_argument('--after-job', help='wait for the cancelled original to finish terminating')
    args = parser.parse_args()
    if not args.batch.replace('-', '').isalnum(): parser.error('alphanumeric/hyphen batch required')
    path = HERE / 'runs' / args.batch / 'submissions.json'
    manifest = json.loads(path.read_text())
    assert all(v.get('job_id') is None for v in manifest['cases'].values()), 'batch already submitted; do not duplicate'
    cases = list(manifest['cases'])
    assert cases and set(cases) <= set(RUN['CASES'])
    if args.extended_limits: assert set(cases) <= set(RUN['EXTENDED_HOURS'])
    if args.after_job: assert args.after_job.isdigit()
    manifest['revision'] = subprocess.check_output(['git', 'rev-parse', 'HEAD'], cwd=ROOT, text=True).strip()
    manifest['submitted_utc'] = datetime.datetime.now(datetime.timezone.utc).isoformat()
    previous = {}
    for case in cases:
        backend = case.split('-')[0]
        limits = RUN['policy'](case, args.extended_limits)
        chain = None if args.scheduling == 'independent' else ('disk' if '-disk-' in case else case) if args.scheduling == 'split' else backend
        argv = ['sbatch', '--parsable', '--mem=' + str(limits['host_request_bytes'] >> 30) + 'G',
                '--output=experiments/encoding-table/slurm-%j.log']
        if backend == 'gpu': argv.append('--gres=gpu:h200:1')
        argv.append('--time=' + str(limits['slurm_seconds'] // 60))
        dependency = previous.get(chain) if chain is not None else None
        dependency = dependency or args.after_job
        if dependency: argv.append('--dependency=afterany:' + dependency)
        argv.extend(['experiments/encoding-table/run.sbatch', args.batch, case])
        if args.extended_limits: argv.append('--extended-limits')
        if args.dry_run:
            job = 'dry-' + str(len(previous)) + '-' + case
        else:
            job = subprocess.check_output(argv, cwd=ROOT, text=True).strip().split(';')[0]
            assert job.isdigit(), job
        if chain is not None: previous[chain] = job
        manifest['cases'][case] = {'job_id': job, 'command': argv, 'policy': limits}
        if not args.dry_run:
            manifest['status'] = 'submitting'
            path.write_text(json.dumps(manifest, indent=2) + '\n')
        print(case, job, flush=True)
    if not args.dry_run:
        manifest['status'] = 'submitted'
        path.write_text(json.dumps(manifest, indent=2) + '\n')


if __name__ == '__main__':
    main()
