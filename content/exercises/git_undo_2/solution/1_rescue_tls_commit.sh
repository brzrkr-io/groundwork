#!/usr/bin/env bash
# Task 1: Rescue the reset commit
#
# "Add TLS cert rotation" was lost by a git reset --hard. Find it in the
# reflog and create a branch named rescue-tls pointing at it. Stay on main.
#
# This script runs inside the sandbox repo: cd sandbox && bash ../1_rescue_tls_commit.sh
set -euo pipefail

git branch rescue-tls 'HEAD@{5}'
