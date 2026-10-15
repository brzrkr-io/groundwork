# Shell variables

Deploy scripts take a service name, an environment or a tag, and build
commands from them. In bash those values live in *variables*, arrive as
*arguments*, and travel to the programs you start through the
*environment*. This exercise shows how to store them, use them, hand them
on, and keep them safe with quotes.

**New here:** `name=value`, `$name`, `${name}`, script arguments `$1` and
`$2`, defaults with `${2:-...}`, command substitution `$(...)`, what
`set -euo pipefail` does, `export`, `NAME=value command`, `printenv` and
`env`, `unset`, and unset versus empty.

## Learn

### Making a variable

```bash
region=eu-west-1
replicas=3
```

No `var`, no `:=`, and **no spaces around `=`**. Every value is text, even
`3`.

> **Why is it written like this?** **Bash requires** no spaces: `region =
> eu-west-1` means "run a command named `region` with two arguments, `=` and
> `eu-west-1`", and bash says `region: command not found`. **Convention:**
> lowercase names for your own variables; UPPERCASE is for *environment
> variables*, which every program gets from the system, like `HOME` and
> `PATH`.

### Using a variable: `$`

Put `$` in front of the name to get the value. Inside **double** quotes bash
still swaps it in; inside single quotes it doesn't:

```bash
region=eu-west-1
echo "Region: $region"     # Region: eu-west-1
echo 'Region: $region'     # Region: $region
```

When letters, digits or `_` follow the name, wrap the name in braces so bash
knows where it ends:

```bash
echo "${region}_backup"    # eu-west-1_backup
echo "$region_backup"      # looks for a variable named region_backup
```

> **Why is it written like this?** **Convention:** put every `$variable` in
> double quotes. Without quotes, bash splits the value on spaces and expands
> `*`, which breaks on file names with spaces. *Shell variables: fix the
> quoting bugs* shows what goes wrong.

### Script arguments: `$1`, `$2`

When you run `bash deploy.sh api v2`, the script gets the arguments as `$1`
(`api`) and `$2` (`v2`). `$#` is how many there are. Copy them into named
variables at the top, so the rest of the script reads well:

```bash
# bash scale.sh api 5
service="$1"
count="$2"
echo "Scaling $service to $count pods"   # Scaling api to 5 pods
```

### Defaults: `${2:-...}`

`${2:-staging}` means "`$2` if there is one, otherwise `staging`". It works
for any variable:

```bash
# bash deploy.sh api
env="${2:-staging}"
echo "$env"                 # staging
```

### Command substitution: `$(...)`

