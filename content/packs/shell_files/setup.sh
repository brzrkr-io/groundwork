#!/usr/bin/env bash
# Builds this exercise's practice folder in sandbox/ (or in the directory
# given as the first argument). Run it again any time to start over: it
# deletes the old sandbox first.
set -euo pipefail

cd "$(dirname "$0")"
dir="${1:-sandbox}"
rm -rf "$dir"
mkdir -p "$dir"
cd "$dir"

# The web app as it runs today.
mkdir -p app/logs
printf 'service: web\nport: 8080\nreplicas: 3\n' > app/config.yaml
printf 'DB_PASSWORD=change-me\n' > app/.env
printf '09:14 GET /health 200\n09:15 GET /orders 500\n' > app/logs/app.log

# The release before it, no longer needed.
mkdir -p old-release
printf 'service: web\nport: 80\nreplicas: 2\n' > old-release/config.yaml
