# GPU-038 — Parity correction and retry submitted

Timestamp: 2026-11-12T17:51:29.390636+00:00
Agent: Codex
Status: handoff

Job74091 failed its s8 resident packed oracle after4m37s, before any matched/full
runs. Scalar/matrix/build and stream-boundary checks passed. Failed run/source/
evidence/cache identities reconcile366 exact sources; logs/profiles/library remain
unchanged. See failure-review.json and accounting in runs/packed-01. This is a
failed validation, not new s16 memory/time evidence.

Root cause: canonical_packed_zero reduced chunk parities into shared sums[0],
but stored thread0 local sum. Corrected the final atomic OR to sums[0], matching
established owning parity reduction. Added independent s8 K1 constant-word test
spanning16 chunks, with first chunk parity0 and nonzero joined parity. Both
resident/streamed layouts test values1/255/0/17, whole packed bytes and memcheck;
small oracle diagnostics identify repeat/first byte/counters. Deferred engine
cleanup now applies on failed tests too. Ordinary owning/default kernels unchanged.

Fixed sm90 compile-only passes with retained fixed-provenance.json; combined Go
admission/lifecycle/accessor checks pass0.017 s. Python sink/metadata rejection,
explicit local/global parity fixture, AST/bash/whitespace checks pass. Updated
README/report/scaling plan with failure and pending retry, README checked. No
claim of runtime success until fresh validation. Existing failed proof preserved.

Submitted **job74094**, fresh `packed-02`, same explicit streamed/packed options
and oneH20016CPU/40 GiB GPU/128 GiB host/96 GiB RSS/cache/2h budgets. Existing
small gates run first, including new parity regression and memcheck; only then
matched s14 and full s16 pipelines/checks. No budget increase or default change.
Results `experiments/gpu-packed-output/runs/packed-02/results.json`;
log `experiments/gpu-packed-output/slurm-74094.log`;
cache `.cache/liftrs/gpu-packed-output/packed-02/` outside Git.
Estimate20–45 minutes after scheduling, new atomic/sink runtime still unknown.

No monitoring/polling/log following after submission; wait for user completion
ping. Original task remains active: validate GPU output, then implement/validate
GPU host-backed transforms, then CPU output and CPU disk-backed transforms.
Dependent gates must not skip this failure; do not stop at reporting future pass
without continuing the authorized next gate. No s17/cryptographic/TRL5 claim.
Code/evidence/docs staged, uncommitted; new live artifacts await completion ping.
