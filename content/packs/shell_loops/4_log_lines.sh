#!/usr/bin/env bash
set -euo pipefail

# Task 4: how long is each log?
# Usage:   bash 4_log_lines.sh DIR
# For every file in DIR whose name ends in .log, in alphabetical order,
# print "NAME: N lines", where NAME is the file name without the directory.
# Example: bash 4_log_lines.sh testdata/logs_old
#          billing.log: 3 lines
#          search.log: 2 lines

# TODO(4): loop over DIR/*.log and print each file's name and line count.
