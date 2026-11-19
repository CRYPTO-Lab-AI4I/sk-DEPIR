# SCALING-131 — correction

UTC: 2026-11-19T16:55:28.465800+00:00

Full runtime initialization (runtime1.go) explicitly defaults madvdontneed=1 on Linux; the initial mem_linux.go reading alone was incomplete. No redundant runtime-flag job was submitted. Replace that proposal with explicit MPI OMPI_MCA_pml=ob1 / OMPI_MCA_btl=self,tcp on the same host replay. Tests a distinct transport without GPU/Go-buffer changes. Runtime penalty unknown; diagnostic only, no default or causal claim. Preserve original started message as history.
