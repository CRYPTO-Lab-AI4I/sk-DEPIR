# Scaling study steps 1–2 completed

- Timestamp: 2026-10-28 22:59 UTC
- Work item: SCALE-001
- Status: completed (study steps 1–2)

Recorded the agreed [five-step study](../docs/encoding-scaling.md): resource audit,
execution designs, focused prototypes, local calibration, then design selection.
The baseline and starting code are shared encoding with C++ transforms; Go
preprocessing and independent references remain. Steps 3–5 remain planned; no
production encoder, API limit, resource guard or backend default changed.

Step 1 audits discovery, dense transform setup, retained product trees, generic
division, JSON/loading/map validation, prefix/full FFT buffers and canonical
output scatter. [Model/report/results](../experiments/encoding-resources/report.md)
provide exact logical counts conditional on catalogue K, with input/source hashes,
width assumptions and reproducible scripts. K is verified only at s=13 in this
study; future rows use sampling estimates. Principal payloads are not RSS bounds.
At s=15 current principal arrays account for ~19.23 GiB; the proposed lean path is
~14.42 GiB before allocation/GC/OS overhead. Current retained tree payload alone
is ~54 GiB. These findings refine earlier proportional RAM forecasts without
claiming actual s=15 feasibility. s>=16 size accounting and s>=17 element widths
need changes; s=19 capacity uncertainty can cross a prefix-padding threshold,
and the s=20 point estimate doubles L relative to N/2.

Step 2 compares lean in-memory CPU, tiled external-memory CPU, and optional GPU
execution of the same tile semantics. Sparse constants, a root-only cyclic product
builder, Frobenius-paired base-field factors and direct reciprocal construction of
G are candidate preprocessing improvements, not implemented changes. The study
accounts for keeping/removing/regenerating spectra and explicitly notes current
uncached overlap-add costs more transforms than an ideal single-domain product.
Arbitrary-basis Taylor/inverse transforms, bounded discovery/preparation and grid
output routing are required prototypes; naive mmap or independent arbitrary tiles
are insufficient. Output routing can be combined with in-memory FFTs before moving
all transforms to disk. Large spectra/public artifact distribution remain open.

Read paper.pdf pages 31–32 with pypdf following the PDF skill; reviewed primary
Samanta–Badakhshan–Gong general-basis AFFT and Bailey hierarchical-memory FFT
references linked in the study. No specialization speedup or external-memory pass
count was assumed. I/O figures use explicit scenario parameters, not measurements.

Verification: Go type probe confirms 8-byte int and 72-byte logical tree node;
probe compile/test/vet and Python syntax pass. Model recurrence and threshold
assertions, deterministic regeneration, all recorded source/input hashes, changed
document file links and `git diff --check` pass; see
[verification](../experiments/encoding-resources/verification.json). CPU profiling
is inapplicable to this static accounting/probe; existing cache/profile data remain
untouched. No large benchmarks or unchanged full-suite reruns were needed.

README checked/updated, with links and current study status. Updated roadmap,
algorithm review, encoder contract and experiment index. Next: step-3 focused
prototypes, beginning with reusable setup/workspace fixes and exact tile semantics;
measure actual memory/I/O before selecting hardware or admitting larger runs.
