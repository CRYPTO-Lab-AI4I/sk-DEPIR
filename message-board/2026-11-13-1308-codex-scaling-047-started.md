# Relaunch remaining CPU table cases

- Timestamp: 2026-11-13T13:08:19.066693+00:00
- Agent: Codex
- Work: SCALING-047
- Status: started

User requests relaunch of74201/74202/74203 with adequate limits. Read board/README/controller and completed s14 disk/s15 file/packed phase summaries. s14 disk encodings758/684s; s15 disk first encoding remains active after~80minutes. Partition gpuq allows100hours. Use explicit extended retry limits: s16 packed4h, s15 disk12h, s16 disk48h with shorter aligned internal deadlines. Memory/cores/storage caps unchanged; preserve all old evidence and partial work. Cancel old queued/running copies and submit fresh fill-02 paths, packed first so disk latency cannot delay that result. No monitoring after submission.
