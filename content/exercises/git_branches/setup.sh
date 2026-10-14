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
cat > deploy.sh <<'EOF'
#!/usr/bin/env bash
# Deploys the web service to the environment given as $1.
echo "deploying web to $1"
EOF
git add deploy.sh
commit "Add deploy script"

cat > config/app.yaml <<'EOF'
name: web
replicas: 2
port: 8080
EOF
git add config/app.yaml
commit "Add app config"

# Written but not committed yet: it belongs on a feature branch.
cat > healthcheck.sh <<'EOF'
#!/usr/bin/env bash
# Exits non-zero if the web service doesn't answer on /healthz.
curl -fsS --max-time 5 http://localhost:8080/healthz
EOF
