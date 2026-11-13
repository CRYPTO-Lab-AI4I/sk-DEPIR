# CPU-044: complete disk-backed CPU route validated

Timestamp: 2026-11-12T22:32:51.014812+00:00
Agent: Codex
Status: completed

Job **74141** passed; completed0:0 in8m33s on gnode17/NFS. Reconciled401 source
identities and all sealed cache/evidence hashes with
`timeout 25s python3 -B experiments/cpu-disk-encoding/analyze.py online-01`.
All9 small Go fixtures/constant cases and the ASan repeat pass, including wide
header,error/workspace recovery,ownership/Close/existing-file refusal. Ordinary
Go/CPP,combined API build/vet pass. Two complete encodes per s10 mode agree on
all146641 systematic positions,16 sampled curves per encode, canonical/packed
hashes and1027 readbacks. [Report](../experiments/cpu-disk-encoding/report.md),
[analysis](../experiments/cpu-disk-encoding/runs/online-01/analysis.json),
[accounting](../experiments/cpu-disk-encoding/runs/online-01/accounting.log).

File-fast first/reused0.3315/0.3106s; disk2.8604/1.1356s, fixed-order profiled
samples. Disk owning output4MiB plus conservative native buffers1.125MiB and
8MiB backing files replace file-fast8MiB arrays, excluding message,metadata,
runtime and filesystem cache. Logical I/O490210116 read/163054404 write bytes
per encoding,98637/40040 syscalls. Write timing1.864s first/0.0765s reuse shows
filesystem variability; no physical storage extrapolation. Constructors were
timed in load-online (0.0156s disk/0.0103s fast), not tiny harness setup closures.
Group peak1056096256bytes and cgrouppeak1520631808bytes include build/validation/
cache; final mincore43files7446528bytes is not a peak or total RAM bound. Slurm
MaxRSS163672KiB is a separate per-task sampling view. All artifacts/profiles and
preprocessing retained; large files stay outside Git with paths/hashes in report.

Updated report, [consolidated plan](../docs/encoding-scaling.md#current-memory-route-status),
root README and C++/shared API READMEs with measured costs and validation sizes.
README checked, obsolete pending/remaining-sequence statements reconciled and
CUDA directory included in implementation map. Local documentation links checked,
whitespace checks pass. No code changed this review; no additional tests or job
needed after sealed result reconciliation. No monitoring occurred before ping.

All four approaches now have CPU/GPU implementations. Compact/streamed-code
routes have historical fulls16 evidence; GPU packed sink fulls16, CPU packed
sink small/ASan/matcheds10; owning host/disk-backed routes small/ASan/fulls10,
components through2^22. Backed packed sinks and other combinations are not all
implemented. Faster defaults stay; slower modes explicitly selected. Fulls16
CPU disk storage32GiB exceeds16GiB new-cache admission and remains unapproved;
no s17 execution support/feasibility or TRL5 completion claim. Next is ordered
item5: compare combinations and full lifetimes/budgets, then larger validation
and34-bit arithmetic/coordinate/serialization/preprocessing work. No new job or
budget increase follows from this completed review.
