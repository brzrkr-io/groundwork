#!/usr/bin/env bash
# Builds this exercise's practice repo in sandbox/ (or in the directory given
# as the first argument). Run it again any time to start over: it deletes
# the old sandbox first.
set -euo pipefail

cd "$(dirname "$0")"
repo="${1:-sandbox}"
rm -rf "$repo"
mkdir -p "$repo"
cd "$repo"

# Ignore your personal Git settings, and use fixed names and dates, so the
# repo (and every commit hash in it) is the same on every machine.
export GIT_CONFIG_GLOBAL=/dev/null GIT_CONFIG_NOSYSTEM=1
export GIT_AUTHOR_NAME="Sam Rivera" GIT_AUTHOR_EMAIL="sam@example.com"
export GIT_COMMITTER_NAME="Sam Rivera" GIT_COMMITTER_EMAIL="sam@example.com"

# commit MESSAGE commits what is staged, one minute after the previous commit.
minute=0
commit() {
	minute=$((minute + 1))
	export GIT_AUTHOR_DATE="@$((1788253200 + minute * 60)) +0000"
	export GIT_COMMITTER_DATE="$GIT_AUTHOR_DATE"
	git commit -q -m "$1"
}
git init -q -b main
git config user.name "Sam Rivera"
git config user.email "sam@example.com"

mkdir -p config
cat > config/app.yaml <<'EOF'
name: web
replicas: 2
timeout: 10s
max_connections: 100
log_level: info
EOF
git add config/app.yaml
commit "Add web config"

# Two teammates branch off at the same time.
git switch -q -c add-metrics
cat > config/metrics.yaml <<'EOF'
path: /metrics
port: 9090
EOF
git add config/metrics.yaml
commit "Add metrics endpoint"

git switch -q main
git switch -q -c raise-limits
cat > config/app.yaml <<'EOF'
name: web
replicas: 2
timeout: 15s
max_connections: 200
log_level: info
EOF
git add config/app.yaml
commit "Raise connection limit to 200"

# Meanwhile, main moves on and touches the same lines.
git switch -q main
cat > config/app.yaml <<'EOF'
name: web
replicas: 2
timeout: 30s
max_connections: 100
log_level: info
EOF
git add config/app.yaml
commit "Raise timeout to 30s"
