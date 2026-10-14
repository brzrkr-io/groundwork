#!/usr/bin/env bash
set -euo pipefail

# Task 3: print the total number of lines in all the .conf files under the
# folder given as $1, as one number. File names may contain spaces. For
# example, after bash setup.sh:
#   bash 3_count_conf_lines.sh sandbox
# prints
#   14

find "$1" -type f -name '*.conf' -print0 | xargs -0 cat | wc -l
