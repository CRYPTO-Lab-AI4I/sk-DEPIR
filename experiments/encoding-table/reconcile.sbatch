#!/bin/bash
#SBATCH --job-name=depir-table-review
#SBATCH --partition=gpuq
#SBATCH --nodes=1
#SBATCH --ntasks=1
#SBATCH --cpus-per-task=4
#SBATCH --mem=8G
#SBATCH --time=00:20:00
set -euo pipefail
module load slurm/slurm/25.05
exec python3 -B experiments/encoding-table/reconcile.py "$@"
