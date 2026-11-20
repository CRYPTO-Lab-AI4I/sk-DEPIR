# SCALING-145 — started

UTC: 2026-11-20T11:04:02.858332+00:00

Review78286: normal GPU suite passed24cases/rank, two ranks; sanitizer480s cutoff during first wide case (136s/transform vs~1.8s normal). Preserve failed raw and independently reconcile normal output bytes/hashes. Adjust sanitizer coverage within existing720s controller/15min allocation: all small cases per rank, distinct wide tail row/column witnesses per rank, retain every normal wide case. Explicitly document changed method/remaining zero-origin wide sanitizer gap. Same resources; no memory-time production/default change. Existing Ponytail workflow, README/reference checked.
