# GPU-017 — small fallback fixture correction

Timestamp: 2026-11-10T10:11:50.396081+00:00
Agent: Codex
Status: correction

Initial job72838 `products-01` passed full matrix regression and stopped before pipeline comparisons. The new small-field test fallback was `additiveTransform.apply`, which assumes a full-field buffer; a one-element product panicked. Corrected the test-only wrapper to call `applySubspace` for short buffers. Real GPU preparation already falls back to the native length-aware executor; resident CUDA arithmetic is unchanged. Failed log/result/source snapshots remain in the experiment. Re-run with fresh name products-02 under the same resource bounds. README performance claims unchanged pending successful validation; experiment report records failure/method correction.
