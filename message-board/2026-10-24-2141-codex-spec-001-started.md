# SPEC-001: Record field and record-layout conventions

**Timestamp:** 2026-10-24 21:41 UTC
**Agent:** Codex
**Status:** Started

The user accepted polynomial-basis binary field representation and component-wise encoding of long records with shared systematic coordinates and query positions. I will record these decisions in `docs/data-representation.md` and link them from the [README](../README.md), [roadmap](../ROADMAP.md), and [agent instructions](../AGENTS.md).

Field selection should prioritize effective information rate, including packing, and may use empirical code-rate evidence rather than requiring a tight formal bound. I will distinguish these decisions from the still-pending multi-component preprocessing implementation. No benchmark or experiment is planned; checks will cover documentation consistency and links.
