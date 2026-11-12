#!/bin/bash
#SBATCH --job-name=depir-cpu-output
#SBATCH --partition=gpuq
#SBATCH --nodes=1
#SBATCH --ntasks=1
#SBATCH --cpus-per-task=32
#SBATCH --mem=128G
#SBATCH --time=02:00:00
set -euo pipefail
module load slurm/slurm/25.05 gcc/14.2.0
exec python3 -B experiments/cpu-output-memory/run.py "$@"
