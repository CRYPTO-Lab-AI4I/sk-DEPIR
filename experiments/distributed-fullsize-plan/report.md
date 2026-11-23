# Full-size distributed preprocessing admission

SCALING-196. Full s18 preprocessing is **not run / not admitted**. The assembled
command and all completed-stage recovery modes passed representative s10 job82199
([results](../distributed-discovery-restart/runs/recovery-01/analysis.json)). This
step determines actual root-list imbalance and full-job headroom before requesting
full-size resources. It changes no library, protocol or discovery algorithm.

## Current capacity audit (projections, not allocations)

For m2,t5,s18: N=68,719,476,736; exact K=16,300,262,095 from the
[dimension measurement](../s18-preprocessing-admission/report.md). Sixteen workers,
one per H200, are a provisional topology, not a fixed design limit.

| Item | Source-derived size / outstanding issue |
|---|---|
| GPU principal arrays | Three full-domain shard arrays: **96 GiB/card on16 cards**; tables, Taylor/transfer windows, context and allocator headroom are additional. Current server field/Taylor limits support s18; full allocations unmeasured. |
| Host generator stage | A full-domain uint32 shard is **16 GiB/worker**. Prefix-input Newton has **64 GiB/worker in four base arrays**; dense inputs have80 GiB. Add scratch, field tables, roots still live, native/runtime/MPI/context and unreclaimed predecessors. The old48 GiB estimate covered inputs alone. |
| Discovery/local roots | Ownership is not uniform. Root-list capacity and rounded local/subgroup polynomial domains need actual counts; dividing K by16 is insufficient. Local CUDA executors allocate lazily as products grow; their global-N limit is not an up-front three-N allocation. |
| Persistence | Canonical code about **572.723 GiB**, generator **195.277 GiB**, raw support **121.446 GiB**, all compact prefix checkpoints about **512 GiB**. About **1.37 TiB physical /1.49 TiB logical** including the support hardlink alias, before temporary files, logs/profiles and safety margin. |
| Restart granularity | Completed discovery, root and Newton prefixes recover. Partial sampling and unfinished local/subgroup root trees do not. A long root phase may still be lost. |
| Runtime and host reserve | Full-size end-to-end latency and runtime/native/GC allowance are unmeasured. The small-case768 MiB declared reserve is not a full-size bound. Choose a roomy explicit reserve and deadline after the inventory; do not infer them from GPU payload alone. |

No slower memory strategy is selected here. Faster resident execution remains the
preferred path when admitted memory permits it. Any later savings involving extra
transforms, transfers or I/O require an explicit cost comparison and option.

## Read-only retained-support measurement

[Scanner](scan_support.cpp), [runner](run.py), [Slurm entry](run.sbatch),
[focused checks](test_scan.py), [check result](checks.json). Producing revision is
recorded in each run's metadata; source SHA256 is recorded before compilation.
Input is the existing independently verified full s17 support archive:

`.cache/liftrs/gpu-s17-full-preprocessing/s17-m2-t5/support/support.bin`

3,876,566,203 sorted uint64 little-endian indices; **31,012,529,624 bytes
(28.883 GiB)**. SHA256
`1d5251ecf13148683a28bff04b7e3f1c5f71cd8e3094eb083e147cb28c2e2419`.
No input copy or rewrite. One sequential pass checks order, range, exact count/EOF
and full SHA256 while counting production closed-orbit ownership for2/4/8/16/32/64
workers. Folding64 buckets is exact for those power-of-two worker counts.
Histograms describe this retained support, not a proof of s18 distribution.

**Resources:** one eligible HPC node,2 CPU cores,4 GiB requested host RAM,zero GPUs;
8 MiB scan buffer,1 GiB child address-space cap,32 MiB per-file cap;720-second
controller deadline including build,15-minute Slurm cutoff. This is below the
existing approved2-node/8-core/16-GiB/2-GPU validation ceilings, with no new data
cache. Scratch contains only frozen source/executable; retained evidence is tiny.
The archive already exists and remains intact. About29 GiB is read, not allocated.

Measured tiny native-s10 scanner execution is milliseconds; full31-GB storage
latency is unknown. Budget **1–10 minutes excluding queue**, with a15-minute
scheduler cutoff and12-minute controller. This is a bandwidth/count inventory,
not a preprocessing performance baseline. CPU sampling is omitted for this
single-pass diagnostic: `/usr/bin/time -v` saves CPU/elapsed/peak RSS, and timestamped
progress saves completed/total indices, elapsed time and peak RSS every30seconds.
No GPU or preprocessing profile is replaced or removed. Child limits and timeout
remain active; no agent monitoring after submission.

