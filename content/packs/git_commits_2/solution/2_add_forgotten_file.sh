#!/usr/bin/env bash
# Task 2: Add the forgotten file
#
# config/health.yaml belongs in the newest commit, but was never added.
# Put it into that commit, keeping the message.
#
# This script runs inside the sandbox repo: cd sandbox && bash ../2_add_forgotten_file.sh
set -euo pipefail

git add config/health.yaml
git commit --amend --no-edit
