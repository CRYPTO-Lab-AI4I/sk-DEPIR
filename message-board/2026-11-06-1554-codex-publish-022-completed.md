# PUBLISH-022 completed: s15 checkpoint ready for publication

Timestamp: 2026-11-06T15:59:16.795308+00:00. Agent: Codex.

Reconciled the completed [local s15 run](../experiments/parallel-s15/report.md) with
root README, roadmap, encoder/scaling/algorithm docs and implementation READMEs.
Exact K=216,914,047 and polynomial 0x8003 now have standalone supported/verified
[catalogue evidence](../pirconfig/configs_catalogue.json); historical paper/MC
estimates and planner snapshots are preserved. Median native encoding is 393.19 s,
fresh cached 393.69 s, with 20.55/18.46 GiB whole-process peak RSS. No application,
cryptographic integration or production-security claim is added. Preparation is
explicitly interrupted by the manual break, rather than an unchanged latency
baseline. Large cache files remain outside Git; profiles and source provenance
are retained. Seventeen monitor-native samples are losslessly compressed with
raw copies retained locally and a checksum manifest.

Publication checks found a completion-order race: the original runner exposed
passed before sealing its manifest, allowing a derived analysis file into initial
evidence. Its exact captured bytes are preserved and verified separately; raw
results are unchanged. Future runner status stays finalizing until hashes are
sealed, supervisors wait, and derived analysis is excluded. A publication smoke
and repeated analysis validate this correction. Catalogue/planner tests, three
policy tests, Python compilation, source/binary/artifact/evidence verification,
updated documentation links and diff whitespace pass. Prior full Go/native
suites, race checks and vet remain valid; no new arithmetic change requires their
repetition. [Verification](../experiments/parallel-s15/publication-verification.json)
records current sources and commands/outcomes.

Next recommendation is the [s16 memory-design gate](../docs/encoding-scaling.md#completed-local-s15-checkpoint-and-next-step):
audit retained s15 profiles/lifetimes, specify buffer reuse/fusion/blocking and
2^32 cardinality changes, compare bounded-RAM and larger commodity-RAM designs,
then validate against retained small/s15 references before a new large-run budget.
No s16 implementation or benchmark is started here. Commit/push follows staged
verification. README check completed and current status updated.
