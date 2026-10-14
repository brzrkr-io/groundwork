#!/usr/bin/env bash
# Task 1: Create the repository
#
# Turn this folder into a Git repository whose first branch is main.
# Then set the name "Sam Rivera" and the email "sam@example.com" for
# this repository only (no --global).
#
# This script runs inside the sandbox repo: cd sandbox && bash ../1_init_repo.sh
set -euo pipefail

git init -b main
git config user.name "Sam Rivera"
git config user.email "sam@example.com"
