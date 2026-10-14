#!/usr/bin/env bash
set -euo pipefail

# Task 2: grade a disk usage percentage.
# Usage:   bash 2_disk_level.sh PERCENT
# 90 or more: print "critical: PERCENT%"
# 75 or more: print "warning: PERCENT%"
# otherwise:  print "ok: PERCENT%"
# Example: bash 2_disk_level.sh 82
#          warning: 82%

percent="$1"

# TODO(2): write the if / elif / else statement.
