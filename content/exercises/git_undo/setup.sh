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
printf 'host: db-old.internal\nport: 5432\npool: 20\n' > config/db.yaml
cat > docs/runbook.md <<'EOF'
# web runbook

## Deploy

Run ./deploy.sh production.

## Roll back

Run ./deploy.sh production with the previous tag checked out.
EOF
git add .
commit "Add database config and runbook"

cat > deploy.sh <<'EOF'
#!/usr/bin/env bash
# Deploys the web service to the environment given as $1.
echo "deploying web to $1"
EOF
git add deploy.sh
commit "Add deploy script"
git tag v1.0

# This one took production down.
printf 'host: db-new.internal\nport: 5432\npool: 20\n' > config/db.yaml
git add config/db.yaml
commit "Point web at the new database cluster"

# An edit that lost the rollback section by accident.
cat > docs/runbook.md <<'EOF'
# web runbook

## Deploy

Run ./deploy.sh production. Watch the dashboard for 10 minutes.
EOF
git add docs/runbook.md
commit "Update deploy steps in the runbook"

# A local experiment that went nowhere. It was never pushed.
git switch -q -c experiment
printf 'ttl: 1h\n' > config/cache.yaml
git add config/cache.yaml
commit "Try caching everything for an hour"
printf 'ttl: 24h\n' > config/cache.yaml
git add config/cache.yaml
commit "Try caching everything for a day"
git switch -q main
