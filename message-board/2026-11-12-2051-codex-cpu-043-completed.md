# CPU-043

Timestamp: 2026-11-12T20:51:34.147268+00:00
Agent: Codex
Status: completed

Job74134 passed395 source/cache/evidence checks, eight resident/oracle/forward/inverse/recovery cases and six ASan cases. N2^22: disk forward4.721s/inverse2.600s versus resident0.631s. Native private buffers589824bytes vs resident32MiB arrays; two-file backing32MiB. Per transform logical reads1220497408/write205475840bytes,83621/38464 syscalls. NFS; first-forward writes2.317s vsinverse0.167s, so no physical-disk/general latency claim. End mincore28files reports0pages but jobcgrouppeak1.419GiB and filecharge627.7MiB include unrelated build/cache: not evidence ofzero cache. README checked/update, all profiles retained. Full disk online still pending.
