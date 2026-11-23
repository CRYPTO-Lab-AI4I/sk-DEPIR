# SCALING-193 prepared

UTC: 2026-11-23T16:13:21.890133+00:00. Agent: Codex. Status: prepared.

82102verified passed:920source/105evidence/476cache seals, all four arms and all
remaining arithmetic/native/proof/checkpoint/final-byte/alias checks. Controller
191.398s; assembled restart8.297s/381.387MiB peak RSS per worker. Evidence/profiles
retained in [restart report](../experiments/distributed-workflow-resume/report.md).

Added [MPI server command](../cmd/pir-preprocess-mpi/README.md), strict explicit
JSON policy and optional backend-unavailable stub, shared MPI lifecycle/secure
caller randomness/cold+restart/publication, mandatory profiles, local watchdogs,
resource/path preflight, and exact-K output/checkpoint admission before GPU
discovery allocations. No test or oracle dependency in the executable; old CPU
client and single-node command unaffected. Fullsize host/GC/native phase headroom
and larger topology still need review; runtime guards are not a complete model.

[Command report](../experiments/distributed-command/report.md) links code/scripts,
limits/timing/storage changes and validation. Actual executable cold/restart will
be compared with pilot cold/restart; omitted diagnostic support/root-stage/identity
copies are documented, all live/math/proof checks retained. Shared vector launcher
adds explicit command/wrapper/profile-binary overrides, preserving test defaults.
A real shell-wrapper regression verifies quoting and CLI profile flags.

Ordinary CPU and race tests pass; actual tagged command/reference binary builds
and tagged CPU admission tests pass without MPI/GPU initialization.26Python suites,
script syntax and authored whitespace checks pass. Actual --check-config smoke
passed without backend initialization. Frozen-source/native preflight remains
before same-budget submission. README checked/updated and primary scaling table,
preprocessing/plan/storage docs now mark82102passed and command prepared.
No fullsize admission or command GPU pass claimed. Next submit bounded command
cold/restart under existing caps, then complete phase/topology/resource admission.
