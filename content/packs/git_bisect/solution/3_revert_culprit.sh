#!/usr/bin/env bash
# Task 3: Revert the culprit
#
# Undo the commit tagged first-bad on main with a revert, keeping the
# default message. Afterwards check.sh passes again.
#
# This script runs inside the sandbox repo: cd sandbox && bash ../3_revert_culprit.sh
set -euo pipefail

git revert --no-edit first-bad
