# Width-aware reusable scaling code data

Implements the format/loader part of step3 in the
[s17 execution plan](../../docs/encoding-scaling.md#current-execution-plan-for-s17).
[Representation/API specification](../../docs/scaling-code-data.md),
[implementation](../../encoding/liftrs-shared/scaling_code.go),
[tests](../../encoding/liftrs-shared/scaling_code_test.go),
[example](../../encoding/liftrs-shared/scaling_code_example_test.go),
[controller](run.py), [Slurm script](run.sbatch).
Runtime validation submitted as **job74396**, source checkpointf24682c;
[submission record](submission.json). No post-submit monitoring. No full s17 code data, discovery/preparation or
encoding is produced; only bounded compatibility/representation checks.

New APIs read existing LRSprobe1/2 and LRSprobe3 at s17–20. P values remain uint32,
spectrum slots stay uint32 through s16 and become uint64 thereafter. Explicit new
polynomial/delta/generator/slot metadata prevents reinterpretation. A trusted
SHA-256 plus count and payload-byte caps are mandatory. Resident arrays use one copy
at their selected width; streaming verification uses no domain arrays. Existing
CPU/GPU paths and caches are not changed. There is no selected slower memory mode;
standalone verification followed by loading would require two full I/O passes,
and remains an explicit caller choice, not the loader default.

## Reproduce

```sh
python3 -B experiments/s17-code-data/run.py --check
module load slurm/slurm/25.05
sbatch --output=experiments/s17-code-data/slurm-%j.log \
  experiments/s17-code-data/run.sbatch data-01
```

Use a fresh run name. Commands, exact revision/source hashes, environment,
results/logs/CPU profiles go to `runs/data-01/`; matched Go/native test executables
remain in `.cache/liftrs/s17-code-data/data-01/` outside Git. No existing large
cache is read; small preparation fixtures are regenerated deterministically from
seed170057. All reusable existing caches/profiles are preserved.

## Checks and limits

Prepare a real s4,m2,t2 code using the existing reference, serialize with the old
writer, load/verify/write through the new API and compare every coefficient,
spectrum value, systematic position and canonical byte. Confirm no legacy
spectrum widening; an accepted legacy wide-count/small-s header writes the
canonical smaller count width. The example uses the same bounded fixture.

Inspect complete headers throughs20, including2^32+ counts, offsets and native
coordinate/exponent witnesses above32bits, without their payloads. Reject big
headers under small allocation caps; streaming a header-only big file must hit
EOF rather than allocate a domain. Check all truncated new-header lengths,
malformed count/field/tower/generator/width/reserved metadata, overflow shapes,
mandatory/mismatched digests, value ranges, P constant, trailing/truncated payload,
short reads and writer failures. Exercise8-byte values above 2^32 across64KiB
block boundaries and reject narrowing overflow. The old encoder loader must
continue rejecting the new format.

No full-wide payload can be round-tripped at bounded size because its declared
N is2^34 or larger. Actual8-byte element codecs are tested independently using
16391 values; full new headers are checked separately. Full byte identity/range
and polynomial proofs on a genuine s17 cache remain later admission gates.

Run targeted Go checks and full shared Go-only and liftrs_cpp regressions, using
package working directories for the retained test binaries. CPU profiles and
matching binaries are retained. Memory/latency clocks describe validation only,
not encoding. One node,4CPU,8GiB RAM request/6GiB RSS guard,4GiB Go soft memory,
2GiB new retained cache/4GiB scratch. Worker10m/controller15m/Slurm20m, no GPU,
Go1.27.3/GCC14.2. No budget increase or post-submit monitoring.
