#!/usr/bin/env bash
set -euo pipefail

# Builds the practice tree in sandbox/, so you can try your scripts by hand:
#   bash setup.sh
#   bash 1_list_logs.sh sandbox
# Run it again whenever you want a fresh copy (for example after task 4
# deleted some files). It removes the old sandbox/ first.

rm -rf sandbox
mkdir -p sandbox/api sandbox/nginx sandbox/worker/jobs

printf 'listen = 0.0.0.0\nport = 8080\nworkers = 4\n' > sandbox/app.conf

printf '10.0.0.7 GET /health 200\n10.0.0.7 GET /health 200\n10.0.0.9 POST /orders 201\n10.0.0.9 GET /orders/42 200\n' > sandbox/api/access.log
printf '10.0.0.7 GET /health 200\n10.0.0.7 GET /health 200\n10.0.0.9 GET /orders 200\n10.0.0.9 GET /orders 500\n10.0.0.3 GET /health 200\n10.0.0.3 GET /metrics 200\n' > sandbox/api/access.log.1
printf 'db_host = db.internal\ndb_port = 5432\npool_size = 10\ntimeout = 5s\n' > sandbox/api/api.conf
printf 'ERROR db timeout after 5s\nERROR db timeout after 5s\n' > sandbox/api/error.log

printf '192.0.2.10 GET / 200\n192.0.2.10 GET /app.js 200\n192.0.2.44 GET / 200\n192.0.2.44 GET /favicon.ico 404\n192.0.2.51 GET / 200\n' > sandbox/nginx/access.log
printf 'worker_processes 2;\nevents { worker_connections 1024; }\nhttp {\n  server { listen 80; }\n}\n' > sandbox/nginx/nginx.conf

printf 'cron: cleanup started\ncron: removed 12 temp files\ncron: cleanup finished\n' > sandbox/worker/jobs/cron.log
printf 'INFO worker v1 started\nINFO worker v1 stopped\n' > "sandbox/worker/old worker.log"
printf 'queue = jobs\nprefetch = 20\n' > "sandbox/worker/queue settings.conf"
printf 'INFO worker v2 started\nINFO job 81 done\nINFO job 82 done\nWARN job 83 retry\n' > sandbox/worker/worker.log

# The old files: last modified on 1 January 2024. The rest are from today.
touch -t 202401010000 \
	sandbox/api/access.log.1 \
	sandbox/api/api.conf \
	sandbox/api/error.log \
	sandbox/worker/jobs/cron.log \
	"sandbox/worker/old worker.log"

echo "Built sandbox/. Try: bash 1_list_logs.sh sandbox"
