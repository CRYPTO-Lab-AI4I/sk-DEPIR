#!/bin/bash
#SBATCH --job-name=depir-cuda-reuse
#SBATCH --partition=gpuq
#SBATCH --nodes=1
#SBATCH --ntasks=1
#SBATCH --cpus-per-task=16
#SBATCH --mem=128G
#SBATCH --gres=gpu:1
#SBATCH --time=02:00:00
set -euo pipefail
module load slurm/slurm/25.05 gcc/14.2.0 cuda12.9/toolkit/12.9.1
exec python3 -B experiments/cuda-buffer-reuse/run.py "$@"
