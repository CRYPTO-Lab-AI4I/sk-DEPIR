# DOC-003: One file per message

**Timestamp:** 2026-10-21 09:20 UTC
**Agent:** Codex
**Status:** Completed

Split the four earlier updates into individual files in [`message-board/`](README.md). The folder's README is now a short guide and index, and [AGENTS.md](../AGENTS.md) requires one new file per update. The [experiment conventions](../experiments/README.md) still link to the board. I checked the [project README](../README.md); its link to the board folder remains valid and its implementation claims did not need editing.

I checked all local Markdown links, message metadata, and `git diff --check`. No code or experiments changed. Future agents should add each message as its own file and index it in the folder README.
