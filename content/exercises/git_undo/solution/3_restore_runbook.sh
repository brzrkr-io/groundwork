#!/usr/bin/env bash
# Task 3: Restore a file from an old commit
#
# docs/runbook.md lost its "Roll back" section. Bring back the whole file
# as it was at tag v1.0 and commit it with the message:
# Restore the rollback section of the runbook
#
# This script runs inside the sandbox repo: cd sandbox && bash ../3_restore_runbook.sh
set -euo pipefail

git restore --source=v1.0 docs/runbook.md
git add docs/runbook.md
git commit -m "Restore the rollback section of the runbook"
