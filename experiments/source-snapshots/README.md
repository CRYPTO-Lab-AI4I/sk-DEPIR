# Historical measured source snapshots

Files are named `<sha256>.txt` and contain exact source bytes from the uncommitted
native-integration measurements before directory separation. They are evidence,
not active implementations. The `.txt` extension prevents Go/C++ tools from
building duplicate historical packages. `experiments/source_evidence.py` resolves
current paths, moved paths, these content-addressed snapshots and Git history
without modifying recorded result manifests. Existing binaries and expensive
preprocessing remain in the ignored local cache. These files do not add new
performance measurements.
