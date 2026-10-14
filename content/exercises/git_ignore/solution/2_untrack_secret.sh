#!/usr/bin/env bash
# Task 2: Stop tracking the committed secret
#
# .env (a real password) was committed. Stop tracking it but keep the
# file on disk, add .env to .gitignore, and commit both with the message:
# Stop tracking .env
#
# This script runs inside the sandbox repo: cd sandbox && bash ../2_untrack_secret.sh
set -euo pipefail

git rm --cached .env
echo '.env' >> .gitignore
git add .gitignore
git commit -m "Stop tracking .env"
