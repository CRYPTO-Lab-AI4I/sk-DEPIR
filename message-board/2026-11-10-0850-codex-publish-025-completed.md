# PUBLISH-025 — checkpoint ready for publication

Timestamp: 2026-11-10T08:50:47.457137+00:00
Agent: Codex
Status: completed

Expanded the current step in [encoding scaling](../docs/encoding-scaling.md#current-step-larger-memory-preparation-efficiency), linked from README, ROADMAP and s16 gate. Sub-steps cover targeted profiling, specialized/native CPU products, bounded batch parallelism, root-generation reuse, matched GPU trials with setup/transfers, then evidence-based s16 admission. Initial targeted CPU profile budget is four cores, 16 GiB requested/12 GiB RSS guard, 30-minute allocation/25-minute controller deadline, at most 65,536 saved support exponents and 2^22 transform points. No GPU or full retry is launched by this checkpoint.

Publishing the staged first s16 attempt/stop evidence, private lookup improvement, experiment reports/scripts/results/source snapshots and message-board history. Re-ran both comparison and validation-02 evidence analyzers successfully; prior full suites/race/vet/native checks remain applicable to unchanged arithmetic. README and documentation links checked; staged whitespace check follows. Large caches/binaries remain outside Git. Commit/push requested explicitly; targeted profile follows publication.
