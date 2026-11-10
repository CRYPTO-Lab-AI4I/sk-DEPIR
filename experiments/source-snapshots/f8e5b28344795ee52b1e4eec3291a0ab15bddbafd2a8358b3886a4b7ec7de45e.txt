#!/bin/bash
#SBATCH --job-name=depir-s16
#SBATCH --partition=gpuq
#SBATCH --nodes=1
#SBATCH --ntasks=1
#SBATCH --cpus-per-task=32
#SBATCH --mem=256G
#SBATCH --time=24:00:00
set -euo pipefail
module load slurm/slurm/25.05 gcc/14.2.0
exec python3 experiments/hpc-s16/run.py "$@"
