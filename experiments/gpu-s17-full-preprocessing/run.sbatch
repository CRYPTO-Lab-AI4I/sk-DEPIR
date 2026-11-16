#!/bin/bash
#SBATCH --job-name=depir-preprocess-s17
#SBATCH --partition=gpuq
#SBATCH --nodes=1
#SBATCH --ntasks=1
#SBATCH --cpus-per-task=32
#SBATCH --exclusive
#SBATCH --mem=864G
#SBATCH --gres=gpu:h200:4
#SBATCH --exclude=gnode03
#SBATCH --time=4-04:00:00
set -euo pipefail
module load slurm/slurm/25.05 gcc/14.2.0 cuda12.9/toolkit/12.9.1
exec python3 -B experiments/gpu-s17-full-preprocessing/run.py "$@"
