#!/bin/bash
#SBATCH --job-name=depir-cpu-modules
#SBATCH --partition=gpuq
#SBATCH --nodes=1
#SBATCH --ntasks=1
#SBATCH --cpus-per-task=4
#SBATCH --mem=8G
#SBATCH --time=00:20:00
set -euo pipefail
module load slurm/slurm/25.05 gcc/14.2.0
exec python3 -B experiments/cpu-encoding-modules/run.py "$@"
