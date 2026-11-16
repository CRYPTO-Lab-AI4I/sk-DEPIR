# SCALING-082 — Exact counts pass; full resource plan reviewable

- UTC: 2026-11-16T10:09:14.152149+00:00
- Agent: Codex
- Status: Handoff (awaiting numeric budget agreement)

75810 passes ongnode06 in245.873s includingbuilds/checks. Analyzer reconciles
513source/41evidence/85cache identities. All workflows/references, race/ASan/vet
and full/count-only CUDA memcheck pass, zero errors/leaks. Exact dimensions:
s8K7381,s16K918672421,s17K3876566203,N17179869184. Count-only s17 CLI6.276s
includes setup/profile/metadata, not discovery/preparation latency. All cache/
profiles retained. [Analysis](../experiments/s17-preprocessing-admission/runs/counts-01/analysis.json),
[count report](../experiments/s17-preprocessing-admission/report.md#completed-count-admission-scaling-082).

[Exact sizes](../experiments/s17-preprocessing-admission/s17-sizes.json): support
28.883GiB, root14.441GiB, canonical142.441GiB,G49.559GiB,total235.324GiB. Corrected
conservative discovery array model includes both possible2GiB bitmaps=189.765GiB;
final retained support/P/G/spectrum220.883GiB. Neither is peak RSS; Newton/root/
identity/GC/driver/cache costs additional and unknown.

Prepared [full controller/report](../experiments/gpu-s17-full-preprocessing/report.md)
and budget.json: exclusive128cores/32workers,864GiBrequest/800GiBRSS/720GiBsoft,
4H200112GiBowned/card,512GiBretained/16GiBscratch,45hworker/46hcontroller/48hSlurm.
No fixed node; excludegnode03. Explicitlocalcandidate adds extra transposes;
peerdefault untouched, full speedup unknown.8/16GPU options still require cross-node
work. Budget approved=false/no submission. Mandatory full resource agreement is
the final gate, rather than assuming capacity or prior small approvals suffice.

Controller reuses validated CLI/peer probe/boundedstop/RSS/filesystem helpers,
checks count, snapshots exact budget/admission JSON, guards resources/headroom/
time and logs current worker phases. Avoid hidden235GiB exit rehash; CLI full
identity/readback required, later independent raw rehash separately admitted.
Startup availability and128CPU/864GiB allocation verified before large work.
No useful checkpoint/partial/profile deletion or slower memory mode selected.

Long host ordering/Frobenius, normalization and reused-root scans now emit
advancing65536-element chunk counters with existing30s throttling. Same arithmetic,
ordering/closure/field checks and buffers, no new pass/large array. Existing native
Go-oracle/restart/count checks and new cross-chunk progress/duplicate test pass;
full controller plan/refusal/log-tail tests, Python AST/check-mode,bash,gofmt/diff
checks pass. [Evidence](../experiments/gpu-s17-full-preprocessing/local-checks.json).
GPU code kernels unchanged; no full-run peak/latency claimed.

README checked/updated, preprocessing guide/scalingplan/reference and reports
updated for exact component results and pending full run. No new job submitted.
Request user agreement to the concrete full-node RAM/GPU-owned/storage/time/CPU
budget, explaining AGENTS.md cap rule. Once approved, record it and submit full
run, then hand off without monitoring. CPU setup investigation remains deferred.
