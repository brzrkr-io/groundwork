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

# author NAME EMAIL sets who wrote the next commits.
author() {
	export GIT_AUTHOR_NAME="$1" GIT_AUTHOR_EMAIL="$2"
}

# app_config REPLICAS TIMEOUT MAX_CONNECTIONS writes config/app.yaml.
app_config() {
	mkdir -p config
	printf 'name: web\nreplicas: %s\nport: 8080\ntimeout: %s\nmax_connections: %s\n' "$1" "$2" "$3" > config/app.yaml
}

author "Sam Rivera" sam@example.com
app_config 2 10s 100
git add config/app.yaml
commit "Add web config"
git tag v1.0.0

author "Priya Shah" priya@example.com
printf '#!/usr/bin/env bash\necho "deploying web to $1"\n' > deploy.sh
git add deploy.sh
commit "Add deploy script"

author "Alex Kim" alex@example.com
app_config 2 10s 500
git add config/app.yaml
commit "Raise connection limit for the sale"

author "Priya Shah" priya@example.com
mkdir -p docs
printf '# web runbook\n\nDeploy with ./deploy.sh production.\n' > docs/runbook.md
git add docs/runbook.md
commit "Add runbook"

author "Sam Rivera" sam@example.com
app_config 3 10s 500
git add config/app.yaml
commit "Scale web to 3 replicas"
git tag v1.1.0

author "Alex Kim" alex@example.com
app_config 3 5s 500
git add config/app.yaml
commit "Lower timeout to 5s"

author "Priya Shah" priya@example.com
app_config 3 5s 100
git add config/app.yaml
commit "Restore normal connection limit"

author "Jordan Lee" jordan@example.com
printf '# web runbook\n\nDeploy with ./deploy.sh production.\nPage #web-oncall if it fails.\n' > docs/runbook.md
git add docs/runbook.md
commit "Add escalation to the runbook"
