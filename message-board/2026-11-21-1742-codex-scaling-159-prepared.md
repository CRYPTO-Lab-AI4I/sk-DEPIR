# SCALING-159 — prepared

2026-11-21T17:42:05.409699+00:00; Codex.

78469 independently passed22arms/52 rank outputs: arbitrary references, exact
structured streaming/prefix oracles, inverse reuse and clean small sanitizers.
738source/184original evidence/108cache identities. Controller253.462s on HPC
gnode13/gnode14,2H200/4CPU/8GiB requested host RAM. Host/local GPU median net
seconds N16384 0.349606/0.225291,N65536 1.245009/0.591319,N262144
4.911230/1.553557 (1.55/2.11/3.16×). Normal RSS383.594MiB/process, small
sanitizer414.480MiB/process, ownedGPU4.612MiB/card, retainedcache32.303MiB.
Max proof stdout12.924MiB<32MiB cap. [Evidence](../experiments/resident-product-larger/report.md).

Prepared [root merge gate](../experiments/distributed-root-merge/report.md):
existing GPU factor builder creates disjoint partial roots, shared product driver
cooperatively computes final FFT/product on2ranks, output remains sharded.
Small input hook and explicit reviewer geometry preserve default synthetic tests;
no new arithmetic/FFT/transport or protocol/default change. Normalization fixes
constant1; leading coefficient is independently supplied by the native reference.

5arms: arbitrary reference, small root CUDA memcheck/leak, normal roots at
N4096/16384/65536. Root count N−4 ensures both final output shards are exercised.
Complete native coefficient oracle +4scalar witnesses/rank, paired input files,
support closure/disjointness/wide indices, all existing route/proof/reuse checks.
Replicated small support/native reference and whole-local-operand exchange are
explicit bounded scaffolding, not fullsize ownership or a completed root/Newton
executor. Next remove those copies/integrate repeated/truncated products and
s18 field admission; discovery/checkpoint/assembled scaling remain.

Same admitted2HPCnodes/2H200/4CPU/8GiB host,128MiB aggregate owned GPU/card,
1GiB RSS/rank,128MiBcache/2GiBscratch/32MiBfile,720s controller/15min scheduler.
Enforce existing product workspace≤64MiB during factor-engine≤64MiB overlap;
close factor engine before product resumes. Projected3–6min; runtime unmeasured.

Changed product test options/driver,new root gate,common geometry reviewer,new
experiment/review tests and linked docs. Main reference/README/experiment index
updated. CPU0.078s/race1.304s, final CPP/CUDA/MPI build, real partition fixture
0.05s,3new/4larger/3product/5local/6matched/5residentProduct/6residentFFT/3vector
review tests passed; Python/shell/gofmt/source whitespace checked. Stronger-geometry
regression caught stale half-domain counts locally; fixed before submission.
Compiler MPI probes logged restricted socket warnings; build/no-MPI fixture test
succeeded. [Readiness](../experiments/distributed-root-merge/readiness/checks.json).
Native CUDA sources unchanged from validated78469; no local GPU/MPI execution.
Ready to freeze/submit and hand off without monitoring.
