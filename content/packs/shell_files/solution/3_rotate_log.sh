#!/usr/bin/env bash
# Task 3: Rotate the log
#
# Inside app/logs, rename app.log to app.log.1, then create a new, empty
# app.log.
#
# This script runs inside the sandbox: cd sandbox && bash ../3_rotate_log.sh
set -euo pipefail

# cd here only changes this script's folder, not your terminal's.
cd app/logs
mv app.log app.log.1
touch app.log
