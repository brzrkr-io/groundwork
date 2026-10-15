# Shell functions

Long deploy scripts repeat themselves: log this, check that, build a name.
Bash *functions* give those steps a name, just like Go functions. They work a
little differently, though: arguments arrive as `$1`, `$2`, and a function
answers either with an exit status or by printing.

**New here:** `name() { ... }`, function arguments, `local`, a function's
exit status, capturing a function's output with `$(...)`, `command -v`,
`! command`.

## Learn

### Defining and calling a function

```bash
banner() {
	echo "=== $1 ==="
}

banner "release 1.4"     # === release 1.4 ===
banner "rollback"        # === rollback ===
```

You call a function like any command: its name, then its arguments, with no
parentheses. Inside, the arguments are `$1`, `$2`, ... and `$#`, just as in
a script. The function must be defined **above** the first line that calls
it, because bash runs the file from the top.

> **Why is it written like this?** **Bash requires** the `()` after the name
> (it never holds parameters) and a space or new line after `{`.
> **Convention:** `snake_case` names, one function per job, and the
> functions at the top of the script, before the main part.

### local variables

Variables in bash are **global** by default: a function that sets `name`
overwrites the caller's `name`. `local` keeps a variable inside the
function:

```bash
greet() {
	local who="$1"
	echo "hello, $who"
}
```

It's what Go does for you automatically. Use `local` for every variable a
function creates.

> **Watch out:** `local out=$(some_command)` hides a failure. The line's
> status is `local`'s, which is always 0, so `set -e` never sees
> `some_command` fail. Declare first, then assign: `local out` on one line,
> `out=$(some_command)` on the next.

### Answering yes or no: the exit status

A function's exit status is the status of the last command it ran. So a
function whose last line is a test answers yes or no, and works directly in
an `if`:

```bash
is_weekend() {
	local day="$1"
	[[ "$day" == "Sat" || "$day" == "Sun" ]]
}

if is_weekend "Sat"; then
	echo "no deploys today"
fi
```

`return N` stops a function early with status N, like `exit` does for the
whole script. And `exit` inside a function still stops **the whole script**.

### Answering with a value: print it

A bash function can't return text the way a Go function returns a string.
Instead it **prints** the value, and the caller captures the output with
`$(...)`, as you did with commands in *Shell variables*:

```bash
backup_path() {
	local service="$1"
	echo "/var/backups/$service.tar.gz"
}

path=$(backup_path "billing")
echo "writing $path"       # writing /var/backups/billing.tar.gz
```

`$(...)` runs the function in a copy of the shell: whatever it sets there,
even without `local`, vanishes when it's done. Only what it printed comes
back.

| The caller wants | The function | The caller writes |
|---|---|---|
| yes or no | ends with a test, or `return N` | `if check_x ...; then` |
| a value | prints it with `echo` | `x=$(get_x ...)` |

Don't use `return` for values: it carries only a status from 0 to 255,
never text.

### Is a tool installed?

`command -v NAME` prints where a command lives and exits 0, or prints
nothing and exits 1 when there's no such command. `> /dev/null` throws away
what it prints (as in *Redirection: >, >>, 2> and here-docs*), and `!`
flips the exit status:

```bash
if ! command -v terraform > /dev/null; then
	echo "please install terraform" >&2
fi
```

## Your task

Each script already has its main part, at the bottom. It calls a function
that doesn't exist yet, so right now bash says `command not found`. Write
each function where the `TODO` is.

### 1. Write the log helpers

**Goal:** in `1_log.sh`, define `log_info`, which prints `[INFO] MESSAGE` to
stdout, and `log_error`, which prints `[ERROR] MESSAGE` to stderr. MESSAGE is
the function's first argument.

```bash
bash 1_log.sh api
# [INFO] deploy started: api
# [INFO] pulling image
# [ERROR] image pull slow, retrying
# [INFO] deploy finished: api
```

The tests check both streams: three `[INFO]` lines on stdout, one `[ERROR]`
line on stderr.

**Think it through:**
- Inside a function, which variable holds its first argument?
- How did *Exit codes, && and ||* send an `echo` to stderr?

### 2. Check the ports

**Goal:** in `2_ports.sh`, define `valid_port`, which succeeds (status 0)
when its argument is a number from 1 to 65535, and fails (status 1)
otherwise.

```bash
bash 2_ports.sh 8080 0 65536
# 8080 ok
# 0 invalid
# 65536 invalid
```

Every argument is a whole number. `valid_port` prints nothing: the loop
below it uses its exit status.

**Think it through:**
- Which test is true for "at least 1 **and** at most 65535"?
- What is the function's exit status when that test is its last command?

### 3. Build the image references

**Goal:** in `3_release.sh`, define `image_ref`, which prints
`registry.example.com/NAME:TAG`, where NAME is its first argument and TAG its
second. Use `local` variables.

```bash
bash 3_release.sh testdata/release.txt
# deploy api -> registry.example.com/api:1.4.2
# deploy worker -> registry.example.com/worker:1.4.2
# deploy billing -> registry.example.com/billing:3.0.0
```

The loop reads `testdata/release.txt` (`NAME TAG` per line) and captures
what `image_ref` prints.

**Think it through:**
- How does a bash function hand a value back to its caller?
- Which two `local` variables does the function need, and which of `$1`
  and `$2` fills each?

### 4. Require the tools

**Goal:** in `4_require.sh`, define `require`, which does nothing when the
command named by its argument exists. When it doesn't, `require` prints
`missing tool: TOOL` to stderr and stops the whole script with exit
status 1.

```bash
bash 4_require.sh grep sed
# all tools found

bash 4_require.sh grep kubectl-groundwork sed; echo $?
# missing tool: kubectl-groundwork
# 1
```

**Think it through:**
- Which command answers "is this tool installed?", and how do you hide what
  it prints?
- `return 1` or `exit 1`: which one stops the **whole** script by itself?
  (`return 1` passes here only because `set -e` then stops the script;
  called inside an `if`, the script would carry on.)

## Learn more

- [Bash manual: Shell Functions](https://www.gnu.org/software/bash/manual/html_node/Shell-Functions.html)
- [Google Shell Style Guide: Functions](https://google.github.io/styleguide/shellguide.html#function-names) (naming, `local`, where functions go)
