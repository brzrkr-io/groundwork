#!/usr/bin/env bash
# Task 3: Ignore every env file except the example
#
# Add rules to .gitignore so every file ending in .env is ignored, except
# example.env, and commit with the message:
# Ignore env files except the example
#
# This script runs inside the sandbox repo: cd sandbox && bash ../3_ignore_env_files.sh
set -euo pipefail

printf '%s\n' '*.env' '!example.env' >> .gitignore
git add .gitignore
git commit -m "Ignore env files except the example"
