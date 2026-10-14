#!/usr/bin/env bash
# Builds this exercise's sandbox/ (or the directory given as the first
# argument): a shared repository, origin.git, and Alex's clone of it. Run it
# again any time to start over: it deletes the old sandbox first.
set -euo pipefail

cd "$(dirname "$0")"
dir="${1:-sandbox}"
rm -rf "$dir"
mkdir -p "$dir"
cd "$dir"

# Ignore your personal Git settings, and use fixed names and dates, so the
# repos (and every commit hash in them) are the same on every machine.
export GIT_CONFIG_GLOBAL=/dev/null GIT_CONFIG_NOSYSTEM=1
export GIT_AUTHOR_NAME="Alex Kim" GIT_AUTHOR_EMAIL="alex@example.com"
export GIT_COMMITTER_NAME="Alex Kim" GIT_COMMITTER_EMAIL="alex@example.com"

# commit MESSAGE commits what is staged, one minute after the previous commit.
minute=0
commit() {
	minute=$((minute + 1))
	export GIT_AUTHOR_DATE="@$((1788253200 + minute * 60)) +0000"
	export GIT_COMMITTER_DATE="$GIT_AUTHOR_DATE"
	git commit -q -m "$1"
}

# origin.git plays GitHub or GitLab: a bare repository (history, but no
# files to edit) that the whole team pushes to and fetches from.
git init -q --bare -b main origin.git

# Alex, a teammate, clones it and pushes the first commits. (Git warns that
# the clone is empty, which is expected here.)
git clone -q origin.git alex 2> /dev/null
cd alex
git config user.name "Alex Kim"
git config user.email "alex@example.com"
mkdir -p config
cat > config/app.yaml <<'EOF'
name: web
replicas: 3
port: 8080
EOF
git add config/app.yaml
commit "Add web config"
cat > deploy.sh <<'EOF'
#!/usr/bin/env bash
# Deploys the web service to the environment given as $1.
echo "deploying web to $1"
EOF
git add deploy.sh
commit "Add deploy script"
printf '# web\n\nThe web service. Deploy with: bash deploy.sh ENV\n' > README.md
git add README.md
commit "Add README"
git push -q origin main
