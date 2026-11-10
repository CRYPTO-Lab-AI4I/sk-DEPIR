#!/bin/bash
#SBATCH --job-name=depir-cpu-kernels
#SBATCH --partition=gpuq
#SBATCH --nodes=1
#SBATCH --ntasks=1
#SBATCH --cpus-per-task=4
#SBATCH --mem=16G
#SBATCH --time=00:30:00
set -euo pipefail
module load slurm/slurm/25.05 gcc/14.2.0
exec python3 -B experiments/preparation-kernels/run.py cpu "$@"
