#!/usr/bin/env bash
# Task 2: Start the conflicting merge
#
# Merge raise-limits into main. It conflicts in config/app.yaml, so git
# merge stops and exits with status 1. Make the script carry on anyway:
# the merge stays in progress for task 3.
#
# This script runs inside the sandbox repo: cd sandbox && bash ../2_start_conflicting_merge.sh
set -euo pipefail

# TODO(2): start merging raise-limits; don't let the expected failure stop the script.
