#!/usr/bin/env bash
set -euo pipefail

# Task 2: announce a restart.
# Usage:   bash 2_restart.sh SERVICE
# Example: bash 2_restart.sh payments
#          Restarting payments on all nodes

service="$1"
# Double quotes: bash swaps in $service. Single quotes would not.
echo "Restarting $service on all nodes"
