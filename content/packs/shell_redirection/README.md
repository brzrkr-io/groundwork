# Redirection: >, >>, 2> and here-docs

A cron job's output has to go somewhere: a log file, an error file, or
nowhere at all. Scripts also write config files. *Redirection* sends a
command's output streams into files, and a *here-doc* writes several lines
from inside the script.

**New here:** `>`, `>>`, `2>`, `2>&1`, `/dev/null`, here-docs (`<<EOF` and
`<<'EOF'`) with `cat >`.

## Learn

### Recap

- Every program has three streams: *stdin* (0) for input, *stdout* (1) for
  results and *stderr* (2) for messages, as in *The operating system for
  programmers*.
- `< FILE` feeds a file to stdin, as in *Shell variables*;
  `echo "..." >&2` sends one line to stderr, as in *Exit codes, && and ||*.

### Output into a file: `>` and `>>`

| You write | stdout goes |
|---|---|
| `cmd > FILE` | into FILE, **replacing** what was there (creates it if needed) |
| `cmd >> FILE` | onto the **end** of FILE (creates it if needed) |

```bash
date +%F > backup.log           # backup.log now holds only today's date
echo "dump ok" >> backup.log    # a second line, below the date
```

`>` empties the file **before** the command starts. Run the first line
again and the old lines are gone; run the second again and they pile up.

> **Why is it written like this?** **Bash requires** the redirection to
> name a file; it can go anywhere on the line, but **convention** puts it
> at the end. Bash sets it up and then runs the command, which never knows
> its output went to a file.

### Errors into a file: `2>`

The number in front says which stream to move. `>` alone means `1>`:
stdout.

```bash
bash backup.sh > backup.out 2> backup.err   # results in one file, errors in another
rm -f nginx.pid 2> /dev/null                # errors thrown away
```

`/dev/null` is a special file that swallows everything written to it. Use
it only for output you've decided you never need.

### Both into one file: `2>&1`

`2>&1` means "send stderr wherever stdout points **right now**". Bash reads
redirections left to right, so the order matters:

| You write | stdout | stderr |
|---|---|---|
| `cmd > all.log 2>&1` | `all.log` | `all.log` (stdout already pointed there) |
| `cmd 2>&1 > all.log` | `all.log` | **the terminal** (where stdout pointed when `2>&1` ran) |

Lines land in the file in the order the program printed them.

> **Why is it written like this?** **Bash requires** `&1` to mean "stream
> 1"; `2>1` would create a file named `1`. **Convention:** cron jobs and CI
> steps write `> job.log 2>&1` so the log tells the whole story.

### Here-docs: several lines at once

A *here-doc* gives a command several lines of input, written right in the
script. With `cat > FILE`, those lines become the file:

```bash
port=8443
cat > nginx.conf <<EOF
server {
    listen $port;
}
EOF
```

Everything between the first line and the line holding only `EOF` goes
into `nginx.conf`, with `$port` filled in: `listen 8443;`. Inside a
here-doc you need no quotes around variables.

Quote the marker, `<<'EOF'`, and bash leaves every `$` alone, so the file
gets exactly what you typed: `listen $port;`. Use that for files that hold
`$` themselves, like scripts.

> **Why is it written like this?** **Bash requires** the closing `EOF` alone
> on its line, at the very start, with no spaces. **Convention:** the
> marker is `EOF`, but any word works as long as both ends match.

## Your task

`testdata/deploy.sh` is a pretend deploy: results on stdout, progress and
warnings on stderr. Run it once with `bash testdata/deploy.sh` and see both
streams arrive in your terminal. Your scripts send them into files instead.

The tests run each script in a temporary copy of this folder. When you run
them by hand, they write their files here; delete them whenever you like.

### 1. Start the deploy log

**Goal:** make `1_deploy_log.sh` write the line `deploy started` into
`deploy.log`, replacing whatever was in it, then add the line `build ok` at
the end.

```text
$ bash 1_deploy_log.sh
$ bash 1_deploy_log.sh
$ cat deploy.log
deploy started
build ok
```

However many times it runs, `deploy.log` holds exactly these two lines. The
script prints nothing to the terminal.

**Think it through:**
- Which redirection starts the file over, and which adds to it?
- What happens on the second run if both lines use the same one?

### 2. Split results and warnings

**Goal:** make `2_split_output.sh` run `bash testdata/deploy.sh` with its
stdout in `result.txt` and its stderr in `warnings.txt`.

```text
$ bash 2_split_output.sh
$ cat result.txt
api: ok
worker: ok
$ cat warnings.txt
checking api...
warning: worker restarted once
```

Nothing reaches the terminal.

**Think it through:**
- Which number names stderr?
- Can one command line have two redirections?

### 3. Put everything in one log

**Goal:** make `3_one_log.sh` run `bash testdata/deploy.sh` with both
stdout and stderr in `all.log`, in the order they were printed.

```text
$ bash 3_one_log.sh
$ cat all.log
checking api...
api: ok
warning: worker restarted once
worker: ok
```

Nothing reaches the terminal.

**Think it through:**
- Where must stdout point at the moment you copy it to stderr?
- Which comes first on the line: `> all.log` or `2>&1`?

### 4. Write the config with a here-doc

**Goal:** make `4_write_config.sh` write `app.yaml` with a here-doc, using
the service name from `$1` and the replica count from `$2`.

```text
$ bash 4_write_config.sh api 3
$ cat app.yaml
service: api
replicas: 3
image: registry.example.com/api:stable
```

The starter already stores `$1` in `service` and `$2` in `replicas`. The
service name appears twice: on the first line and inside the image. A
second run replaces the whole file. The script prints nothing.

**Think it through:**
- Which command sends its input into a file, and which redirection
  replaces the file?
- Should the `EOF` marker be quoted, when you want the variables filled in?

## Learn more

- [Bash manual: Redirections](https://www.gnu.org/software/bash/manual/html_node/Redirections.html)
- [Greg's Wiki: Here documents](https://mywiki.wooledge.org/HereDocument)
- [The Linux Command Line: I/O Redirection](https://linuxcommand.org/lc3_lts0070.php)
