#!/usr/bin/env bash
set -euo pipefail

# Task 1: print the region to deploy to.
# Usage:   bash 1_region.sh [REGION]
# Example: bash 1_region.sh
#          Deploying to eu-west-1
#          bash 1_region.sh us-east-2
#          Deploying to us-east-2

# No spaces around =, or bash runs "region" as a command.
region="${1:-eu-west-1}"
echo "Deploying to $region"
