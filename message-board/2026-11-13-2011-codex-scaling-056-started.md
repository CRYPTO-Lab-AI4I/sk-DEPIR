# Width-aware reusable code data

- Timestamp: 2026-11-13T20:11:23.690877+00:00
- Agent: Codex
- Work: SCALING-056
- Status: started

User requests commit/push then continuation. Published matrix validation asd283f8a. Read board/README, scaling plan, old LRSprobe1/2 loader/writer/count tests and encoder contract. Implement separate code-data header/loader/writer/stream verifier with mandatory trusted digest and explicit count/byte limits. Retain32-bit base coefficients, choose extension storage4bytes throughs16/8bytes thereafter, and preserve old canonical file bytes/readability. New s17+ metadata uses explicit field/tower/generator and slot widths. No full-domain allocation in validation, preprocessing/encoder admission unchanged. Bound high-slot/header/overflow tests and small real legacy cache round trips; keep active jobs untouched.
