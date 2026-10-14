#!/usr/bin/env bash
# Task 2: Bring back the deleted branch
#
# The branch feature/alerts was deleted before it was merged. Recreate it,
# pointing at its last commit, "Add disk space alert". Stay on main.
#
# This script runs inside the sandbox repo: cd sandbox && bash ../2_restore_deleted_branch.sh
set -euo pipefail

git branch feature/alerts 'HEAD@{2}'
