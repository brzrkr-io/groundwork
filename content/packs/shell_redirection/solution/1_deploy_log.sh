#!/usr/bin/env bash
set -euo pipefail

# Task 1: Start the deploy log
#
# Write the line "deploy started" to deploy.log, replacing whatever was in
# it. Then add the line "build ok" at the end.
#
# Usage:   bash 1_deploy_log.sh

# > empties the file first; >> adds to the end.
echo "deploy started" > deploy.log
echo "build ok" >> deploy.log
