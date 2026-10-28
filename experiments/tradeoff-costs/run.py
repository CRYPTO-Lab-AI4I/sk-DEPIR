#!/usr/bin/env python3
"""Run the fixed 21-case bounded tradeoff cost grid; no database is encoded.

Uses the checked-in configuration snapshot, builds the Go driver, and writes
results.json and summary.csv beside this script. See report.md for methodology.
"""
import argparse
import csv
import hashlib
import json
import os
from pathlib import Path
import platform
import subprocess
import tempfile
from datetime import datetime, timezone

ROOT = Path(__file__).resolve().parents[2]
HERE = Path(__file__).resolve().parent
SOURCES = [
    "go.mod", "go.sum", "experiments/tradeoff-costs/boundary-search-results.json", "experiments/greedy-costs/results.json", "planner/evaluate.go", "planner/costs.go",
    "planner/greedy.go", "planner/tradeoffs.go", "planner/batches.go", "planner/application.go", "pirconfig/catalogue.go", "pirconfig/json.go",
    "experiments/tradeoff-costs/main.go", "experiments/tradeoff-costs/run.py",
]


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--go", default=os.environ.get("GO", "go"))
    parser.add_argument("--render-only", action="store_true",
                        help="Regenerate options.md from saved results; do not rerun search.")
    args = parser.parse_args()
    if args.render_only:
        render_options(json.loads((HERE / "results.json").read_text()))
        return
    hashes = {p: hashlib.sha256((ROOT / p).read_bytes()).hexdigest()
              for p in SOURCES}
    snapshot = HERE / "pir-configs.json"
    with tempfile.TemporaryDirectory(prefix="sk-depir-tradeoff-costs-") as temp:
        executable = str(Path(temp) / "tradeoff-costs")
        subprocess.run([args.go, "build", "-o", executable,
                        "./experiments/tradeoff-costs"], cwd=ROOT, check=True)
        completed = subprocess.run([executable, "-catalogue", str(snapshot)],
                                   cwd=ROOT, check=True, capture_output=True, text=True)
    result = json.loads(completed.stdout)
    assert len(result["scenarios"]) == 21
    assert all(case["application_bytes"] // case["record_bytes"] >= 5000
               for case in result["scenarios"])
    assert any(case["application_bytes"] // case["record_bytes"] == 5000
               for case in result["scenarios"])
    # Check against the actual historical table, not only today's baseline code.
    historical = json.loads((ROOT / "experiments/greedy-costs/results.json").read_text())
    old = {(x["application_bytes"], x["record_bytes"]): x["evaluation"]["costs"]
           for x in historical["scenarios"]}
    for case in result["scenarios"]:
        baseline = old[(case["application_bytes"], case["record_bytes"])]
        assert case["original_greedy"] == baseline
        witness = case["original_greedy_witness_index"]
        assert witness in case["representative_indices"]
        costs = case["result"]["alternatives"][witness]["evaluation"]["costs"]
        assert costs["storage_bits"] <= baseline["storage_bits"]
        assert costs["communication_bits"] <= baseline["communication_bits"]
    result["experiment"] = {
        "timestamp_utc": datetime.now(timezone.utc).isoformat(),
        "minimum_application_records": 5000,
        "base_revision": subprocess.check_output(
            ["git", "rev-parse", "HEAD"], cwd=ROOT, text=True).strip(),
        "working_tree_dirty": bool(subprocess.check_output(
            ["git", "status", "--porcelain"], cwd=ROOT, text=True).strip()),
        "source_sha256": hashes,
        "input": str(snapshot.relative_to(ROOT)),
        "catalogue_sha256": hashlib.sha256(snapshot.read_bytes()).hexdigest(),
        "platform": platform.platform(), "python_version": platform.python_version(),
        "command": ["go", "run", "./experiments/tradeoff-costs", "-catalogue",
                    str(snapshot.relative_to(ROOT))],
    }
    (HERE / "results.json").write_text(json.dumps(result, indent=2) + "\n")
    fields = ["application_bytes", "record_bytes", "alternative", "logical_record_bits",
              "storage_bits", "upload_bits", "download_bits", "communication_bits",
              "reference_config_id", "reference_capacity_bits", "reference_storage_overhead",
              "representative", "original_greedy_witness", "storage_fraction", "communication_rate", "full_download_only",
              "frontier_trimmed", "count_choices_trimmed", "state_limit_reached", "baseline_limit_reached", "states_explored"]
    with (HERE / "summary.csv").open("w", newline="") as file:
        writer = csv.DictWriter(file, fieldnames=fields, lineterminator="\n")
        writer.writeheader()
        for case in result["scenarios"]:
            search = case["result"]
            for index, alternative in enumerate(search["alternatives"]):
                plan, evaluation = alternative["plan"], alternative["evaluation"]
                costs = evaluation["costs"]
                writer.writerow({
                    "application_bytes": case["application_bytes"],
                    "record_bytes": case["record_bytes"], "alternative": index,
                    "logical_record_bits": plan["logical_record_bits"],
                    **{name: costs[name] for name in ["storage_bits", "upload_bits", "download_bits", "communication_bits"]},
                    **{name:case[name] for name in ["reference_config_id","reference_capacity_bits","reference_storage_overhead"]},
                    "representative": index in case["representative_indices"],
                    "original_greedy_witness": index == case["original_greedy_witness_index"],
                    "storage_fraction": costs["storage_bits"]/costs["application_bits"],
                    "communication_rate": evaluation["communication_rate"],
                    "full_download_only": evaluation["full_download_only"],
                    **{name: search[name] for name in ["frontier_trimmed", "count_choices_trimmed", "state_limit_reached", "baseline_limit_reached", "states_explored"]},
                })
    render_options(result)
    print(f"Wrote {len(result['scenarios'])} tradeoff scenarios to results.json and summary.csv")


