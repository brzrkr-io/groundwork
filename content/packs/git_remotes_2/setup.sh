#!/usr/bin/env bash
# Builds this exercise's sandbox/ (or the directory given as the first
# argument): the shared repository origin.git, Alex's clone alex/, and your
# clone web/. Run it again any time to start over: it deletes the old
# sandbox first.
set -euo pipefail

cd "$(dirname "$0")"
dir="${1:-sandbox}"
rm -rf "$dir"
mkdir -p "$dir"
cd "$dir"

# Ignore your personal Git settings, and use fixed names and dates, so the
# repos (and every commit hash in them) are the same on every machine.
export GIT_CONFIG_GLOBAL=/dev/null GIT_CONFIG_NOSYSTEM=1

# author NAME EMAIL makes the following commits by that person.
author() {
	export GIT_AUTHOR_NAME="$1" GIT_AUTHOR_EMAIL="$2"
	export GIT_COMMITTER_NAME="$1" GIT_COMMITTER_EMAIL="$2"
}

# commit MESSAGE commits what is staged, one minute after the previous commit.
minute=0
commit() {
	minute=$((minute + 1))
	export GIT_AUTHOR_DATE="@$((1788253200 + minute * 60)) +0000"
	export GIT_COMMITTER_DATE="$GIT_AUTHOR_DATE"
	git commit -q -m "$1"
}

# origin.git plays GitHub or GitLab. Alex pushes the first commits. (Git
# warns that the first clone is empty, which is expected here.)
git init -q --bare -b main origin.git
author "Alex Kim" "alex@example.com"
git clone -q origin.git alex 2> /dev/null
cd alex
git config user.name "Alex Kim"
git config user.email "alex@example.com"
mkdir -p config
printf 'name: web\nreplicas: 3\nport: 8080\n' > config/app.yaml
git add config/app.yaml
commit "Add web config"
cat > deploy.sh <<'EOF'
#!/usr/bin/env bash
# Deploys the web service to the environment given as $1.
echo "deploying web to $1"
EOF
git add deploy.sh
commit "Add deploy script"
git push -q origin main
cd ..

# You clone it and start a feature branch, which you push for review.
author "Sam Rivera" "sam@example.com"
git clone -q origin.git web
cd web
git config user.name "Sam Rivera"
git config user.email "sam@example.com"
git switch -q -c feature/retries
mkdir -p lib
cat > lib/retry.sh <<'EOF'
# retry N CMD... runs CMD up to N times, until it succeeds.
retry() {
	local n="$1"
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
source lib/retry.sh
retry 3 echo "deploying web to $1"
EOF
git add deploy.sh
commit "Retry failed deploys"
git push -q -u origin feature/retries 2> /dev/null
git switch -q main
cd ..

# Meanwhile Alex pushes a change to main.
author "Alex Kim" "alex@example.com"
cd alex
printf 'FROM alpine:3.20\nCOPY deploy.sh /deploy.sh\n' > Dockerfile
git add Dockerfile
commit "Pin base image to alpine 3.20"
git push -q origin main
cd ..

# And you commit on main without fetching first. Your git push will be
# rejected.
author "Sam Rivera" "sam@example.com"
cd web
printf 'name: web\nreplicas: 4\nport: 8080\n' > config/app.yaml
git add config/app.yaml
commit "Raise web replicas to 4"
