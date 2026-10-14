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

cat > deploy.sh <<'EOF'
#!/usr/bin/env bash
# Deploys the web service to the environment given as $1.
set -euo pipefail
echo "deploying web to $1"
EOF
git add deploy.sh
commit "Add deploy script"

mkdir -p config
printf 'name: web\nreplicas: 2\n' > config/app.yaml
git add config/app.yaml
commit "Add app config"

# A feature branch starts here...
git switch -q -c feature/retries
mkdir -p lib
cat > lib/retry.sh <<'EOF'
# retry N CMD... runs CMD up to N times until it succeeds.
retry() {
	local n=$1
	shift
	for ((i = 1; i <= n; i++)); do
		"$@" && return 0
		sleep 1
	done
	return 1
}
EOF
git add lib/retry.sh
commit "Add retry helper"

cat > deploy.sh <<'EOF'
#!/usr/bin/env bash
# Deploys the web service to the environment given as $1.
set -euo pipefail
source "$(dirname "$0")/lib/retry.sh"
retry 3 echo "deploying web to $1"
EOF
git add deploy.sh
commit "Retry failed deploys"

# ...while main moves on.
git switch -q main
echo "FROM alpine:3.20" > Dockerfile
git add Dockerfile
commit "Pin base image to alpine 3.20"

printf '# web\n\nRun ./deploy.sh <env> to deploy.\n' > README.md
git add README.md
commit "Add README"
