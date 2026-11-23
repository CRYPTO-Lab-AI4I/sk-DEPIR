# Distributed GPU preprocessing command

`pir-preprocess-mpi` is an experimental Linux server entry around the shared
MPI/GPU preprocessing runtime. It accepts m=2,s=2..18 in source; full s18 execution
is **not validated or resource-admitted**. CPU client encoding and the earlier
single-node `pir-preprocess` command are unchanged. No CUDA/MPI dependency is added
to ordinary builds; requesting distributed execution without the explicit build
fails rather than falling back.

```sh
module load gcc/14.2.0 openmpi/gcc/64/4.1.8 cuda12.9/toolkit/12.9.1
CGO_ENABLED=1 CC=mpicc CXX=mpicxx go build \
 -tags liftrs_cpp,liftrs_cuda,liftrs_mpi -o pir-preprocess-mpi ./cmd/pir-preprocess-mpi
./pir-preprocess-mpi --config /absolute/job.json --check-config
# Inside an admitted allocation, create a fresh shared output directory and use
# a fail-fast MPI launcher. rank.sh supplies separate logs and fresh CPU profiles.
srun --mpi=pmix_v3 --kill-on-bad-exit=1 bash cmd/pir-preprocess-mpi/rank.sh \
 /absolute/log-prefix /absolute/pir-preprocess-mpi --config /absolute/job.json
```

The launcher must allocate the agreed cores/host RAM/GPUs and consistent visible
CUDA ordinals. The shared directory and library must exist on every rank; each
process checks them before MPI work. The command is not a scheduler/resource
approval mechanism. Never turn a small example into a full run just by increasing
limits. [Bounded validation](../../experiments/distributed-command/report.md).

Example policy for the admitted **s8,4workers/two H200s** validation (JSON byte
values are bytes, `max_*_words` are element counts):

```json
{
  "spec": {"FieldDegree": 8, "Dimension": 2, "CurveDegree": 5},
  "ranks": 4,
  "directory": "/absolute/fresh-shared-output",
  "library": "/absolute/libliftrs-cuda.so",
  "device": 0,
  "device_bytes": 67108864,
  "max_local_words": 16384,
  "max_good_words": 65536,
  "max_root_words": 65536,
  "max_archive_words": 0,
  "rss_bytes": 1073741824,
  "storage_bytes": 33554432,
  "file_bytes": 33554432,
  "deadline_seconds": 220,
  "newton_input": "prefix",
  "spectrum": "stream",
  "discovery_result": "stream",
  "proof_log": "trace"
}
```

All memory/log modes are explicit. `dense` Newton inputs, `resident` spectrum and
`resident` discovery results remain selectable; resident discovery also needs a
nonzero archive-element cap. Streaming/prefix policies are not silently selected
when a faster resident path fits. `trace` retains full proof rows and summaries;
`summary` retains bounded ordered hashes/counts, preserving runtime checks;
`full` keeps historical full rows without summary hash work. Logging costs are
not assumed free. Native progress defaults to bounded Taylor logs and enabled CPU
progress; explicit existing environment settings are retained.

Cold discovery uses `crypto/rand.Reader`. `--test-seed=N` explicitly selects
reproducible validation sampling and must not be mistaken for secure randomness.
Experiment receipts retain the exact command/seed; direct users must retain their
launch configuration/source/library provenance too.

Restart in a **different fresh output directory**, adding `resume_directory`,
`source_sha256` and `known_prefix` to the same code/rank/mode policy. The source
hash is that of the trusted original `sN-checkpoint-source.json`, not a digest
accepted from an untrusted producer. Root and selected Newton manifests bind it;
rank0 hashes support once, all ranks agree metadata and validate their own shards.
The shared numerical tail finishes Newton, verifies the full identity, and publishes
code/generator with a reused immutable archive. Same-filesystem hard links are
required. Keep the old cache immutable while aliases exist. No resharding or
recovery from unfinished discovery/root construction is implemented.

The command owns MPI on its calling thread and aborts peers on runtime failure.
Profiles are mandatory, fresh and per rank. Its one-second watchdog checks peak
RSS, elapsed time and logical output storage; SIGTERM/resource failure flushes the
profile, exits nonzero and leaves completed commits. Use a fail-fast launcher so
peers terminate too. This sampler can overshoot briefly and is additional to Slurm
hard limits; it is not a complete memory model. Filesystem walks tolerate names
renamed during atomic commits; symlinks in owned output are rejected. File-size
rlimits cover profiles as well as payloads. Profiles/logs are outside the output
storage budget and must have their own admitted controller storage.

Once exact K is known, before GPU discovery allocations, admission accounts for
all canonical payloads, compact root/Newton checkpoints, logical archive aliases
and bounded metadata. It repeats before later root/Newton phases. Existing list,
local-shard, rounded root-stage and actual CUDA owned-byte guards remain. Admission
currently checks named allocations and runtime RSS; **complete fullsize host/GC/
native working-set headroom and larger topology still need review and measurements**.
This command alone does not establish s18 feasibility or TRL5 acceptance.
