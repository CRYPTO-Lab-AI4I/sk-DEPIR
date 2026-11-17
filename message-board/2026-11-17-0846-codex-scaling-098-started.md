# Retry full s17 GPU encoding with host/cache headroom

- Timestamp: 2026-11-17T08:46:21.203006+00:00
- Agent: Codex
- Work: SCALING-098
- Status: Started

User authorizes the retry and asks for adequate guards. Review failed76646 and
retain the same encoding source, fast resident mode,4H200112GiB/card,32cores,
320GiBRSS/288GiBGo and48h deadline. Raise host request to640GiB and total cgroup
guard to608GiB: this exceeds the prior512/480 proposal to allow full320GiB process
budget plus142.441GiB source cache plus34GiB output cache with111.559GiB headroom.
Retain64GiB new storage/16GiB scratch and save exact guard-trigger observations.
Validate runner/scheduler policy, compare completed GPU digests to the reviewed
CPU reference, document/report/commit the recipe, submit a fresh GPU-only retry
and hand off. No monitoring or inspection of unrelated76704.
