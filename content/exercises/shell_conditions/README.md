# if and test

Before a deploy touches anything, a good script checks: is the config there?
Is the disk full? Is this a real environment? Bash's `if` makes those
decisions, and `[[ ... ]]` asks the questions.

**New here:** `if`/`elif`/`else`/`fi`, `[[ ... ]]` tests, file tests (`-f`,
`-d`), string tests (`-z`, `==`), number tests (`-ge`, `-gt`, ...), `&&` and
`||` inside a test, and `if` with any command.

## Learn

### if runs a command and checks its status

```bash
if grep -q 'maintenance: on' site.conf; then
	echo "skipping deploy"
else
	echo "deploying"
fi
```

`if` runs the command after it. Exit status 0 means "true" and the `then`
part runs; anything else runs the `else` part. `elif` adds more branches,
and `fi` (`if` backwards) ends the statement. There are no braces.

> **Why is it written like this?** **Bash requires** `then` after the
> condition (on the next line, or after a `;` on the same line) and `fi` at
> the end. **Convention:** `; then` on the `if` line, and the body indented
> by one tab or a few spaces. Bash itself ignores indentation.

> **Watch out:** in *Exit codes, && and ||* you read `$?` on the line after
> a command. Don't do that with `set -e`: a `grep -q` that finds nothing
> stops the script right there, and the line that checks `$?` never runs.
> Put the command **in** the `if`. A condition is allowed to fail.

### [[ ... ]] asks a question

Most conditions are `[[ ... ]]`, a command that only answers 0 (true) or 1
(false). Spaces inside the brackets are required, around every part:

```bash
if [[ -d /var/backups ]]; then
	echo "backup directory exists"
fi
```

| Test | True when |
|---|---|
| `-f PATH` | PATH is a regular file (not a directory) |
| `-d PATH` | PATH is a directory |
| `-e PATH` | PATH exists at all |
| `-z "$s"` | the string is empty |
| `-n "$s"` | the string is not empty |
| `"$a" == "$b"` | the strings are equal |
| `"$a" != "$b"` | the strings differ |

`&&` and `||` work inside: `[[ "$tier" == "web" || "$tier" == "api" ]]`.
And `!` turns a test around: `[[ ! -f "$path" ]]`.

You'll also meet the older form, `[ ... ]`, a command also named `test`
(`[ -d DIR ]` is the same as `test -d DIR`). It works on every shell, but it
splits unquoted variables and has no `&&` inside. `[[ ]]` is the bash way,
and the one used here; the tests accept either.

### Comparing numbers

`==` compares **text**. Numbers have their own operators:

| Operator | Means |
|---|---|
| `-eq` / `-ne` | equal / not equal |
| `-gt` / `-ge` | greater than / greater or equal |
| `-lt` / `-le` | less than / less or equal |

```bash
if [[ "$replicas" -lt 2 ]]; then
	echo "warning: no redundancy"
fi
```

> **Why is it written like this?** **Bash requires** the letter operators
> for numbers: inside `[[ ]]`, `<` and `>` compare text alphabetically, so
> `"10" < "9"` is true. **Convention:** quote variables inside `[[ ]]` too.
> It's not always required there, but it's one habit for every place.

### Order the branches

Bash tries `if` and each `elif` from the top and runs the **first** branch
whose test is true. With thresholds, check the biggest first:

```bash
if [[ "$load" -ge 8 ]]; then
	echo "overloaded"
elif [[ "$load" -ge 4 ]]; then
	echo "busy"
else
	echo "fine"
fi
```

## Your task

Write four checks a deploy script runs first. Look in `testdata/` for the
config and the two deploy logs.

### 1. Check the config file

**Goal:** make `1_check_config.sh` print `found PATH` when `$1` is a regular
file, and otherwise print `missing PATH` to stderr and exit 1.

```bash
bash 1_check_config.sh testdata/app.conf
# found testdata/app.conf

bash 1_check_config.sh testdata/missing.conf; echo $?
# missing testdata/missing.conf
# 1
```

PATH is `$1` exactly as given. A directory doesn't count:
`bash 1_check_config.sh testdata` prints `missing testdata` and exits 1.

**Think it through:**
- Which test is true only for a regular file?
- On the missing path, which two things happen, and where does the message
  go?

### 2. Grade disk usage

**Goal:** make `2_disk_level.sh` print `critical: N%` when the number `$1` is
90 or more, `warning: N%` when it's 75 or more, and `ok: N%` otherwise.

```bash
bash 2_disk_level.sh 95
# critical: 95%

bash 2_disk_level.sh 75
# warning: 75%

bash 2_disk_level.sh 8
# ok: 8%
```

The starter stores `$1` in `percent`.

**Think it through:**
- "90 or more" is which operator?
- Which branch must come first, so 95 isn't graded as a warning?

### 3. Validate the environment

**Goal:** make `3_check_env.sh` accept only `production` and `staging`.

| `$1` | Prints | Exit status |
|---|---|---|
| missing or empty | `usage: 3_check_env.sh ENV` to stderr | 2 |
| `production` or `staging` | `ENV is a valid environment` | 0 |
| anything else | `unknown environment: ENV` to stderr | 2 |

```bash
bash 3_check_env.sh staging
# staging is a valid environment

bash 3_check_env.sh prod
# unknown environment: prod
```

The starter stores `$1` in `env`, with an empty default so a missing
argument doesn't trip `set -u`.

**Think it through:**
- Which test is true for an empty string? Do that check first, and `exit`.
- How do you say "production or staging" in one test?

### 4. Look for a rollback

**Goal:** make `4_rollback_check.sh` print `LOG: rolled back` when the log
file `$1` contains `ROLLBACK`, and `LOG: clean` otherwise.

```bash
bash 4_rollback_check.sh testdata/deploy_ok.log
# testdata/deploy_ok.log: clean

bash 4_rollback_check.sh testdata/deploy_rollback.log
# testdata/deploy_rollback.log: rolled back
```

**Think it through:**
- `if` runs any command. Which command answers "is this text in the file?"
- Does that command need `[[ ]]` around it?

## Learn more

- [Bash manual: Conditional Constructs](https://www.gnu.org/software/bash/manual/html_node/Conditional-Constructs.html) (`if` and `[[ ]]`)
- [Bash manual: Bash Conditional Expressions](https://www.gnu.org/software/bash/manual/html_node/Bash-Conditional-Expressions.html) (every `-f`, `-z`, `-gt`, ...)
- [Greg's Wiki: BashFAQ 031](https://mywiki.wooledge.org/BashFAQ/031) (`[` versus `[[`)
