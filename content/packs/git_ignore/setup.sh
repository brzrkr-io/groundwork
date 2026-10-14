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
cat > build.sh <<'EOF'
#!/usr/bin/env bash
# Builds the web binary into bin/.
mkdir -p bin && echo "pretend this is a binary" > bin/web
EOF
cat > config/app.yaml <<'EOF'
name: web
port: 8080
EOF
cat > config/example.env <<'EOF'
# Copy to .env and fill in real values. Never commit the copy.
DB_PASSWORD=change-me
EOF
git add .
commit "Add build script and config"

# Someone committed their local settings, password and all.
echo "DB_PASSWORD=hunter2" > .env
git add .env
commit "Add local settings"

# Things that should never be committed.
mkdir -p bin logs
echo "pretend this is a binary" > bin/web
echo "deploy started" > logs/deploy.log
echo "panic: nil map" > debug.log
echo "DB_PASSWORD=pr0d-s3cret" > config/prod.env
