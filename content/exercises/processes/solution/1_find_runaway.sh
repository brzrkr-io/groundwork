#!/usr/bin/env bash
set -euo pipefail

# Task 1: print the PID and the parent PID of the process whose name is $1,
# on one line, PID first. For example, after bash setup.sh:
#   bash 1_find_runaway.sh report-export
# prints something like
#   48213 1

pid="$(pgrep -x "$1")"
ps -o pid=,ppid= -p "$pid"
