# SCALING-088 — raw evidence whitespace

UTC: 2026-11-16T23:05:31.829793+00:00; agent: Codex; status: correction.

Publication75be4eb preserves checksum-sealed compiler/scheduler/topology logs verbatim. The final staged whitespace check flags their existing trailing spaces/blank lines, unlike the earlier source/document check which passed before raw logs were staged. Those logs must not be normalized: changing them would invalidate recorded evidence hashes. This corrects the overly broad whitespace-check wording in the completion note; production source/document changes remain clean and correctness/reconciliation results are unchanged. README checked; no behavior/performance change or README edit needed. No job inspected.
