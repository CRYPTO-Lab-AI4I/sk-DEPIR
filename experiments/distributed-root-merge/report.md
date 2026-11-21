# Local GPU roots with a distributed final merge

Status: prepared; no new GPU result yet. This follows [passed larger products
78469](../resident-product-larger/report.md). Producer revision and exact source
identities will be retained by the controller; [readiness checks](readiness/checks.json)
record local validation. Work SCALING-159.

[Runner](run.py), [Slurm entry](run.sbatch), [independent reviewer](analyze.py),
[review regressions](test_review.py), [opt-in root driver](../../encoding/liftrs-shared/distributed_root_product_gpu_test.go),
[shared product driver](../../encoding/liftrs-shared/distributed_product_gpu_test.go).

## Implementation and correctness boundaries

Each rank uses the existing deterministic root fixture and Frobenius-preserving
partitioner. Its GPU constructs the normalized polynomial for its own root subset
through the existing factor builder. Ranks exchange local polynomials, take their
owned coefficient intervals, and jointly compute the final product through the
same resident FFT/Taylor/pointwise schedule and proof-checked MPI adapters that
passed78469. Final GPU output stays sharded; no executor computes the final
product alone. Factor-engine memory is released before distributed FFT work.

A small input hook in the shared test driver supplies real operand shards, counts,
leading coefficient and an independent expected shard. It does not add another
FFT, product recurrence or transport implementation. The common reviewer accepts
explicit polynomial geometry; default synthetic tests retain their former counts
and leading-one checks. Root polynomials have constant coefficient1, but their
leading coefficients need not be1.

This is a **bounded integration gate**, not a production distributed root executor.
It deliberately replicates the small support and complete native CPU reference
on each rank and exchanges each complete local operand. Removing those copies
and integrating the rank-owned reduction tree are subsequent work; this layout
must not become an implicit fullsize memory or runtime claim. Synthetic closed
supports are not certified PIR code supports. Newton/generator integration and
s18 field admission have not been implemented by this change.

Every output coefficient is compared against a complete existing native CPU root
construction. Four independent scalar extension-field root witnesses supplement
that reference on each rank. Saved reference shards reconstruct its complete
hash and degree; the reviewer checks constant, leading and padded coefficients,
base-field range, disjoint sorted Frobenius-closed support shards (including wide
indices), both paired local/peer files and input-preservation claims. Existing
complete transpose/Taylor route and matched payload proofs, full-extension
pointwise checks, right-spectrum preservation and inverse-allocation reuse remain.

Review regressions reject bad peer hashes, missing remote-source proofs and an
incorrect result whose producer expectation and checksum have also been changed.
These parser fixtures exercise evidence validation; actual root mathematics is
checked against the independent native/scalar implementation during GPU execution.

## Workload, timing and resources

Five MPI arms, with one source freeze/build/preflight/allocation:

| Arm | Evaluation values | Roots | Validation |
|---|---|---|---|
| Arbitrary-input regression | 16,384 | Synthetic coefficient inputs | Existing schoolbook oracle, both product shapes |
| Root CUDA memcheck/leak test | 4,096 | 4,092 | Complete native/scalar root oracle |
| Normal root merge | 4,096 | 4,092 | Complete native/scalar root oracle |
| Normal root merge | 16,384 | 16,380 | Complete native/scalar root oracle |
| Normal root merge | 65,536 | 65,532 | Complete native/scalar root oracle |

The s17-field root subsets contain two-element Frobenius orbits, balanced equally
across two ranks. Each operand has N/2−1 coefficients; the result has N−3,
leaving three explicit zero coefficients. Both output shards contain
coefficients of the final polynomial. Larger root cases are not sanitized. No full PIR
support discovery or preprocessing cache is created. The shared fixture generation
is deterministic (seed109+s+count), preserving provenance.

Final-product timings preserve78469 boundaries: reset/upload, distributed product,
readback and proof logging are inside; diagnostic right-spectrum capture is timed
separately. Final coefficient comparison, fsync/publication and input/reference
preparation are outside that product clock, inside complete arm/controller time.
`root-input` records reference/witness time, local GPU root time (including close),
operand exchange and complete input preparation separately. The complete-input
clock includes local setup, partitioning, saved evidence and other host work;
these subtimes need not sum. Constructor setup of the shared resident product is
separate. Go profiling remains enabled in normal arms; native/GPU work appears
as opaque cgo calls. These are correctness/component measurements, not unprofiled
end-to-end preprocessing latency or a five-arm strong-scaling comparison.

Use the existing admitted **two HPC nodes / two H200s / four CPU cores / 8 GiB
requested host RAM**, one GPU/two cores/4 GiB per node. Same **128 MiB aggregate
owned GPU/card**, **1 GiB RSS/rank**, no pinned buffers, **128 MiB retained cache /
2 GiB scratch / 32 MiB per file**, **720-second controller / 15-minute scheduler**.
Exclude gnode03; otherwise allow eligible nodes. No resource increase is requested.
The product workspace must be at most64 MiB before the input hook runs; its
concurrent factor engine has an enforced64 MiB cap. The factor engine closes
before the product resumes under its existing128 MiB cap. Context/driver memory
is additional. This bounds overlapping owned allocations rather than summing
sequential workspaces or silently ignoring concurrent allocation.

At largest N, the resident three-array payload is0.75 MiB/card plus bounded
windows/tables. Local-operand exchange sends128 KiB/rank and receives128 KiB/rank.
A replicated complete CPU reference is256 KiB/rank, plus support/partition/maps
and native workspace. Saved output/reference/pair/support files are small;
projected retained cache is below40 MiB and proof logs below16 MiB/file, based
on78469's measured12.924 MiB at the larger262144-value domain. No disk matrix is
used, useful prior caches remain and no memory-saving default is selected.
These are estimates until this gate is measured.

Prior78469 controller time253.462 s covered22 product arms, repeated samples,
profiling, sanitizer and compilation. Earlier root worker tests validated complete
native/GPU roots through262144 exponents. The new five-arm gate uses at most65532
roots. Budget **3–6 minutes estimated** for controller/builds/checks/profiles, with
headroom inside720 seconds; each arm retains the validated240-second launcher /
220-second test cutoff. This is a budget estimate, not a measured new root time.

Reuse shared-node path/executable/wrapper/library/cwd/TMPDIR preflight and UCX
registration-cache controls. Timestamped root-stage progress, completed work,
transform passes, communication proofs and memory samples flush to retained logs.
Completed arms/cache hashes/profiles save before later work. No agent monitoring
occurs after submission; wait for the user's completion ping before analysis.

```sh
python3 -B experiments/distributed-root-merge/test_review.py
python3 -B experiments/distributed-root-merge/run.py --check
module load slurm/slurm/25.05
sbatch --parsable --output=experiments/distributed-root-merge/slurm-%j.log \
  experiments/distributed-root-merge/run.sbatch roots-01
# After the user's completion ping:
python3 -B experiments/distributed-root-merge/analyze.py roots-01
```

Raw results/commands/profiles/logs: `runs/roots-01/`. Reusable sealed inputs,
executables and coefficient evidence: `.cache/liftrs/distributed-root-merge/roots-01/`
(outside Git; manifest hashes retained with results). The source snapshot records
actual producing revision and source files. No fullsize result is available.
After success, connect partitioned ownership/repeated or truncated products to the
same executor for Newton/generator identities, with explicit s18 field admission.
Distributed discovery, coordinated checkpoints/publication and assembled pipeline
scaling remain before fulls18 preprocessing.
