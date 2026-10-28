# Static encoding resource study

[model.py](model.py) produces [results.json](results.json) and [tables.md](tables.md)
from the catalogue and source audit. This is **analytical accounting**, not a
benchmark. The [scaling study](../../docs/encoding-scaling.md) contains the inventory,
three designs and five-step plan. No encoder algorithms or budgets change here.

## Inputs and reproduction

Base revision: f2aedf8 plus this study's model/probe/docs; result manifests record
exact source and input SHA-256 values. Catalogue s=13 uses verified K; s=14..20 use
Monte Carlo point capacities and retain prefix-size ranges from the recorded
confidence intervals. Baseline is the shared encoder with C++ transforms and Go
preprocessing, not a fresh Go comparison run.

```sh
GOCACHE=/tmp/sk-depir-go-cache /tmp/sk-depir-go-toolchain/go/bin/go run \
  ./experiments/encoding-resources > experiments/encoding-resources/layout.json
PYTHONPYCACHEPREFIX=/tmp/depir-pycache \
  python3 experiments/encoding-resources/model.py
```

[layout.go](layout.go) uses Go reflection to check existing type sizes without
constructing encoder arrays. [layout.json](layout.json) records eight-byte host
integers and a 72-byte logical product-tree node. It is not an RSS measurement.
Environment: Python 3.9 standard library, Go 1.27.3, macOS 26.6.2 arm64 on the same
48 GiB Apple M5 Pro host. No native workload, profiling run, GPU or storage benchmark
was needed: the model performs only integer arithmetic and reads existing results.
Previous CPU profiles and expensive caches remain retained under the existing
[profiling policy](../README.md#profiling-and-preprocessing-retention).

## Model and checks

- P has K+1 coefficients; G has N−K; the optional cached spectrum has N. Current
  support is K host integers. Principal buffer scenarios count M, H, prefix FFT
  arrays, the full y buffer and a separate canonical word.
- Current element slots are 4 bytes; hypothetical s>=17 rows use 8-byte packed
  extension elements. This is a width scenario, not implemented support.
- Dense transform setup reaches consecutive polynomial arrays of N/2+1 and N+1.
- Logical product-tree polynomial lengths use the exact recurrence
  `C(1)=2; C(K)=K+1+C(floor(K/2))+C(ceil(K/2))`; 2K−1 node headers are included.
  Built-in small and power-of-two checks verify this recurrence independently.
- The candidate online plan retains P+spectrum only; no G/support values are
  needed by cached online encoding once K/specification are independently verified.
  Principal workspace formulas are documented in design A. These are buffer
  schedules to implement and measure, not existing resource guarantees.
- Fixed-width packed storage separately counts s bits per base coefficient,
  2s bits per spectrum value and 2s bits per support index. Aligned counts are
  also recorded. Metadata/checksums/section alignment and any compression are
  excluded. Packed widths do not change canonical field representation.
- Current uncached convolution splits G at N/2, so it has three N and three N/2
  transforms plus the prefix work. A proposed dedicated uncached convolution
  could use three N transforms; it is not implemented. Both work-volume counts
  are recorded separately from the cached baseline.
- Transform traffic uses the explicit scenario parameter p: one sweep reads and
  writes its vector once. Record p=2,4,8 as sensitivity cases; no pass count is
  proven. Point products/routing are separate fields and must be added rather
  than silently treated as included. Routing assumes one bucket-partition level
  with four-byte local offsets; extra levels and bounded fanout add traffic. An earlier partition with a bucket
  span above 2^32 needs wider offsets; the model records another 24N bytes for a
  read/write pass over 12-byte intermediate pairs.

No payload count includes Go spare capacity, allocation classes, dead-but-not-yet-
collected arrays, JSON/raw-message copies, support-validation map overhead, runtime,
OS page cache or native metadata. Consequently **none is a peak-RSS bound**.
Retained tree slice capacities can exceed logical lengths. The source audit lists
division scratch without pretending all allocations have identical lifetimes.
No new runtime forecast or larger-case admission is produced.

## Findings

The generated tables show that earlier proportional RAM forecasts mixed necessary
state with avoidable setup/loading overhead. At s=15, the principal current-array
scenario is 19.23 GiB, and a lean schedule is 14.42 GiB; setup and real allocation
behavior remain unresolved. In particular, retained tree coefficients and node
headers alone are about 54 GiB. Sparse setup and root-only cyclic preparation are
therefore concrete prototype priorities, not proof that first use fits locally.

At s=20 the proposed in-memory schedule still needs about 27 TiB of principal
payload and a packed P+spectrum cache about 5.65 TiB. A full packed artifact is
about 8.8 TiB. The study does not establish moderately sized public preprocessing
or local feasibility at that scale. Tiled/external-memory designs need their exact
Taylor/inverse/output schedules demonstrated before assuming low pass counts.

There are two discontinuities absent from a smooth extrapolation: s>=17 needs
wider packed extension values, and s=20's point capacity doubles the prefix FFT
length relative to N/2. The s=19 sampling interval crosses the same padding
threshold. Exact capacity is necessary before actual encoding, as before.

Measured anchors remain unchanged: s=13 native-backed encoding 153.41 s median,
loaded process peak RSS about 2.26 GiB, reusable Go preparation 1,765.37 s and
whole first-use peak RSS about 5.34 GiB. They come from the original
[native comparison](../native-encoding/report.md) and
[first-use run](../cyclic-s13/report.md), which use distinct measurement boundaries.
The model must not be presented as another measured benchmark or as a correction
to those observations.


## Verification

[verification.json](verification.json) records successful Go layout-probe compile,
vet and test commands, Python syntax checks, model recurrence/padding assertions,
deterministic result regeneration, source/input checksums and current local links.
The probe package has no workload tests; the model's assertions check the analytical
formulas. No production encoder code changed, so the existing full encoder suites
and expensive performance baselines were not repeated.

Primary algorithm references reviewed for candidate design B are linked in the
[study](../../docs/encoding-scaling.md#b-tiled-external-memory-cpu-path). Their
hierarchical/matrix structure motivates prototypes; neither published additive
benchmarks nor classical FFT pass counts are imported as performance evidence.
