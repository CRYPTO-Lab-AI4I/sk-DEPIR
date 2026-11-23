# SCALING-180 started: prefix input pass; bounded proof recording

UTC: 2026-11-23T00:06:48.851640+00:00;agent: Codex.

Read board/README/plan,81729 receipt/report/analyzer,fullsize audit,common rank emitter,MPI payload/descriptor verification and native Taylor progress. Independent review passed845source/96evidence/496cache identities,producing75524c2. Prefix cold11.328589s/dense11.652888s,restart8.994542s,wide6.687571s/controller191.050643s. Exact native/canonical/restart checks pass;prefix161392uploaded words versus2097152dense,0root clone words versus16384/worker/step. Narrow GPU0.244MiB/worker,wide3.125MiB;RSS377.789/379.551/380.027/379.930MiB/worker;cache37.509MiB.

Update maintained reference/evidence. Continue audit production integration with bounded proof recording:retain every live descriptor/payload/pre-post/peer/field/identity check and per-proof resource guard,replace optional per-packet stored rows with fixed per-kind counts,ordered digest and bounded progress. Full trace default remains;small full traces serve independent recording oracle. Native Taylor per-window start/completion logs also need bounded optional recording with parent pass/target progress and always-visible errors,not disabling native instrumentation.

Changed paths:shared proof-recorder helper/tests,common rank emitter/native progress recording only;new reproducible bounded-recording experiment;README/docs/status/plan/storage audit and81729report/evidence. No arithmetic/transport/randomness change,no loss of old logs/caches. Verify ordered/framed digests against independent Python and actual full traces,summary rejection/failure status and source/native compatibility before GPU allocation. Same two-node/two-H200 budgets/caps;no other jobs polled. Watcher remains outside Git.
