#!/bin/bash
#SBATCH --job-name=depir-cpu-regression-s14
#SBATCH --partition=gpuq
#SBATCH --nodes=1
#SBATCH --ntasks=1
#SBATCH --cpus-per-task=32
#SBATCH --mem=32G
#SBATCH --exclude=gnode03
#SBATCH --time=02:00:00
set -euo pipefail
module load slurm/slurm/25.05 gcc/14.2.0
exec srun --cpu-bind=cores python3 -B experiments/cpu-encoding-regression/run.py "$@"
