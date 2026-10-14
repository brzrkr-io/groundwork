#!/usr/bin/env bash
# Task 1: Clone the repository
#
# Clone origin.git into a folder named web. Then, inside web, set the name
# "Sam Rivera" and the email "sam@example.com" for that repository only.
#
# This script runs inside the sandbox folder (where origin.git and alex/
# are): cd sandbox && bash ../1_clone.sh
set -euo pipefail

git clone origin.git web
cd web
git config user.name "Sam Rivera"
git config user.email "sam@example.com"
