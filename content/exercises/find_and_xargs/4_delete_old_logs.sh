#!/usr/bin/env bash
set -euo pipefail

# Task 4: delete every .log file under the folder given as $1 that was last
# modified more than 7 days ago. Leave every other file alone. For example,
# after bash setup.sh:
#   bash 4_delete_old_logs.sh sandbox
# deletes sandbox/api/error.log, sandbox/worker/jobs/cron.log and
# "sandbox/worker/old worker.log". Run bash setup.sh again to get them back.

# TODO(4): delete the old .log files with find.
