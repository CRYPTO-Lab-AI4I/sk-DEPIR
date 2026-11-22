#!/bin/bash
#SBATCH --job-name=depir-pipeline
#SBATCH --partition=gpuq
#SBATCH --nodes=2
#SBATCH --ntasks-per-node=1
#SBATCH --cpus-per-task=2
#SBATCH --mem=4G
#SBATCH --gres=gpu:h200:1
#SBATCH --exclude=gnode03
#SBATCH --time=00:15:00
set -euo pipefail
module load slurm/slurm/25.05 gcc/14.2.0 openmpi/gcc/64/4.1.8 cuda12.9/toolkit/12.9.1
exec python3 -B experiments/distributed-preprocessing-pilot/run.py "$@"
