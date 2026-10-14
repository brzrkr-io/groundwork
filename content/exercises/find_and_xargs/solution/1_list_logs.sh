#!/usr/bin/env bash
set -euo pipefail

# Task 1: print the path of every .log file under the folder given as $1,
# sorted, one per line. Files like access.log.1 don't count. For example,
# after bash setup.sh:
#   bash 1_list_logs.sh sandbox
# prints
#   sandbox/api/access.log
#   sandbox/api/error.log
#   sandbox/nginx/access.log
#   sandbox/worker/jobs/cron.log
#   sandbox/worker/old worker.log
#   sandbox/worker/worker.log

find "$1" -type f -name '*.log' | sort