def render_options(result):
    """Render saved plan costs without changing search selection or provenance.

    Download rate is B/D; the displayed inverse D/B includes every returned
    field element, including dummies. A full reference code has B=s and D=L*s,
    hence reference D/B=L, independent of application-record width.
    """
    configs = {c["id"]: c for c in json.loads(
        (HERE / "pir-configs.json").read_text())["configurations"]}
    lines = ["# Selected application tradeoffs", "",
             "Each cell: storage / application · download / record (absolute per-lookup download).",
             "★ means no worse than original greedy in storage and total communication. Whole-download-only options are omitted.", "",
             "| Database | Record | Reference N/K · D/B | Lower-storage option | Minimum communication option |",
             "|---|---|---:|---:|---:|"]
    for case in result["scenarios"]:
        alternatives = [a for a in case["result"]["alternatives"] if not a["evaluation"]["full_download_only"]]
        baseline = case["original_greedy"]
        def matches(a):
            costs = a["evaluation"]["costs"]
            return costs["storage_bits"] <= baseline["storage_bits"] and costs["communication_bits"] <= baseline["communication_bits"]
        indices = []
        if alternatives:
            first = next(i for i, a in enumerate(alternatives) if matches(a))
            last = len(alternatives)-1
            indices = sorted({first, last})
            if len(indices) == 1 and len(alternatives) > 1:
                indices = [last-1, last]
        cells = []
        for index in indices:
            alternative = alternatives[index]
            costs = alternative["evaluation"]["costs"]
            size = costs["download_bits"]/8
            absolute = f"{size/1e9:.3f} GB" if size >= 1e9 else f"{size/1e6:.3f} MB"
            storage = costs["storage_bits"]/(case["application_bytes"]*8)
            inverse_rate = costs["download_bits"]/(case["record_bytes"]*8)
            cells.append(f"{storage:.3f}× · {inverse_rate:,.0f}× ({absolute})" + (" ★" if matches(alternative) else ""))
        if len(cells) == 1:
            cells = ["—", cells[0]]
        cells += ["—"]*(2-len(cells))
        database = "1 TB" if case["application_bytes"] == 10**12 else f"{case['application_bytes']//10**9} GB"
        record = {1000:"1 KB", 10000:"10 KB", 1000000:"1 MB", 10000000:"10 MB"}[case["record_bytes"]]
        lines.append(f"| {database} | {record} | {case['reference_storage_overhead']:.3f}× · {configs[case['reference_config_id']]['query']['length']:,}× | " + " | ".join(cells) + " |")
    lines += ["", "Approximate search: see results.json for state, count and frontier reduction flags. Reference overhead excludes application padding."]
    (HERE / "options.md").write_text("\n".join(lines) + "\n")


if __name__ == "__main__":
    main()
