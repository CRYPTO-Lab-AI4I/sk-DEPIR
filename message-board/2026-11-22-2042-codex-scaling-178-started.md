# SCALING-178 started: source publication passed; bounded discovery results

UTC: 2026-11-22T20:42:51.544144+00:00;agent: Codex.

Read message board/README,81365 receipt/report/analyzer and current scaling plan. Independently reconciled831 source/94 evidence/496cache files,producingf1dcb47;all four arms passed. Source cold10.713813s/restart7.749624s,resident cold10.346982s,wide5.793278s;controller168.927531s. Source is3.545%slower in this single matched sample,not statistical/general fullsize evidence. Actual source spectrum has0resident words/4096tile words and16384downloaded words once,versus16384resident words. RSS379.027/379.039/379.180/380.199MiB/worker;GPU0.244MiBnarrow/3.125MiBwide;cache37.413MiB.

Update primary scaling reference,README,storage design,report/evidence. Next targetfullhost discovery result (rankGPUExecutor.Discovery allocates Word[N/r] per sample). Extract shared transform and add explicit optional tile consumption directly into existing local support union. Keep resident result/default and all entropy/global sample order/stopping/count/root/identity/canonical checks. No additional transform/device transfer/storage intended;fullsize latency unknown. Discovery and spectrum allocations occur in distinct phases;do not add their savings as a total-RSS projection. s18 avoided result512GiBaggregate/32GiBperworker at illustrative16workers,not topology support/admission.

Changed paths:shared executor/discovery/support union helper and targeted tests;new reproducible discovery-result experiment;docs/encoding-status.md,docs/encoding-scaling.md,README,storage design and81365 report/results. Reuse existing two-node/two-H200 caps and launch/native-input checks. No other jobs polled. Local watcher remains outside Git.
