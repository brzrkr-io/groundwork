#!/usr/bin/env bash
set -euo pipefail

# Task 4: in the folder $1, set secret.env to mode 600 and deploy.sh to mode
# 755, whatever they were before. Then write $2 into $1/token, so that the
# new file has mode 600 from the moment it exists. For example, after
# bash setup.sh:
#   bash 4_fix_permissions.sh sandbox/app tok-123
#   ls -l sandbox/app

# TODO(4): fix the two modes, then write the token so it starts as 600.
