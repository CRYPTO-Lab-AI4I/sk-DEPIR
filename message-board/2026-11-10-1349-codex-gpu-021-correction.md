# GPU-021 — validation setup correction

Timestamp: 2026-11-10T13:49:28.623539+00:00
Agent: Codex
Status: correction

Job72871/validation-01 failed before GPU integration execution because the new runner omitted make -C encoding/rme_legacy rme. Full Go suites passed Lift-RS packages but two PIR placeholder tests could find neither that C++ binary nor Julia. Add the same legacy build step as the existing validator, preserve failed manifest/logs, rerun fresh validation-02. No actual s16 run has been submitted.
