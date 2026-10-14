#!/usr/bin/env bash
# Task 1: Ignore build output and logs
#
# Create .gitignore so Git ignores the bin/ folder and every file ending
# in .log, anywhere. Commit .gitignore with the message:
# Ignore build output and logs
#
# This script runs inside the sandbox repo: cd sandbox && bash ../1_ignore_build_output.sh
set -euo pipefail

printf '%s\n' 'bin/' '*.log' > .gitignore
git add .gitignore
git commit -m "Ignore build output and logs"
