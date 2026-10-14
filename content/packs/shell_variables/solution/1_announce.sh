#!/usr/bin/env bash
set -euo pipefail

# Task 1: announce a deploy.
# Usage:   bash 1_announce.sh SERVICE ENV
# Example: bash 1_announce.sh checkout production
#          Deploying checkout to production

service="$1"
env="$2"
echo "Deploying $service to $env"
