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

cat > check.sh <<'EOF'
#!/usr/bin/env bash
# Exits 0 if config/app.yaml makes web listen on port 8080, and 1 if not.
grep -q '^port: 8080$' config/app.yaml
EOF
mkdir -p config docs alerts
cat > config/app.yaml <<'EOF'
# web service settings
name:     web
port: 8080
replicas: 2
EOF
git add .
commit "Add web config and its check"

# note FILE LINE SUBJECT appends LINE to FILE and commits it.
note() {
	echo "$2" >> "$1"
	git add "$1"
	commit "$3"
}

note docs/runbook.md "# web runbook" "Start the runbook"
note deploy.sh 'echo "deploying web to $1"' "Add deploy script"
git tag v2.0.0
note alerts/web.yaml "alert: WebDown" "Alert when web is down"
note docs/runbook.md "Deploy with ./deploy.sh production." "Document deploys"
note alerts/web.yaml "for: 5m" "Wait 5 minutes before paging"
note deploy.sh 'echo "done"' "Print done after deploying"
note docs/runbook.md "Roll back by deploying the previous tag." "Document rollbacks"
note alerts/web.yaml "severity: page" "Page on web outages"

# The culprit: a tidy-up with a typo hidden in it.
cat > config/app.yaml <<'EOF'
# web service settings
name: web
port: 8008
replicas: 2
EOF
git add config/app.yaml
commit "Tidy up config formatting"

note docs/runbook.md "Escalate to #web-oncall." "Add escalation path"
note deploy.sh 'echo "check the dashboard"' "Remind deployers to check the dashboard"
note alerts/web.yaml "team: web" "Route web alerts to the web team"
note docs/runbook.md "Dashboards: grafana/web." "Link the dashboard"
note deploy.sh '# end of deploy' "Mark the end of the deploy script"
