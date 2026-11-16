# SCALING-080 — Complete persistent workflow validation passes

- UTC: 2026-11-16T08:17:55.931335+00:00
- Agent: Codex
- Status: Completed

User-requested one-time accounting:75342COMPLETED0:0,3m42s ongnode04, unchanged
resources after pin removal. Saved resultpassed, controller221.366s. Command
python3 -B experiments/gpu-preprocessing-workflow/analyze.py validation-02
reconciles509source/37evidence/73cache identities against209c4f9. [Report](../experiments/gpu-preprocessing-workflow/report.md#completed-persistent-workflow-validation-scaling-080),
[analysis](../experiments/gpu-preprocessing-workflow/runs/validation-02/analysis.json).
All peer-copy/transpose probes, publication/corruption/Go-oracle, race/ASan/vet,
peer/local1/4GPU fresh/root/final/CUDA-free readback, full canonical identities
and fresh executable CUDA memcheck pass; zero errors/leaked bytes. No guard hit.

Updated README, preprocessing guide, scaling plan/reference and experiment report
for completed small-runner gate and remaining fulls17 admission. Earlier gnode03
fault cause unresolved, all evidence/profiles/cache retained. No math/code/
default/budget changes or new jobs. Fulls17 remains unrun/unadmitted. README check
and git diff --check pass; exact script/results/source provenance linked.
