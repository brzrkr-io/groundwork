# Grep and pipes

When something breaks at 3 a.m., the answer is somewhere in a log file with
thousands of lines. `grep` finds the lines that matter, and pipes let you
chain small commands until only the answer is left.

**New here:** `grep` and its flags `-i`, `-v`, `-c`, `-E`, `-o`; pipes `|`;
`wc -l`; `cut -d= -f2`; grep's exit status.

## Learn

### grep: keep the matching lines

`grep PATTERN FILE` prints every line of FILE that contains PATTERN, and
nothing else. Say `deploys.txt` holds:

```text
api v1.4.2 ok
worker v2.0.0 failed
api v1.4.3 ok
billing v0.9.1 failed
```

```bash
grep failed deploys.txt
# worker v2.0.0 failed
# billing v0.9.1 failed
```

The match is **case-sensitive** (`Failed` would not match) and can be
anywhere in the line, even inside a longer word.

Put the pattern in quotes when it holds spaces or special characters. To
build a pattern from a variable, use double quotes, as in *Shell variables*:

```bash
grep 'v1.4' deploys.txt
service=worker
grep "$service v" deploys.txt
# worker v2.0.0 failed
```

> **Why is it written like this?** **grep requires** the pattern first and
> the file second. **Convention:** single quotes around fixed patterns, so
> bash leaves every character alone; double quotes only when a `$variable`
> must be swapped in.

### Flags change what grep does

A *flag* is an argument starting with `-` that switches on an option. Flags
go before the pattern:

| Flag | Meaning | Example |
|---|---|---|
| `-i` | ignore case: `failed` also matches `FAILED` | `grep -i failed deploys.txt` |
| `-v` | in*v*ert: print the lines that do **not** match | `grep -v ok deploys.txt` |
| `-c` | print how many lines match, not the lines | `grep -c api deploys.txt` prints `2` |

```bash
grep -v ok deploys.txt
# worker v2.0.0 failed
# billing v0.9.1 failed
```

### Patterns: either-or, and "any of these characters"

The pattern is a *regular expression*: a small language for describing text.
Plain letters match themselves. Two extras are enough for this exercise.

With `-E` (*extended* patterns), `|` means "or":

```bash
grep -E 'worker|billing' deploys.txt
# worker v2.0.0 failed
# billing v0.9.1 failed
```

Square brackets match **one** character from a set, and `*` after something
means "repeated any number of times (even zero)". So `[0-9]*` is a run of
digits, and `[a-z0-9]*` a run of lowercase letters and digits:

```bash
grep 'v[0-9]*\.[0-9]*\.[0-9]*' deploys.txt   # any vX.Y.Z version
```

(`\.` is a real dot. A bare `.` means "any one character".)

> **Why is it written like this?** **grep requires** `-E` for `|`; without
> it, `|` is just a character to look for. **Convention:** always put a
> pattern with `|`, `[`, `*` or `$` in single quotes, or bash grabs those
> characters for itself before grep sees them.

### -o: print only the part that matched

Normally grep prints the whole line. With `-o` it prints only the matching
text, each match on its own line. That's how you pull one value out of a
line:

```bash
echo 'pod=api-7c9d node=node-3' | grep -o 'node=[a-z0-9-]*'
# node=node-3
```

### Pipes: one command's output is the next one's input

The *pipe* `|` connects two commands: everything the left one prints becomes
the input of the right one. A command in a pipe that has no file name reads
that input instead.

```bash
grep failed deploys.txt | grep -v billing
# worker v2.0.0 failed
```

Read it left to right: "take the failed lines, then drop the billing ones".
You can chain as many as you like. Two helpers are handy at the end of a
pipe:

- `wc -l` counts the lines it receives (`grep failed deploys.txt | wc -l`
  prints `2`).
- `cut -d= -f2` splits each line at every `=` (the *delimiter*, `-d`) and
  keeps the second piece (the *field*, `-f`):

```bash
echo 'node=node-3' | cut -d= -f2
# node-3
```

> **Why is it written like this?** **Bash requires** nothing more than the
> `|` between commands. **Convention:** many small commands, each doing one
> thing, joined with pipes. It's the Unix way, and you can build a pipe one
> step at a time: run the first command, look at its output, add the next.
> *Columns with cut and awk* covers `cut` properly.

### grep's exit status

Like every command, grep exits with a status (*Exit codes, && and ||*):

| Status | Meaning |
|---|---|
| `0` | at least one line matched |
| `1` | **nothing matched** |
| `2` | a real error, such as a missing file |

