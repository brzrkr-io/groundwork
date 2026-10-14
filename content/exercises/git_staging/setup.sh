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
name: api
replicas: 2
timeout: 10s
EOF
cat > deploy.sh <<'EOF'
#!/usr/bin/env bash
# Deploys the api service to the environment given as $1.
set -euo pipefail
echo "deploying api to $1"
EOF
git add .
commit "Add api config and deploy script"

# Work in progress: one real fix, staged; personal notes, staged by mistake;
# a debug line, not staged.
cat > config/app.yaml <<'EOF'
name: api
replicas: 2
timeout: 30s
EOF
git add config/app.yaml
mkdir -p notes
cat > notes/todo.txt <<'EOF'
- ask Priya why the api times out on Mondays
- don't forget the dentist on Thursday
EOF
git add notes/todo.txt
cat > deploy.sh <<'EOF'
#!/usr/bin/env bash
# Deploys the api service to the environment given as $1.
set -euo pipefail
set -x # DEBUG: remove before committing
echo "deploying api to $1"
EOF
