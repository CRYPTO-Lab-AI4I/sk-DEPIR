# Guarded s16 job launched; user will request the next check

- Timestamp: 2026-11-09 23:27:11 UTC
- Agent: Codex
- Work ID: SCALE-012
- Status: handoff (job continues; interactive monitoring stopped at user request)

Completed the source-based audit, budget selection and small admission validation.
[Audit](../experiments/hpc-s16/audit.json), [policy](../experiments/hpc-s16/policy.json)
and [report](../experiments/hpc-s16/report.md) distinguish analytical payloads from
unmeasured large-case feasibility. Estimated principal online arrays are 74.28 GiB;
the conservative preprocessing scenario is 96.56 GiB before runtime/GC overhead.

Admission job 72766 passes both full suites, race/vet and tiny first-use/fresh-cache
checks. Support, online P/spectrum, G and complete output digests match the retained
s6 reference. [Admission analysis](../experiments/hpc-s16/runs/validation-01/analysis.json)
verifies 240 sources, sealed evidence, binary and cache hashes. Controller unit,
syntax and documentation-link checks pass; runtime sources remained unchanged
between admission and submission.

Launched Slurm job **72768**, last confirmed running on gnode02 with 32 CPUs,
256 GiB requested memory and a 24-hour limit. Last observed worker phase was
discovery; exact K, preparation and encoding outcomes remain unverified.
The controller retains its 192 GiB Go soft target, 224 GiB sampled RSS guard,
64 GiB cache cap, 128 GiB scratch cap, node-memory checks and deadlines.
It runs a fresh cached process only after initial success, seals evidence and
writes a summary plus a result-board message. No automatic retry or cache deletion.

The user explicitly requested no monitoring and will ping when jobs finish.
Stop interactive polling; leave the job's resource guards and automatic recording
active. Do not cancel, pause, restart or query the job until requested.
On the next user ping, inspect terminal evidence with
`python3 -B experiments/hpc-s16/analyze.py first-01` and verify completion.

Changed paths: experimental preparation shape guard, shared/private first-use
harness and new s16 test driver; `experiments/hpc-s16/` audit/controller/policy/
tests/report and admission evidence; measured source snapshots; root/package/
experiment READMEs, s16/handoff documentation and this board item. The harness
streams support without a second K-entry array and uses the versioned online
writer; metadata is atomic and synced. Default public discovery remains s<=13.
Completed code stages persist on shared NFS under
`.cache/liftrs/hpc-s16/s16-m2-t5/`; source-node and earlier caches are untouched.

README check completed before launch: experimental admission is distinct from
demonstrated s16 feasibility, application/cryptographic integration and TRL 5.
No large-case success, exact capacity or performance claim is added. Preserve
all live files; the running controller may append results after this handoff.
