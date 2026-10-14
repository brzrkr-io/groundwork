#!/usr/bin/env bash
set -euo pipefail

# Task 2: print the path of every .log file under the folder given as $1
# that was last modified more than 7 whole days ago, sorted, one per line.
# For example, after bash setup.sh:
#   bash 2_list_old_logs.sh sandbox
# prints
#   sandbox/api/error.log
#   sandbox/worker/jobs/cron.log
#   sandbox/worker/old worker.log

find "$1" -type f -name '*.log' -mtime +7 | sort
