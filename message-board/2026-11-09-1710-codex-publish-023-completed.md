# HPC checkpoint verified for publication

- Timestamp: 2026-11-09 17:10:00 UTC
- Agent: Codex
- Work ID: PUBLISH-023
- Status: completed (publication checks; commit and push follow)

Reviewed all pending orientation/setup/HPC work for the user's requested commit
and push. Publish the [HPC experiment](../experiments/hpc-cached-s15/report.md),
its runnable scripts, raw successful/partial evidence, profiles and three exact
source snapshots; [HPC setup](../docs/hpc-setup.md),
[s16 audit](../docs/encoding-s16.md), updated handoff/scaling documentation,
root/experiment READMEs and all associated board receipts. Large ignored caches,
toolchains, matching binaries and packed output remain outside Git at their
documented paths. No arithmetic or supported-parameter changes are included.

Repeated `python3 -B experiments/hpc-cached-s15/analyze.py baseline-02`: passes
all sealed evidence, binary/output hashes, 227 recorded sources and native-clock
containment. Both runner unit checks, shell syntax, documentation/message links
and diff whitespace pass. Inspected recorded Slurm variable names for credential
fields; none are present. Full Go/native suites and race/vet already passed on
the compute node; no implementation change requires repeating that allocation.
Root README checked and accurately distinguishes the passed CPU baseline from
unverified GPU execution, s16 support and TRL 5 completion.

Publication adds this pair of messages and their board entries. Final staged
contents/whitespace will be checked before commit, and push will use ordinary
fast-forward semantics. The next work item is the s16 cardinality/header extension
with old-format compatibility and small-buffer/boundary tests, then a preparation
memory and resource-budget review before a full larger run.
