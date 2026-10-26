#!/usr/bin/env python3
"""Run the fixed 21-case application-plan cost grid; no database is encoded.

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
    "go.mod", "go.sum", "planner/evaluate.go", "planner/costs.go",
    "planner/greedy.go", "planner/application.go", "pirconfig/catalogue.go", "pirconfig/json.go",
    "experiments/application-costs/main.go", "experiments/application-costs/run.py",
]


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--go", default=os.environ.get("GO", "go"))
    args = parser.parse_args()
    hashes = {p: hashlib.sha256((ROOT / p).read_bytes()).hexdigest()
              for p in SOURCES}
    snapshot = HERE / "pir-configs.json"
    with tempfile.TemporaryDirectory(prefix="sk-depir-application-costs-") as temp:
        executable = str(Path(temp) / "application-costs")
        subprocess.run([args.go, "build", "-o", executable,
                        "./experiments/application-costs"], cwd=ROOT, check=True)
        completed = subprocess.run([executable, "-catalogue", str(snapshot)],
                                   cwd=ROOT, check=True, capture_output=True, text=True)
    result = json.loads(completed.stdout)
    assert len(result["scenarios"]) == 21
    assert all(case["application_bytes"] // case["record_bytes"] >= 5000
               for case in result["scenarios"])
    assert any(case["application_bytes"] // case["record_bytes"] == 5000
               for case in result["scenarios"])
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
        "command": ["go", "run", "./experiments/application-costs", "-catalogue",
                    str(snapshot.relative_to(ROOT))],
    }
    (HERE / "results.json").write_text(json.dumps(result, indent=2) + "\n")
    fields = ["application_bytes", "record_bytes", "rows", "row_blocks",
              "logical_record_bits", "fragments", "full_download_bits",
              "storage_fraction", "upload_fraction", "download_fraction", "communication_fraction",
              "pir_databases", "storage_bits", "upload_bits", "download_bits",
              "indexed_storage_fraction", "indexed_communication_fraction",
              "effective_storage_rate", "provisional", "row_configurations"]
    with (HERE / "summary.csv").open("w", newline="") as file:
        writer = csv.DictWriter(file, fieldnames=fields, lineterminator="\n")
        writer.writeheader()
        for case in result["scenarios"]:
            plan, evaluation = case["plan"], case["evaluation"]
            costs = evaluation["costs"]
            writer.writerow({
                "application_bytes": case["application_bytes"],
                "record_bytes": case["record_bytes"], "rows": plan["rows"],
                "row_blocks": len(plan["logical_plan"]["row_blocks"]),
                "logical_record_bits": plan["logical_record_bits"],
                "fragments": (plan["record_bits"]+plan["logical_record_bits"]-1)//plan["logical_record_bits"],
                "full_download_bits": evaluation.get("full_download_bits",0),
                **{name+"_fraction": costs[name+"_bits"]/costs["application_bits"] for name in ["storage","upload","download","communication"]},
                "pir_databases": evaluation["database_count"],
                "storage_bits": costs["storage_bits"],
                "upload_bits": costs["upload_bits"],
                "download_bits": costs["download_bits"],
                "indexed_storage_fraction":case["indexed_alternative"]["costs"]["storage_bits"]/costs["application_bits"],
                "indexed_communication_fraction":case["indexed_alternative"]["costs"]["communication_bits"]/costs["application_bits"],
                "effective_storage_rate": evaluation["effective_storage_rate"],
                "provisional": evaluation["provisional"],
                "row_configurations": ";".join(("download-all" if b.get("download_all") else b["configuration_id"]) for b in plan["logical_plan"]["row_blocks"]),
            })
    print(f"Wrote {len(result['scenarios'])} scenarios to results.json and summary.csv")


if __name__ == "__main__":
    main()
