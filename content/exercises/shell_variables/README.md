# Shell variables

Deploy scripts take a service name, an environment or a tag, and build
commands from them. In bash those values live in *variables* and arrive as
*arguments*. This exercise shows how to store them, use them, and keep them
safe with quotes.

**New here:** `name=value`, `$name`, `${name}`, script arguments `$1` and
`$2`, defaults with `${2:-...}`, command substitution `$(...)`, and what
`set -euo pipefail` does.

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
> lowercase names for your own variables; UPPERCASE is for environment
> variables like `HOME` and `PATH`.

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
newline. Store it in a variable like any other value:

```bash
today=$(date +%F)
echo "Backup for $today"    # Backup for 2026-10-13
```

`wc -l` counts lines. Given a file name it prints the name too; feed it the
file with `<` and it prints only the number:

```bash
wc -l testdata/regions.txt      # 3 testdata/regions.txt
wc -l < testdata/regions.txt    # 3
```

### What `set -euo pipefail` does

| Part | Effect |
|---|---|
| `-e` | stop the script as soon as a command fails |
| `-u` | stop if you use a variable that was never set (catches typos) |
| `-o pipefail` | a pipeline fails if any command in it fails, not just the last |

So if you run `bash 1_announce.sh` with no arguments, it stops with
`$1: unbound variable` instead of printing half a message. That's `-u`
protecting you. `${1:-}` (an empty default) is the way to say "this argument
is optional".

## Your task

Write three small scripts a deploy pipeline would call. Each one takes its
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

Look at `testdata/web_hosts.txt`: one host per line. Store the count in a
variable named `count` first.

**Think it through:**
- Which command counts lines, and how do you make it print only the number?
- How do you put a command's output into a variable?

## Learn more

- [Bash manual: Shell Parameters](https://www.gnu.org/software/bash/manual/html_node/Shell-Parameters.html)
- [Bash manual: Shell Parameter Expansion](https://www.gnu.org/software/bash/manual/html_node/Shell-Parameter-Expansion.html) (the `${...:-default}` family)
- [Bash manual: Command Substitution](https://www.gnu.org/software/bash/manual/html_node/Command-Substitution.html)
