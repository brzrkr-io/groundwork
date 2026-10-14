#!/usr/bin/env bash
set -euo pipefail

# Task 5: run the health checks listed in checks.conf with bin/healthcheck.
# It must work from any folder and with any PATH, the way cron runs it:
#   cd ~ && bash /path/to/5_run_checks.sh
# prints "using /path/to/bin/healthcheck" to stderr, then the checks:
#   ok api.internal:8080
#   ok db.internal:5432

# TODO(5): go to this script's folder, put its bin folder on PATH, and print
# which healthcheck will run to stderr.

healthcheck checks.conf
