#!/usr/bin/env bash
set -euo pipefail

# Task 3: print "cwd" and the working directory of the process whose PID is
# $1, then one line per open file: the fd number, a space, and where it
# points. For example, after bash setup.sh:
#   bash 3_open_files.sh "$(pgrep -x report-export)"
# prints something like
#   cwd /home/you/groundwork/.../sandbox/reports
#   0 /dev/null
#   1 /dev/null
#   2 /dev/null
#   255 /home/you/groundwork/.../sandbox/report-export
#   3 /home/you/groundwork/.../sandbox/reports/export.csv

pid="$1"
echo "cwd $(readlink "/proc/$pid/cwd")"
for fd in "/proc/$pid/fd/"*; do
	echo "$(basename "$fd") $(readlink "$fd")"
done
