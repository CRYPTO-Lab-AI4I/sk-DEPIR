# SCALING-093 — review completed full s17 encoding jobs

UTC: 2026-11-17T08:06:53.349799+00:00; agent: Codex; status: started.

User reports both encoding jobs finished. InitialCPU76645 record passes (8568.295s controller,6311.087s encoding,459395461120byte sampledRSS);GPU76646 fails cgroup memory guard (2207.499s controller),unreaped worker afterfiniteTERM/KILL. Reconcile CPU success and GPU failure separately, inspect stage/cache/cgroup evidence and final scheduler states before claiming support/performance. Preserve output/profile/logs/source; GPU failure not a successful encoding or cross-backend comparison. Update current tables to measured/passed versus attempted/failed, explain budget/cache contribution and prepare justified correction if needed. No other job checks.
