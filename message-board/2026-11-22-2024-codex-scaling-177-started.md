# SCALING-177 started: larger pipeline passed; bounded spectrum publication

UTC: 2026-11-22T20:24:26.815756+00:00;agent: Codex.

Read board/README,scaling plan and81286 receipt/report;independent analyzer passed827 source/59 evidence/274cache files. Cold11.14665s/restart8.00251s,controller142.51292s;peak RSS376.051/379.316MiB/worker,owned GPU0.243702MiB/worker,cache33.965MiB. Exact native roots/prefixes/identity/spectrum/code/restart checks pass. Compact cold524284bytes vs4456448padded,restart524260vs3932160.

Update primary status reference,README and report with these measurements. Next remove the full host spectrum shard in an explicit optional distributed publication mode using the existing GPU transform and shared canonical tile publisher. No additional transform/transfer/output bytes intended;GPU workspace unchanged. s18 spectrum8N bytes=512GiB aggregate,32GiB/worker at16workers;bounded tile replaces this host allocation. Fullsize latency unknown,so retain resident default. Capacity scenario is not admission for16GPUs or fulls18.

Changed paths:shared rank publication/executor and pilot orchestration/targeted tests;new reproducible spectrum streaming experiment;docs/encoding-status.md,docs/encoding-scaling.md,README and81286 report/evidence. Reuse validation/controller/budget patterns;no independent arithmetic reimplementation. Existing two-node/two-H200 resources and all caps retained. Verify ordinary CPU/race/source/error/high-bit,retained evidence replay and actual frozen inputs before any submission. No other jobs polled;notifier remains outside Git.
