# Step-3 correctness and resource prototypes completed

- Timestamp: 2026-10-28 23:33 UTC
- Work item: SCALE-002
- Status: completed (first step-3 prototype round; performance gates remain)

Implemented the three [scaling-study probes](../experiments/scaling-prototypes/report.md).
Opt-in [shared setup/workspace code](../encoding/liftrs-shared/scaling_experimental.go)
adds sparse linearized transform construction, Frobenius-paired base-field
root-only preparation, reciprocal G, defining P·G verification, and minimal online
P/spectrum state. The prefix/full-domain buffers alias a reusable arena, with H
saved before reuse and a separately allocated canonical output. Plan ownership,
limits, payload accounting and concurrency rules are documented, with a runnable
example. Defaults/native kernels/artifact validation remain unchanged; default
FFT setup only shares an extracted identical table-finalization function.

Test-only probes provide sectioned/checksummed incremental online loading, general-
basis matrix/Taylor forward/inverse with affine row cosets, ordinary disk-based
convolution, and bucketed canonical routing. The matrix column offsets require the
image basis; a development test caught a canonical-offset mismatch that was fixed
before the recorded successful runs. The prototype uses scalar ReadAt/WriteAt and
Go tiles, not a high-throughput native external backend. The routing test has a
full vector only as a reference sink; output bytes are logical accounting, while
bucket bytes are actual file calls. Fanout is bounded to 64 in these small probes.

[Runner, analyzer and reports](../experiments/scaling-prototypes/report.md) retain
source manifests, two measured runs, CPU profiles/text views, matching binaries
and reasonably sized preprocessing under ignored .cache/liftrs/scaling-step3/.
Original s=13 cache remains untouched. Exact pre-check-addition sources are in
content-addressed snapshots, preserving hashes rather than rewriting results.

Results: dense/sparse descriptions agree through s=13; only sparse setup is run
at s=14/15, with additional vanishing/Horner/inverse checks. At s=13 dense setup
allocates ~514.83 MiB versus sparse ~2.79 MiB including profiling overhead. s=14/15
sparse description sizes are 148/191 KiB; this is not full encoder support. Small
paired preparation preserves P/G/spectrum and reduces observed single-call times,
but s=9 cumulative allocations increase and sampled peak heap is not improved;
no large-setup RAM guarantee follows. Native online workspaces preserve words on
small fixtures, with smaller repeated-call allocation.

One native s=13 workspace encoding reuses retained setup and takes 153.374 s,
essentially the prior 153.409 s median; no performance speedup is claimed. All
11,926,435 systematic positions and the original full canonical SHA-256 match.
Reusable workspace payload is 316,141,196 bytes; online P/spectrum is 316,141,200.
Encoding allocates ~270.6 million bytes including profiling and its final 256 MiB
output. The ~907 MiB final Go heap is an endpoint, not RSS peak; Go Sys retains
~2.29 GiB from loading. No packed output or full baseline/preprocessing was rerun.

Validation: default/cgo-disabled and native tagged correctness tests, tagged vet,
Go race checks and default/tagged full suites pass in isolated copies. Tiny words
match the independent dense oracle; matrix tests cover every nontrivial split for
s=2..4 and partial/full domain convolution. Bad file checksum/limit and workspace
ownership/input cases are tested. Evidence analyzers, exact source/binary/cache/
profile checks, Python syntax, runnable example, documentation links and whitespace
checks pass; [final verification](../experiments/scaling-prototypes/verification.json)
links outcomes. Go race instrumentation does not instrument C++; kernels are
unchanged and prior native sanitizer evidence remains applicable. New Go profiles
cover calls; detailed native attribution reuses the retained native sample.

README checked/updated; study, algorithm notes, encoder contract, roadmap,
experiment index, backend guides and historical snapshot instructions are current.
Next gates are blocked-I/O implementation, measured preparation scratch/GC
lifetimes, fresh minimal-loader process memory, and native arbitrary-basis tile
execution if selected. Exact large capacity, wider representations, bounded
discovery, public artifact distribution and crypto/grouped output remain open.
No higher budget, hardware/backend selection, larger-size support or complete
TRL 5 acceptance is claimed. Changes are left uncommitted as no publish request
was made for this work item.
