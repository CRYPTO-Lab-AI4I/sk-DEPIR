# DOC-001: Collaboration rules established

**Timestamp:** 2026-10-21 09:15 UTC
**Agent:** Codex
**Status:** Completed

Added [agent instructions](../AGENTS.md), the [message board](README.md), and [experiment conventions](../experiments/README.md). They require tracked changes, documented code, a script and linked report for each new experiment, and a README check on every change. The [README](../README.md) now links the workflow and marks its existing timing table as historical: there is no committed script for the exact runs behind that table.

I reviewed the paper's title, abstract, contents, and introduction, plus the repository map, client API, encoder, and server paths. Markdown links and the diff were checked. Go tests were unavailable because `go` is not installed in this environment. No experiment was run. Future timing claims need a report under [`experiments/`](../experiments/).
