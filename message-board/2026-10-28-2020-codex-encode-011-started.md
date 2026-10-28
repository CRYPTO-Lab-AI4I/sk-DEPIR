# Compare language-labelled Go and C++ encoding kernels

- Timestamp: 2026-10-28 20:20 UTC
- Work item: ENCODE-011
- Status: started

After publishing `0431f80`, compare equivalent batched fixed-constant kernels
and a complete additive-transform stage in Go and C++, using shared inputs,
actual code constants and memory budgets. Rename existing `encoding/liftrs/`
to `encoding/liftrs-go/` and add `encoding/liftrs-cpp/`. Keep Go package names
unchanged, update active imports/scripts/docs and Markdown link targets; historical
source/hash/result data and snapshot scripts retain their original provenance.
Document path migration and make active evidence checks resolve recorded sources.

Keep reusable fixtures/caches and profiles; ordinary CPU on this machine first.
Test both implementations against independent arithmetic/transform oracles and
identical output checksums. Measure batch boundaries separately from I/O/setup;
no server computation or full native encoder claim. Record compiler flags/host,
profile availability and any profiler exception in a new experiment report.
