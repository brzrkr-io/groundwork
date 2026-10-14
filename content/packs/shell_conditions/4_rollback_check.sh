#!/usr/bin/env bash
set -euo pipefail

# Task 4: did a deploy roll back?
# Usage:   bash 4_rollback_check.sh LOG
# When LOG contains "ROLLBACK", print "LOG: rolled back".
# Otherwise print "LOG: clean".
# Example: bash 4_rollback_check.sh testdata/deploy_ok.log
#          testdata/deploy_ok.log: clean

# TODO(4): use grep -q as the if condition.
