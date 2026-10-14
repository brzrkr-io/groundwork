#!/usr/bin/env bash
set -euo pipefail

# Task 2: announce a restart.
# Usage:   bash 2_restart.sh SERVICE
# Example: bash 2_restart.sh payments
#          Restarting payments on all nodes

service="$1"
# TODO(2): this prints "$service" as it is, not the service's name. Fix it.
echo 'Restarting $service on all nodes'
