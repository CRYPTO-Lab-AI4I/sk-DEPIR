#!/bin/bash
#SBATCH --job-name=depir-s18-full
#SBATCH --partition=gpuq
#SBATCH --nodes=8
#SBATCH --ntasks-per-node=2
#SBATCH --cpus-per-task=4
#SBATCH --mem=864G
#SBATCH --gres=gpu:h200:2
#SBATCH --exclude=gnode03
#SBATCH --time=4-04:00:00
set -euo pipefail
module load slurm/slurm/25.05 gcc/14.2.0 openmpi/gcc/64/4.1.8 cuda12.9/toolkit/12.9.1
exec python3 -B experiments/distributed-s18-full-preprocessing/run.py "$@"
