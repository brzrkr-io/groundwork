#!/usr/bin/env bash
set -euo pipefail

# Task 1: print the region to deploy to.
# Usage:   bash 1_region.sh [REGION]
# Example: bash 1_region.sh
#          Deploying to eu-west-1
#          bash 1_region.sh us-east-2
#          Deploying to us-east-2

# TODO(1): this line stops the script with "region: command not found".
# Fix it.
region = "${1:-eu-west-1}"
echo "Deploying to $region"
