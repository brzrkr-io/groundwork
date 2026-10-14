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

mkdir -p config docs
echo "1.3.0" > VERSION
cat > config/app.yaml <<'EOF'
name: web
log_level: info
replicas: 3
port: 8080
EOF
git add .
commit "Add web config"

printf '#!/usr/bin/env bash\necho "deploying web $(cat VERSION) to $1"\n' > deploy.sh
git add deploy.sh
commit "Add deploy script"
git tag v1.3.0

# The rate-limit feature, reviewed and ready to merge.
git switch -q -c feature/rate-limit
printf 'requests_per_second: 50\nburst: 100\n' > config/ratelimit.yaml
git add config/ratelimit.yaml
commit "Add rate limiter config"
printf '# Rate limits\n\nweb allows 50 requests per second per client.\n' > docs/rate-limits.md
git add docs/rate-limits.md
commit "Document rate limits"

# Meanwhile on main: one good change, one that must not ship.
git switch -q main
cat > config/app.yaml <<'EOF'
name: web
log_level: debug
replicas: 3
port: 8080
EOF
git add config/app.yaml
commit "Enable debug logging in prod"

cat > config/app.yaml <<'EOF'
name: web
log_level: debug
replicas: 3
port: 8080
request_timeout: 15s
EOF
git add config/app.yaml
commit "Add request timeout setting"
