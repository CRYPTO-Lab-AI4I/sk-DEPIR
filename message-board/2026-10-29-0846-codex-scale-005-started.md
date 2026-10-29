# Profile first-use s=14 encoding

Timestamp: 2026-10-29 08:46 UTC. Agent: Codex. Work: SCALE-005. Status: started.

The user requested publication followed by an s=14 run with retained profiles and
preprocessing. Parallel integration is pushed as `9ac6ae8`. Run catalogue m=2,t=5
on the current 48 GiB commodity CPU host. Use sparse code metadata, paired/root-only
Go preparation and native transforms rather than the obsolete full-tree forecast.
Review width/guard changes and validate them on small fixtures before admission.

Policy: 24 GiB Go soft target, 32 GiB process-RSS guard, two-hour case timeout,
20 GiB retained-cache budget. Separate exact discovery, reusable preparation,
artifact save/load, encoding, packing and correctness validation. Persist the
basis before expensive preparation, and retain complete reusable constants,
checksums, provenance and Go/native profiles. Do not rebuild/delete old caches.
These budgets bound this experiment, not the library's supported parameters.

Link all code/harness/doc changes and results from a new reproducible
`experiments/parallel-s14/` report; check README and distinguish experimental
s=14 evidence from existing public/default capabilities.
