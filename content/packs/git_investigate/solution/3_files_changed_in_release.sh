#!/usr/bin/env bash
# Task 3: List the files changed between two releases
#
# Print the paths of the files that differ between tags v1.0.0 and
# v1.1.0, one per line, in Git's order.
#
# This script runs inside the sandbox repo: cd sandbox && bash ../3_files_changed_in_release.sh
set -euo pipefail

git diff --name-only v1.0.0 v1.1.0
