#!/usr/bin/env bash
# Task 2: Let Git run the check
#
# Let git bisect run "bash check.sh" on each commit it picks. When it
# has found the first bad commit, tag that commit first-bad, then end
# the bisect so you're back on main.
#
# This script runs inside the sandbox repo: cd sandbox && bash ../2_run_bisect.sh
set -euo pipefail

git bisect run bash check.sh
git tag first-bad refs/bisect/bad
git bisect reset
