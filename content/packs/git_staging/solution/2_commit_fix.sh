#!/usr/bin/env bash
# Task 2: Commit only the timeout fix
#
# Commit what is staged now (the timeout change in config/app.yaml) with
# the message:
# Raise api timeout to 30s
# Leave deploy.sh and notes/ out of it.
#
# This script runs inside the sandbox repo: cd sandbox && bash ../2_commit_fix.sh
set -euo pipefail

git commit -m "Raise api timeout to 30s"
