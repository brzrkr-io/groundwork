#!/usr/bin/env bash
set -euo pipefail

# Task 4: edit the file given as $1 in place: the line that starts with
# LOG_LEVEL= gets the value given as $2. Change nothing else, and print
# nothing. For example:
#   bash 4_set_log_level.sh copy.env debug
# changes the line LOG_LEVEL=info in copy.env to LOG_LEVEL=debug.

sed -i "s/^LOG_LEVEL=.*/LOG_LEVEL=$2/" "$1"
