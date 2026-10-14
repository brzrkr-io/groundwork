#!/usr/bin/env bash
# Task 1: Revert the debug logging
#
# "Enable debug logging in prod" must not ship, and main is shared.
# Undo it on main with a revert, keeping the default message.
#
# This script runs inside the sandbox repo: cd sandbox && bash ../1_revert_debug_logging.sh
set -euo pipefail

# TODO(1): revert the debug logging commit on main.
