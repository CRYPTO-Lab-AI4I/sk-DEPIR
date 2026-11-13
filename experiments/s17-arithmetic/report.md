# Wide packed orbit arithmetic for the s17 port

First implementation step in the
[s17 CPU-first/multi-GPU plan](../../docs/encoding-scaling.md#current-execution-plan-for-s17).
The common component fields already support s≤20. Existing encoder adapters,
coordinates, native buffers and artifacts still pack extension values into32 bits.
The new private scalar adapter uses uint64 for packed a+b*u, keeping its two base
components uint32. It adds no domain array, public constructor or encoder admission.

Status: implementation/static checks complete; runtime validation submitted as
**job74290**, source checkpoint5b9e706. [Submission record](submission.json).
No post-submit monitoring; runtime validation remains pending.
[Controller](run.py), [Slurm script](run.sbatch),
[implementation](../../encoding/liftrs-shared/orbit_arithmetic64.go),
[tests](../../encoding/liftrs-shared/orbit_arithmetic64_test.go).
Results, exact revision, source digests, commands, host/Go details and CPU profile
will be retained in `runs/scalar-01/`; binaries/source/build cache remain in
`.cache/liftrs/s17-arithmetic/scalar-01/` and `/tmp/sk-depir-wide-<job>/`.
No preprocessing is created or reread; existing code caches/profiles remain intact.

```sh
python3 -B experiments/s17-arithmetic/run.py --check
module load slurm/slurm/25.05
sbatch --output=experiments/s17-arithmetic/slurm-%j.log \
  experiments/s17-arithmetic/run.sbatch scalar-01
```

One CPU-only job,4cores,8GiB RAM request/6GiB sampled process-group RSS guard,
4GiB Go soft memory,2GiB new retained cache/4GiB scratch guards. No CUDA, native
build or full grid; Go1.27.3, CGO disabled. Worker10minutes, controller15minutes,
Slurm20minutes. Limits bound compilation and validation, not a large encoding.
Controller/resource guards remain part of execution; no post-submit monitoring.

## Width audit and invariants

| Boundary | Existing representation | Required port work |
|---|---|---|
| Base-field data | uint32 coefficients/components; common fields support s≤20 | Keep32-bit slots; no reason to double these arrays |
| Scalar extension values | `orbitField` packs both components into uint32, s≤16 guard | New private uint64 scalar reference; native equivalents next |
| Orbit/grid coordinates | uint32 omega/points in OrbitBasis and assembly | Widen extension coordinates and validate range before narrowing output values |
| Transform/preprocessing buffers | Go `[]uint32`, C/CUDA uint32 pointer interfaces | Separate native64-bit plans/buffers/metadata; retain32-bit path |
| Counts/cache formats | s-dependent count widths, but coefficient payloads remain32-bit | Explicit value-width/schema compatibility and checked byte offsets; existing caches unchanged |
| GPU distribution | One-device execution | New sharding/communication prototype; a64-bit scalar is not a multi-GPU encoder |

The tower is unchanged: u²=u+delta with the deterministic first trace-one delta,
packed as a|(b<<s). Generator search preserves the first primitive packed element
at least q; use uint64 group orders/exponents and overflow-safe factor bounds.
Canonical internal inputs must be below2^(2s); the future loaders/native boundaries
must enforce that invariant before bulk execution. Inversion rejects zero.

Tests cover s2/8/13/16 and17–20, zero/one/max values and the highest extension bit,
256 deterministic random samples per size (seed170049), independent carryless
polynomial multiplication/long division, inverse and Frobenius identities,
primitive-order witnesses via the component API, and legacy generator/delta and
multiplication agreement through s16. Existing table-free field tests also run.
All tests operate on scalar samples; passing cannot establish full transform,
preprocessing or encoding correctness/performance at s17. Runtime/capacity tables
remain projections. CPU profiles describe correctness validation, not encoding
latency benchmarks. The new scalar path is a reference for subsequent native
work, not a selected high-performance kernel or a change to existing defaults.
