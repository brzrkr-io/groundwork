#!/usr/bin/env bash
# Task 3: Commit a change
#
# Add the line "healthcheck: /healthz" to the end of config/app.yaml,
# then commit that change with the message:
# Add health check path
#
# This script runs inside the sandbox repo: cd sandbox && bash ../3_add_health_check.sh
set -euo pipefail

echo "healthcheck: /healthz" >> config/app.yaml
git add config/app.yaml
git commit -m "Add health check path"
