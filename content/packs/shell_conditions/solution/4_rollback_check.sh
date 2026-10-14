#!/usr/bin/env bash
set -euo pipefail

# Task 4: did a deploy roll back?
# Usage:   bash 4_rollback_check.sh LOG
# When LOG contains "ROLLBACK", print "LOG: rolled back".
# Otherwise print "LOG: clean".
# Example: bash 4_rollback_check.sh testdata/deploy_ok.log
#          testdata/deploy_ok.log: clean

# if runs a command and looks at its exit status: no [[ ]] needed.
if grep -q 'ROLLBACK' "$1"; then
	echo "$1: rolled back"
else
	echo "$1: clean"
fi
