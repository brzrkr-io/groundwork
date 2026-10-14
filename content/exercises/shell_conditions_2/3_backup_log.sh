#!/usr/bin/env bash
set -euo pipefail

# Task 3: did the backup finish?
# Usage:   bash 3_backup_log.sh LOG
# When LOG contains "backup complete", print "LOG: complete".
# Otherwise print "LOG: INCOMPLETE".
# Example: bash 3_backup_log.sh testdata/backup_failed.log
#          testdata/backup_failed.log: INCOMPLETE

# TODO(3): bash won't even start this script ("conditional binary operator
# expected"). Fix the condition.
if [[ grep -q 'backup complete' "$1" ]]; then
	echo "$1: complete"
else
	echo "$1: INCOMPLETE"
fi
