#!/usr/bin/env bash
set -euo pipefail

# Task 2: alert when too many backups failed this week.
# Usage:   bash 2_failure_alert.sh COUNT
# More than 5: print "ALERT: COUNT failed backups".
# Otherwise:   print "ok: COUNT failed backups".
# Example: bash 2_failure_alert.sh 10
#          ALERT: 10 failed backups

count="$1"

# TODO(2): 10 failures print "ok". Fix the comparison.
if [[ "$count" > 5 ]]; then
	echo "ALERT: $count failed backups"
else
	echo "ok: $count failed backups"
fi
