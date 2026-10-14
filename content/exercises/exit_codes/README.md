# Exit codes, && and ||

Every command ends with a number, its *exit status*, that says whether it
worked. CI pipelines, deploy gates and Kubernetes all decide what to do next
from that one number. Here you read it, chain commands on it, and set your
own.

**New here:** exit status and `$?`, `grep -q`, `&&` and `||`, `exit N`,
`>&2` (print to stderr), `{ ...; }` groups.

## Learn

### Every command reports success or failure

When a command finishes, it hands bash a number from 0 to 255:

| Exit status | Meaning |
|---|---|
| `0` | success |
| anything else | failure (the number can say which kind) |

`$?` holds the status of the last command:

```bash
ls /etc/hosts
echo $?            # 0: the file exists
ls /etc/nope
echo $?            # 2: ls failed
```

It's like Go's `error`: `nil` means it worked, and anything else means it
didn't. The *The operating system for programmers* guide lists the usual
numbers.

### grep -q: a yes/no question

`grep PATTERN FILE` prints the lines of FILE that contain PATTERN. With
`-q` (*quiet*), it prints nothing and only answers through its exit status:

| grep -q … | Exit status |
|---|---|
| found a matching line | `0` |
| found none | `1` |
| couldn't read the file | `2` |

A `^` at the start of the pattern means "at the start of the line":
`grep -q '^region:' app.conf` matches `region: eu-west-1` but not
`# region: old`.

A script's own exit status is the status of the **last** command it ran.
So a script whose last line is a `grep -q` answers "yes" or "no" for you.

### && and ||

`&&` and `||` run the next command depending on the status of the one
before:

```bash
make build && echo "build ok"      # echo runs only if make succeeded
make build || echo "build failed"  # echo runs only if make failed
```

Put together, `A && B || C` reads "if A, then B, else C":

```bash
ping -c1 -q db-01 > /dev/null && echo "db-01 up" || echo "db-01 down"
```

> **Why is it written like this?** **Bash requires** nothing special:
> `&&` and `||` look only at the status of the command on their left.
> **Convention:** `A && B || C` is fine when B is a simple `echo`, which
> never fails. For anything bigger, use `if` (next exercise), because if B
> fails, C runs too.

`set -e` stops a script when a command fails, **except** a command on the
left of `&&` or `||`. Failure there is a question you asked, not an
accident.

### Your own exit status and error messages

`exit N` ends the script with status N. Error messages go to *stderr*, the
second output stream, so they don't mix with real output. `>&2` sends one
`echo` there:

```bash
echo "config missing" >&2
exit 4
```

To run **two** commands after `||`, group them in braces. Each command in
the group ends with a `;` or a new line:

```bash
test -d /var/backups || {
	echo "no backup directory" >&2
	exit 4
}
```

> **Why is it written like this?** **Bash requires** spaces inside `{ }` and
> a `;` or newline before the `}`. **Convention:** results to stdout, errors
> to stderr, and a non-zero exit when the script failed, so a pipeline can't
> mistake a failure for success. **Your choice:** which non-zero number;
> pick different ones for different failures.

## Your task

The deploy pipeline reads each service's health file, like
`testdata/health_ok.txt`:

```text
service: checkout
version: 1.4.2
status: ok
```

Open all three files in `testdata/` and note which have `status: ok` and
which have a `version:` line.

### 1. Report health with an exit status

**Goal:** make `1_is_healthy.sh` exit with status 0 when the file given as
`$1` has the line `status: ok`, and 1 when it doesn't. It prints nothing.

```bash
bash 1_is_healthy.sh testdata/health_ok.txt; echo $?
# 0

bash 1_is_healthy.sh testdata/health_down.txt; echo $?
# 1
```

**Think it through:**
- Which command answers "does this file contain this text?" without
  printing anything?
- What is a script's exit status when its last command fails?

### 2. Approve or block the deploy

**Goal:** make `2_gate.sh` print `deploy approved` when the file given as
`$1` has the line `status: ok`, and `deploy blocked` otherwise. It exits 0
either way.

```bash
bash 2_gate.sh testdata/health_ok.txt
# deploy approved

bash 2_gate.sh testdata/health_down.txt
# deploy blocked
```

Write it as one line with `&&` and `||`.

**Think it through:**
- What is A, what is B and what is C in `A && B || C`?
- Why doesn't `set -e` stop the script when the search fails?

### 3. Stop with your own exit code

**Goal:** make `3_preflight.sh` exit with status 3 and print
`preflight: no version in FILE` **to stderr** when no line of the file given
as `$1` starts with `version:`. Otherwise it prints `preflight ok: FILE` as
it does already.

```bash
bash 3_preflight.sh testdata/health_ok.txt
# preflight ok: testdata/health_ok.txt

bash 3_preflight.sh testdata/health_noversion.txt; echo $?
# preflight: no version in testdata/health_noversion.txt
# 3
```

FILE is the file name exactly as given in `$1`. On the failing path, print
nothing to stdout.

**Think it through:**
- Which pattern matches `version:` only at the start of a line?
- You need two commands when the check fails. How do you run both after
  `||`?
- Where does the check go: before or after the `echo` that's already there?

## Learn more

- [Bash manual: Exit Status](https://www.gnu.org/software/bash/manual/html_node/Exit-Status.html)
- [Bash manual: Lists of Commands](https://www.gnu.org/software/bash/manual/html_node/Lists.html) (`&&`, `||` and `;`)
- [GNU grep manual](https://www.gnu.org/software/grep/manual/grep.html)
