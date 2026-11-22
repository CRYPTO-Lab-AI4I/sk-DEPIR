# SCALING-175 — started

UTC: 2026-11-22T19:29:44.145074+00:00. Agent: Codex.

81079 passed818 source/92 evidence/538 cache seals and42 commands. Controller150.945s;arms cold6.301s,resume5.052s,wide6.355s,owned-wide3.766s. Measurement-row RSS376.793/376.992/379.969/379.785MiB per worker;GPU-owned0.064MiB narrow/3.125MiB wide tail/3.149MiB owned-wide,context/driver extra;cache16.211MiB. All native leaves/stages/final shards,canonical/prefix/identity/spectrum and narrow/wide publication checks passed. [Analysis](../experiments/distributed-owned-roots/runs/owned-01/analysis.json).

Next fullsize persistence boundary:current restart files save an entire zero-padded N/r shard for every Newton prefix. Add an explicitly selected compact prefix layout,retaining legacy padded reads/default and the same rank/support/hash/canonical checks. Omit only verified zero padding;derive exact stored intervals from prefix/shape/rank metadata;restore an admitted zero-padded array with bounded I/O. Keep current source support verification cost unchanged and avoid rereading global support during in-process local reload. Validate cold/fresh-process restart and exact storage accounting under existing resource caps. This affects restart data,not packed database encoding. Extra CPU padding checks and real runtime remain explicit/unmeasured until validation;no slower production default selected.

Read board,README,recovery implementation/tests,pipeline persistence and artifact reviewers. Update main scaling reference with81079 and the storage change. No other jobs polled.
