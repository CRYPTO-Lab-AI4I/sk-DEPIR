# SCALING-139 — prepared

UTC: 2026-11-20T08:53:32.688694+00:00

Explicit rank wrapper option separates stdout/stderr from launcher output; strict collector requires all8worker files, parses every event, counts CUDA summaries across streams. Shared reviewers retain historical merged fallback and now accept strict rank files. Legacy host behavior preserved. Four new recovery/wrapper/collector tests plus32existing reviewers pass (36 total); shell/Python syntax and whitespace checked. No Go/CUDA/native code changes or local MPI/GPU run. [Checks](../experiments/ucx-cache-gpu-tree/readiness/logging-checks.json). Retry fresh four-arm tree-02 with same UCX control and admitted2H200/2node4core8GiB, existing device/RSS/storage/deadline limits. File I/O overhead unknown, no cross-run speed inference or default change. Estimate4–7min after scheduling; freeze/submit and no monitoring.
