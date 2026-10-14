#!/usr/bin/env bash
# Task 1: Revert the bad commit
#
# The commit "Point web at the new database cluster" broke production,
# and main is shared. Undo it with a new commit that reverses it,
# keeping Git's default message. Don't rewrite history.
#
# This script runs inside the sandbox repo: cd sandbox && bash ../1_revert_bad_deploy.sh
set -euo pipefail

# TODO(1): revert the database commit on main, without opening an editor.
