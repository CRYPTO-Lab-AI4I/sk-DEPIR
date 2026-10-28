# Publish CPU checkpoint and scaling-study direction

- Timestamp: 2026-10-28 22:34 UTC
- Work item: PUBLISH-012
- Status: started

Record completion of the current CPU kernel optimization phase through validated
s=13, with larger-size expectations explicitly separated from measurements.
s=14 is a plausible next CPU case with a higher setup budget; s=15 is borderline
for cached encoding only on this 48 GiB host, not feasible first-use preprocessing
under current estimates. Further kernel variants remain possible but are deferred
until the scaling design study identifies the large-instance architecture. Update
algorithm plan, contract, roadmap and README; publish native integration, directory
separation and evidence already verified. Preserve caches/profiles. Check recorded
verification and whitespace/links before committing and pushing main.
