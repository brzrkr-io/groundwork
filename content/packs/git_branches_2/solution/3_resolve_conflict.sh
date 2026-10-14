#!/usr/bin/env bash
# Task 3: Resolve the conflict
#
# Finish the merge. config/app.yaml must keep main's timeout (30s) and
# take the branch's max_connections (200); see the README for the exact
# file. Then mark it resolved and commit with the default message.
#
# This script runs inside the sandbox repo: cd sandbox && bash ../3_resolve_conflict.sh
set -euo pipefail

cat > config/app.yaml <<'EOF'
name: web
replicas: 2
timeout: 30s
max_connections: 200
log_level: info
EOF
git add config/app.yaml
git commit --no-edit
