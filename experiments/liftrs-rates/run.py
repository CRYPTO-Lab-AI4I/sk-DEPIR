#!/usr/bin/env python3
"""Reproduce the Lift-RS monomial-rate experiment (no database encoding).

Run from any directory. The checked-in PIR configuration snapshot is the input.
Results include exact counts, deterministic seeds, confidence bounds, source
hashes, compiler/platform, and sampling-only timings. See report.md.
"""
import argparse
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
    "go.mod", "go.sum", "cmd/pir-rate/main.go",
    "encoding/liftrs-shared/monomial.go", "encoding/liftrs-shared/reference.go",
    "encoding/liftrs-shared/field.go", "pirconfig/catalogue.go", "pirconfig/json.go",
    "experiments/liftrs-rates/run.py",
]


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--go", default=os.environ.get("GO", "go"))
    parser.add_argument("--samples", type=int, default=1_000_000)
    parser.add_argument("--seed", type=int, default=20261026)
    parser.add_argument("--confidence", type=float, default=0.99)
    args = parser.parse_args()
    source_hashes = {p: hashlib.sha256((ROOT / p).read_bytes()).hexdigest()
                     for p in SOURCES}
    revision = subprocess.check_output(
        ["git", "rev-parse", "HEAD"], cwd=ROOT, text=True).strip()
    dirty = bool(subprocess.check_output(
        ["git", "status", "--porcelain"], cwd=ROOT, text=True).strip())
    catalogue = HERE / "pir-configs.json"
    cli_args = ["-catalogue", str(catalogue), "-samples", str(args.samples),
                "-seed", str(args.seed), "-confidence", str(args.confidence)]
    with tempfile.TemporaryDirectory(prefix="sk-depir-rate-") as temp:
        executable = str(Path(temp) / "pir-rate")
        subprocess.run([args.go, "build", "-o", executable, "./cmd/pir-rate"],
                       cwd=ROOT, check=True)
        completed = subprocess.run([executable] + cli_args, cwd=ROOT,
                                   check=True, capture_output=True, text=True)
    result = json.loads(completed.stdout)
    result["experiment"] = {
        "timestamp_utc": datetime.now(timezone.utc).isoformat(),
        "base_revision": revision, "working_tree_dirty": dirty,
        "source_sha256": source_hashes,
        "input": str(catalogue.relative_to(ROOT)),
        "platform": platform.platform(), "machine": platform.machine(),
        "python_version": platform.python_version(),
        "command": ["pir-rate", "-catalogue", str(catalogue.relative_to(ROOT))]
                   + cli_args[2:],
    }
    for key, command in {
        "cpu": ["sysctl", "-n", "machdep.cpu.brand_string"],
        "memory_bytes": ["sysctl", "-n", "hw.memsize"],
    }.items():
        try:
            result["experiment"][key] = subprocess.check_output(
                command, text=True, stderr=subprocess.DEVNULL).strip()
        except (OSError, subprocess.CalledProcessError):
            result["experiment"][key] = "unavailable"
    (HERE / "results.json").write_text(json.dumps(result, indent=2) + "\n")
    print("s  m  t  good/samples          rate       interval")
    for e in result["estimates"]:
        print(f'{e["s"]:2} {e["m"]:2} {e["t"]:2} '
              f'{e["good"]:7}/{e["samples"]:<9} '
              f'{e["rate"]:.6f}  [{e["lower_rate"]:.6f}, {e["upper_rate"]:.6f}]')


if __name__ == "__main__":
    main()
