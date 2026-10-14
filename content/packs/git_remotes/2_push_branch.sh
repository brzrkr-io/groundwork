#!/usr/bin/env bash
# Task 2: Push a branch
#
# In web, create and switch to the branch fix/timeout. Add the line
# "request_timeout: 30s" to the end of config/app.yaml, commit it with the
# message "Add a 30s request timeout", and push the branch to origin,
# setting it as the branch's upstream.
#
# This script runs inside the sandbox folder (where origin.git and alex/
# are): cd sandbox && bash ../2_push_branch.sh
set -euo pipefail

# TODO(2): in web, commit the timeout on fix/timeout and push the branch with -u.
