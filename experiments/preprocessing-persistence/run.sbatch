#!/bin/bash
#SBATCH --job-name=depir-persist
#SBATCH --partition=gpuq
#SBATCH --nodes=1
#SBATCH --ntasks=1
#SBATCH --cpus-per-task=2
#SBATCH --mem=4G
#SBATCH --exclude=gnode03
#SBATCH --time=00:15:00
set -euo pipefail
module load slurm/slurm/25.05 gcc/14.2.0
exec python3 -B experiments/preprocessing-persistence/run.py "$@"
