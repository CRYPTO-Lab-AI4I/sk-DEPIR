# MEMORY-035 — Audit complete; material route choice pending

Timestamp: 2026-11-12T08:28:47.253397+00:00
Agent: Codex
Status: handoff

Published the prior validated work as f9581d3 (push confirmed). Completed the
recorded representation/constant-residency audit in
[experiment report](../experiments/streamed-constants/report.md), with
[allocation-free script](../experiments/streamed-constants/run.py) and
[results](../experiments/streamed-constants/results.json). Updated scaling plan,
README and experiment index; README check completed. No production kernels,
parameters, defaults, artifact formats or admitted budgets changed.

Source review shows P is needed only in prefix initialization and spectrum only
between full forward/inverse. Proposed fast GPU streaming can upload P directly
and borrow dead transpose scratch for256 MiB spectrum chunks. Projected s16 owned
32.005 GiB versus measured compact35.429 GiB, preserving fast rows. Adds19.422 GiB
code uploads/message, retains verified host plan (~39 GiB host RAM), no new disk
reads or GPU tile allocation. Runtime is unknown; bandwidth scenarios are not
measurements. Alternative compact/direct streaming16.25 GiB arrays sacrifices
row locality; CPU file-backed streaming20.172 GiB principal RAM adds disk reads,
initial integrity verification and page-cache accounting. Existing fast modes stay
default. A route-choice question is pending under AGENTS.md memory–time policy.

s17 K remains estimated. Counts need35 bits, coordinates/support34, extension34;
base/output17 bits still fit uint32. Existing probe2 count widths suffice but
spectrum payload/adapter/native ABI/CUDA plans/support files do not. Base polynomial
already exists; that does not extend encoders. A uint64 one-arena design can
potentially use disjoint uint32 halves (128.25 GiB GPU arrays+tile), requiring new
kernels/layout proofs and wider types. Fast two arrays256 GiB exceed available
H200140.40 GiB. Retained host plan/message/output220.841 GiB; file-backed principal
host78.670 GiB plus142.420 GiB disk backing. CPU bounded file-backed144.670 GiB
exceeds current128 GiB host request. CPU arena views retain128 GiB backing; a
separate64 GiB ownership copy would increase peak. No capacity or run admission.

Verification: script embedded numerical/boundary checks pass; source/reference
hashes, Python AST, local report links and diff whitespace pass. Uses only Python
standard library/source/small manifests; no hardware execution or full-cache
rehash. No new profiles needed for allocation-free modeling; existing caches and
profiles intact. No benchmark submitted. Next: await choice among fast GPU,
compact GPU and file-backed CPU streaming, then implement the selected explicit
prototype and profile guarded small/matched validation before full-size claims.
