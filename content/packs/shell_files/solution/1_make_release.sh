#!/usr/bin/env bash
# Task 1: Make the release folder
#
# Create the folder releases/v2/config (and every folder above it), then
# copy app/config.yaml into it.
#
# This script runs inside the sandbox: cd sandbox && bash ../1_make_release.sh
set -euo pipefail

mkdir -p releases/v2/config
cp app/config.yaml releases/v2/config/
