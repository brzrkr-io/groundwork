#!/usr/bin/env bash
set -euo pipefail

# Task 4: hand settings to a child program through the environment.
# Usage:   bash 4_run_check.sh ENV
# Example: bash 4_run_check.sh staging
#          DEPLOY_ENV=staging
#          LOG_LEVEL=debug
#
# testdata/check_env.sh must see DEPLOY_ENV set to $1, and LOG_LEVEL set to
# the caller's LOG_LEVEL if it is set and not empty, otherwise debug.

# ${LOG_LEVEL:-debug} also covers an unset LOG_LEVEL, which set -u would reject.
export DEPLOY_ENV="$1"
export LOG_LEVEL="${LOG_LEVEL:-debug}"
bash testdata/check_env.sh