`$(command)` runs a command and gives you what it printed, without the final
newline. Store it in a variable like any other value (`date +%F` prints
today's date as YYYY-MM-DD):

```bash
today=$(date +%F)
echo "Backup for $today"    # Backup for 2026-10-13
```

An assignment like `today=$(...)` needs no quotes around the right side:
bash never splits it. Quote it anyway if you like; both work.

`wc -l FILE` counts the lines in FILE, and prints the file name after the
number. `< FILE` instead feeds the file to the command as its input, so
`wc` doesn't know the name and prints only the number. With a 3-line
`regions.txt`:

```bash
wc -l regions.txt      # 3 regions.txt
wc -l < regions.txt    # 3
```

### What `set -euo pipefail` does

| Part | Effect |
|---|---|
| `-e` | stop the script as soon as a command fails |
| `-u` | stop if you use a variable that was never set (catches typos) |
| `-o pipefail` | a *pipeline* (commands chained together with `\|`, which *Grep and pipes* covers) fails if any command in it fails, not just the last |

So if you run `bash 1_announce.sh` with no arguments, it stops with
`$1: unbound variable` instead of printing half a message. That's `-u`
protecting you. `${1:-}` (an empty default) is the way to say "this argument
is optional".

### The environment: `export`

A plain variable lives only in this shell (or this script). The programs
it starts don't see it. The *environment* is a separate list of
`NAME=value` pairs, and every program you start gets a **copy** of it.
`export` puts a variable into the environment. `printenv NAME` is a
separate program, so it shows only what was exported:

```bash
TF_WORKSPACE=prod
printenv TF_WORKSPACE      # prints nothing: not exported
export TF_WORKSPACE
printenv TF_WORKSPACE      # prod
export TF_LOG=info         # set and export in one go
```

`env` on its own lists the whole environment. For one command only, put
the setting in front of it: `TF_LOG=debug terraform plan`. The variable
isn't set in your shell afterwards.

Because a child gets a copy, nothing it changes comes back. A script that
runs `export TF_LOG=debug` can't change your terminal's `TF_LOG`.

> **Why is it written like this?** **Bash requires** `export` (or the
> `NAME=value command` form) for a child to see a variable. **Convention:**
> UPPERCASE names for exported settings, lowercase for your own working
> variables, so you can tell them apart at a glance.

### Unset versus empty

`unset NAME` removes a variable completely. `NAME=` keeps it, set to the
empty string. Most tools treat both as "not given", and so does `:-`:

| `TF_LOG` is... | `${TF_LOG:-info}` | `${TF_LOG-info}` | `$TF_LOG` under `set -u` |
|---|---|---|---|
| unset | `info` | `info` | stops: unbound variable |
| empty (`TF_LOG=`) | `info` | (empty) | (empty) |
| `debug` | `debug` | `debug` | `debug` |

So `${TF_LOG:-info}` is the safe way to read a setting the caller may or
may not have given, even under `set -u`.

## Your task

Write four small scripts a deploy pipeline would call. Each one takes its
values as arguments.

### 1. Announce the deploy

**Goal:** make `1_announce.sh` print `Deploying SERVICE to ENV`, where
SERVICE is the first argument and ENV the second.

```bash
bash 1_announce.sh checkout production
# Deploying checkout to production

bash 1_announce.sh api staging
# Deploying api to staging
```

Store `$1` in a variable named `service` and `$2` in one named `env`, then
print the line from them.

**Think it through:**
- Which parts of the line are fixed text, and which come from variables?
- Which quotes let `$service` be swapped in?

### 2. Build the image reference

**Goal:** make `2_image_ref.sh` print `registry.example.com/NAME:TAG`, with
the name from `$1` and the tag from `$2`, or `latest` when there's no `$2`.

```bash
bash 2_image_ref.sh api v1.4.2
# registry.example.com/api:v1.4.2

bash 2_image_ref.sh worker
# registry.example.com/worker:latest
```

**Think it through:**
- How do you write "`$2`, or `latest` if it's missing"?
- `registry.example.com/` and `:` are fixed text. Where do the two
  variables go?

### 3. Count the hosts

**Goal:** make `3_count_hosts.sh` print the file name from `$1`, a colon and
a space, the number of lines in that file, then ` hosts`.

```bash
bash 3_count_hosts.sh testdata/web_hosts.txt
# testdata/web_hosts.txt: 4 hosts

bash 3_count_hosts.sh testdata/db_hosts.txt
# testdata/db_hosts.txt: 2 hosts
```

Look at `testdata/web_hosts.txt` (`cat testdata/web_hosts.txt`): one host
per line. Store the count in a variable named `count` first.

**Think it through:**
- Which command counts lines, and how do you make it print only the number?
- How do you put a command's output into a variable?

### 4. Hand the settings to the check

**Goal:** make `4_run_check.sh` run `bash testdata/check_env.sh` so that
it sees `DEPLOY_ENV` set to `$1`, and `LOG_LEVEL` set to the caller's
`LOG_LEVEL` if that is set and not empty, otherwise `debug`.

`testdata/check_env.sh` prints the two settings from its environment, or
`<unset>` for a missing one (look at it with `cat`):

```bash
bash 4_run_check.sh staging
# DEPLOY_ENV=staging
# LOG_LEVEL=debug

LOG_LEVEL=warn bash 4_run_check.sh production
# DEPLOY_ENV=production
# LOG_LEVEL=warn

LOG_LEVEL= bash 4_run_check.sh production
# DEPLOY_ENV=production
# LOG_LEVEL=debug
```

The tests run all three cases and ignore any `DEPLOY_ENV` or `LOG_LEVEL`
set in your own terminal.

**Think it through:**
- `<unset>` in the output means the child didn't get the variable. What
  makes a variable reach a child?
- Which default form treats an empty `LOG_LEVEL` like a missing one, and
  doesn't trip `set -u` when it's unset?

## Learn more

- [Bash manual: Shell Parameters](https://www.gnu.org/software/bash/manual/html_node/Shell-Parameters.html)
- [Bash manual: Shell Parameter Expansion](https://www.gnu.org/software/bash/manual/html_node/Shell-Parameter-Expansion.html) (the `${...:-default}` family)
- [Bash manual: Command Substitution](https://www.gnu.org/software/bash/manual/html_node/Command-Substitution.html)
- [Bash manual: Environment](https://www.gnu.org/software/bash/manual/html_node/Environment.html)
