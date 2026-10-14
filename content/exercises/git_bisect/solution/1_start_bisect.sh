#!/usr/bin/env bash
# Task 1: Start bisecting
#
# check.sh passed at tag v2.0.0 and fails on main today. Start a bisect,
# telling Git that main is bad and v2.0.0 is good.
#
# This script runs inside the sandbox repo: cd sandbox && bash ../1_start_bisect.sh
set -euo pipefail

git bisect start main v2.0.0