Status 1 is how `if grep -q ...` works, but it has a catch. With
`set -euo pipefail`, a grep that matches nothing **stops your script**, and
`pipefail` means that's true even in the middle of a pipe. This bites hardest
with `grep -c`: when nothing matches, it prints `0` **and** exits with
status 1, so the script fails right after printing a perfectly good answer.

When "no match" is fine, add `|| true`: "if this fails, run `true`", a
command that does nothing and always succeeds.

```bash
grep -c rollback deploys.txt || true
# 0           (and the script carries on)
```

> **Why is it written like this?** **grep decides** that "no match" is
> status 1, so scripts can test for it. **Your choice:** only add `|| true`
> where no match is a normal answer. Elsewhere, letting the script stop is
> the safe default.

## Your task

The payments team had a rough morning and the nightly worker run had
problems too. Each script takes a log file as `$1`. Look at the fixtures
first: `testdata/app.log`, `testdata/worker.log` and `testdata/quiet.log`.
Every line looks like this:

```text
2026-10-13T09:15:02Z ERROR payments msg="card declined" req=7f3a9c
```

Time, level (`INFO`, `WARN`, `ERROR` or `FATAL`), service, message, request
id. Run each script by hand from the exercise folder, for example
`bash 1_error_lines.sh testdata/app.log`.

### 1. Print the error lines

**Goal:** make `1_error_lines.sh` print every line of the log `$1` whose
level is `ERROR`, whole and in the original order.

```bash
bash 1_error_lines.sh testdata/worker.log
# 2026-10-13T22:00:15Z ERROR mailer msg="smtp auth failed" req=c2e8f4
# 2026-10-13T22:00:40Z ERROR worker msg="job failed: reports" req=1d4f7a
# 2026-10-13T22:00:52Z ERROR mailer msg="smtp auth failed" req=f85b21
```

The tests also run it on `testdata/app.log` (5 error lines).

**Think it through:**
- Which word appears on exactly the lines you want?
- Where does the log's file name come from?

### 2. Count the problems

**Goal:** make `2_count_problems.sh` print how many lines of the log `$1`
have the level `ERROR` **or** `FATAL`: just the number.

```bash
bash 2_count_problems.sh testdata/app.log
# 6
bash 2_count_problems.sh testdata/worker.log
# 3
bash 2_count_problems.sh testdata/quiet.log
# 0
```

`testdata/quiet.log` has no problems at all. For it, the script must print
`0` **and** finish with status 0, like the other two.

**Think it through:**
- Which flag counts instead of printing, and which one allows "or"?
- What exit status does grep give when the count is 0, and what does
  `set -e` do with it?

### 3. Hide the health checks

**Goal:** make `3_without_health_checks.sh` print the log `$1` without the
health-check lines: every line that does **not** contain `/healthz`, in the
original order.

```bash
bash 3_without_health_checks.sh testdata/quiet.log
# 2026-10-14T03:00:04Z INFO auth msg="token refreshed" req=7b0f3d
# 2026-10-14T03:00:13Z WARN checkout msg="cache miss rate high" req=58e2c7
# 2026-10-14T03:00:26Z INFO payments msg="charge settled" req=0e9b84
```

The tests also run it on `testdata/worker.log` (8 lines left).

**Think it through:**
- Load balancers call `/healthz` every few seconds. Which flag keeps the
  lines that *don't* match?

### 4. List the failed request ids

**Goal:** make `4_request_ids.sh` print the request ids (the part after
`req=`) of the `ERROR` lines of one service, `$2`, in the log `$1`: one id
per line, in the original order, without `req=`.

```bash
bash 4_request_ids.sh testdata/app.log payments
# 7f3a9c
# 0f1e2d
# d4e5f6
bash 4_request_ids.sh testdata/worker.log mailer
# c2e8f4
# f85b21
```

Only `ERROR` lines count: the payments `FATAL` and `WARN` lines stay out.
Request ids are lowercase letters and digits. The tests always pick a
service that has errors, and also run `testdata/app.log api`.

Build it as a pipe, one step at a time: first the right lines, then only the
`req=...` part, then only the id.

**Think it through:**
- In the log, the level and the service sit next to each other. What
  pattern matches "an ERROR line of this service", built from `$2`?
- Which flag prints only the matching text, and what pattern matches
  `req=` followed by the id?
- Which command splits `req=7f3a9c` at the `=` and keeps the second piece?

## Learn more

- [GNU grep manual](https://www.gnu.org/software/grep/manual/grep.html)
- [The Missing Semester: Data wrangling](https://missing.csail.mit.edu/2020/data-wrangling/)
- [The Missing Semester: Shell tools and scripting](https://missing.csail.mit.edu/2020/shell-tools/)
