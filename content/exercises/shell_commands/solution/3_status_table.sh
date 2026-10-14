#!/usr/bin/env bash
set -euo pipefail

# Task 3: print this table with printf. The first column is 10 characters
# wide, then one space, then the second column:
#   SERVICE    STATUS
#   api        running
#   worker     stopped
#   payments   running

printf '%-10s %s\n' SERVICE STATUS
printf '%-10s %s\n' api running
printf '%-10s %s\n' worker stopped
printf '%-10s %s\n' payments running
