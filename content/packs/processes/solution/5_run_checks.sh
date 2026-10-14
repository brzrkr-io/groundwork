#!/usr/bin/env bash
set -euo pipefail

# Task 5: run the health checks listed in checks.conf with bin/healthcheck.
# It must work from any folder and with any PATH, the way cron runs it:
#   cd ~ && bash /path/to/5_run_checks.sh
# prints "using /path/to/bin/healthcheck" to stderr, then the checks:
#   ok api.internal:8080
#   ok db.internal:5432

cd "$(dirname "$0")"
PATH="$PWD/bin:$PATH"
echo "using $(command -v healthcheck)" >&2

healthcheck checks.conf
