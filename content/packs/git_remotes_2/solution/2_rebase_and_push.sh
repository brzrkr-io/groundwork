#!/usr/bin/env bash
# Task 2: Rebase your commit and push main
#
# Put your unpushed commit on main on top of origin/main by pulling with
# rebase, then push main. No merge commits.
#
# This script runs inside your clone: cd sandbox/web && bash ../../2_rebase_and_push.sh
set -euo pipefail

# --rebase replays your unpushed commit on top of what Alex pushed.
git pull --rebase
git push
