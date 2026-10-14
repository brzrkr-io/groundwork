#!/usr/bin/env bash
# Task 2: Commit on the branch
#
# On feature/healthcheck, commit healthcheck.sh with the message:
# Add health check script
#
# This script runs inside the sandbox repo: cd sandbox && bash ../2_commit_on_branch.sh
set -euo pipefail

git add healthcheck.sh
git commit -m "Add health check script"