```sh
module load gcc/14.2.0
g++ -std=c++17 -O3 -Wall -Wextra -Werror \
 experiments/distributed-fullsize-plan/scan_support.cpp -lcrypto -o /tmp/depir-scan-support
python3 -B experiments/distributed-fullsize-plan/test_scan.py /tmp/depir-scan-support
python3 -B experiments/distributed-fullsize-plan/run.py owners-01 --check
module load slurm/slurm/25.05
sbatch --parsable --output=experiments/distributed-fullsize-plan/slurm-%j.log \
 experiments/distributed-fullsize-plan/run.sbatch owners-01
```

Preflight validates retained size/manifest and tools before any payload read.
Eight boundary cases check valid ownership, hash mismatch, partial word, repeated/
descending/out-of-range indices and wrong counts. A retained native s10 comparison
checks all16 owner counts independently. Warning-clean compilation and shell syntax
checks passed. The full scan will publish `runs/owners-01/owners.json`, resource/
progress/build logs and revision/source/input metadata. Submission is not a pass.

## Bounded support-set cross-check

[Source](check_support.cpp), [s8 result](support-s8.json), [s10 result](support-s10.json).
The retained discovered support happens to equal the direct Cartesian good-monomial
predicate at s8 and s10 (zero differences). For s10 on16 workers, largest owner
has24,940 roots, **2.72 times the uniform average**; s8 is about3.01 times.
This motivates checking ownership instead of assuming balance. These are exact
small-case counts, not s18 estimates or a replacement for probabilistic discovery.

The owner formula has a simple exact form when R divides q: write v=a+q*b,
0<=v<q²−1. Multiplication by q modulo q²−1 swaps a,b. Thus the existing
`min(v, q*v mod(q²−1)) mod R` equals `max(a,b) mod R`. This arithmetic identity
alone does **not** establish that the Cartesian set equals discovered support at
other sizes. Paper section4.2.4/Claim4.2 uses sampled codewords to identify the
univariate degree set. No shortcut is adopted without a general justification.

Reproduce the bounded check (under one second on the HPC login host):

```sh
module load gcc/14.2.0
g++ -std=c++17 -O3 experiments/distributed-fullsize-plan/check_support.cpp -o /tmp/check-support
/tmp/check-support 8 5 8 .cache/liftrs/distributed-host-admission/host-01/cold/s8-support.bin
/tmp/check-support 10 5 4 .cache/liftrs/gpu-products/products-02/cpu-0/support.bin
```

After the scan, use actual skew and root rounding to tighten the fullsize inventory;
choose node/card packing, generous host reserve, checkpoint/deadline/storage limits
and a representative large allocation measurement if still necessary. Present the
complete fulls18 resource plan for agreement before its submission.

## Submission82200

Submitted from **7bbf193**; [receipt](submission.json). Results
`runs/owners-01/`, log `slurm-82200.log`. Same bounded resources above.
Estimated **1–10 minutes excluding queue**; no agent monitoring. Fulls18 remains
unrun/unadmitted. Submission is not a successful scan.

## Verified outcome82200

[Analyzer](analyze.py), [analysis](runs/owners-01/analysis.json),
[raw histogram](runs/owners-01/owners.json), [resource record](runs/owners-01/resources.log),
[progress](runs/owners-01/progress.log), [Slurm log](slurm-82200.log).
Receipt/revision7bbf193/sourceSHA, input SHA/count/bytes, folded histograms and
monotonically increasing progress all reconcile. The analyzer does not reread
29 GiB; the producing job checked the complete input SHA during its one pass.

HPCgnode13,1node/2cores/4GiB requested RAM/noGPU. Scan **2 min23.040 s**;
controller including build **2 min23.937 s**; peak process RSS **14.785 MiB**.
All3,876,566,203 indices checked;28.883 GiB input remains unchanged.

| Workers | Largest actual root list | Smallest list | Largest /uniform average | Largest list storage (raw indices) |
|---|---|---|---|---|
| 2 | 2,247,560,698 | 1,629,005,505 | 1.160 times | 16.746 GiB |
| 4 | 1,355,297,730 | 736,780,003 | 1.398 times | 10.098 GiB |
| 8 | 823,825,144 | 338,286,312 | 1.700 times | 6.138 GiB |
| 16 | 482,718,852 | 154,654,880 | 1.992 times | 3.597 GiB |
| 32 | 273,880,228 | 70,187,411 | 2.261 times | 2.041 GiB |
| 64 | 151,323,728 | 31,531,660 | 2.498 times | 1.127 GiB |

List storage excludes append-capacity headroom and polynomial arrays. Scaling the
16-worker maximum only by K18/K17 projects about2.030 billion s18 roots; this is
not an s18 proof. The [prepared full trial](../distributed-s18-full-preprocessing/report.md)
allows4,294,967,295 roots per worker,16H200s/eight nodes,192GiB/worker additional
host allowance and400GiB/worker RSS guard. Host headroom drives two workers/node.
No fulljob is admitted/submitted; explicit larger-resource agreement is next.

Reconcile saved evidence: `python3 -B experiments/distributed-fullsize-plan/analyze.py`.
