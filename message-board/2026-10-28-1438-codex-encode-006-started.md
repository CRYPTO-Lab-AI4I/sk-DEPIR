# Profile the current s=10 cyclic encoder

- Timestamp: 2026-10-28 14:38 UTC
- Work item: ENCODE-006
- Status: started

The s=10 checkpoint is committed and pushed as `c06375c`. Extend the existing
profiling driver with code-specification arguments (retaining s=9 defaults),
then record new s=10 cyclic CPU profiles under `experiments/encoder-profile-s10/`.
Separate code-only method preparation from per-database encoding; exclude artifact
loading, input generation, verification and output I/O. Verify artifacts/words
against committed s=10 hashes and all systematic symbols. Preserve historical
profiles. Report cumulative/flat sample attribution and source-line/call-tree
views to distinguish prefix products, generator-product FFTs and coordinate
placement. Update README and the scaling plan with measured optimization priorities;
do not implement an optimization or enable another field as part of this task.
