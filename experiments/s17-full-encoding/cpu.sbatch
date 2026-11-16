#!/bin/bash
#SBATCH --job-name=depir-encode-s17-cpu
#SBATCH --partition=gpuq
#SBATCH --nodes=1
#SBATCH --ntasks=1
#SBATCH --cpus-per-task=32
#SBATCH --mem=768G
#SBATCH --exclude=gnode03
#SBATCH --time=2-00:00:00
set -euo pipefail
module load slurm/slurm/25.05 gcc/14.2.0
exec srun --cpu-bind=cores python3 -B experiments/s17-full-encoding/run.py cpu "$@"
