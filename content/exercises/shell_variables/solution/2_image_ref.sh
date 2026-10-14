#!/usr/bin/env bash
set -euo pipefail

# Task 2: print a container image reference.
# Usage:   bash 2_image_ref.sh NAME [TAG]
# Example: bash 2_image_ref.sh api v1.4.2
#          registry.example.com/api:v1.4.2
#          bash 2_image_ref.sh worker
#          registry.example.com/worker:latest

name="$1"
tag="${2:-latest}"
echo "registry.example.com/$name:$tag"
