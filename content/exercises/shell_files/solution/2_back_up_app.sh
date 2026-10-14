#!/usr/bin/env bash
# Task 2: Back up the app folder
#
# Copy the whole app folder, hidden files included, so that the copy is
# backup/app.
#
# This script runs inside the sandbox: cd sandbox && bash ../2_back_up_app.sh
set -euo pipefail

# cp -r copies the folder itself, hidden files and all; app/* would skip .env.
mkdir -p backup
cp -r app backup/
