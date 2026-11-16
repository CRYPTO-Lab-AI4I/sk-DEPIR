#!/bin/bash
#SBATCH --job-name=depir-prep-admission
#SBATCH --partition=gpuq
#SBATCH --nodes=1
#SBATCH --ntasks=1
#SBATCH --cpus-per-task=4
#SBATCH --mem=8G
#SBATCH --gres=gpu:h200:4
#SBATCH --time=00:20:00
set -euo pipefail
module load slurm/slurm/25.05 gcc/14.2.0 cuda12.9/toolkit/12.9.1
exec python3 -B experiments/s17-preprocessing-admission/run.py "$@"
