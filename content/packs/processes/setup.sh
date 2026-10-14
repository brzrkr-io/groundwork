#!/usr/bin/env bash
set -euo pipefail

# Builds the practice box in sandbox/ and starts two practice processes in
# the background, so you can try your scripts by hand:
#   bash setup.sh
#   bash 1_find_runaway.sh report-export
# Run it again whenever you want a fresh start. It stops the practice
# processes from an earlier run and rebuilds sandbox/.

# Stop the practice processes from an earlier run, if there are any.
pkill -KILL -x report-export || true
pkill -KILL -x cache-warmer || true

rm -rf sandbox
mkdir -p sandbox/reports sandbox/app

# Task 4: a secret anyone can read, and a script nobody can run.
printf 'DB_PASSWORD=hunter2\n' > sandbox/app/secret.env
printf '#!/usr/bin/env bash\necho "deploying api"\n' > sandbox/app/deploy.sh
chmod 644 sandbox/app/secret.env sandbox/app/deploy.sh

# Task 5: the files come without the x bit, so make the tool runnable.
chmod 755 bin/healthcheck

# The practice processes are small bash scripts. Their first line names bash
# by its full path, so ps shows the script's name instead of "bash".
bash_path="$(command -v bash)"

# report-export is stuck: it ignores TERM, works in sandbox/reports and
# keeps export.csv open.
cat > sandbox/report-export <<EOF
#!$bash_path
trap '' TERM
cd "$PWD/sandbox/reports"
exec 3>> export.csv
while true; do sleep 1; done
EOF

# cache-warmer is well behaved: TERM stops it.
cat > sandbox/cache-warmer <<EOF
#!$bash_path
while true; do sleep 1; done
EOF

chmod 755 sandbox/report-export sandbox/cache-warmer
sandbox/report-export > /dev/null 2>&1 &
echo "started report-export, PID $!"
sandbox/cache-warmer > /dev/null 2>&1 &
echo "started cache-warmer, PID $!"

echo "Built sandbox/. Try: bash 1_find_runaway.sh report-export"
