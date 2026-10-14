#!/usr/bin/env bash
# Task 2: Make the first commit
#
# Stage all three files and commit them with the message:
# Add deploy script and app config
#
# This script runs inside the sandbox repo: cd sandbox && bash ../2_first_commit.sh
set -euo pipefail

git add .
git commit -m "Add deploy script and app config"
