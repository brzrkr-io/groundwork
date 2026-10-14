# if and test: fix the backup checks

The nightly backup job has three safety checks, and none of them work. One
never finds the backup, one thinks 10 is less than 5, and one won't even
start. Each is a small mistake that's easy to make in `if` conditions.

**New here:** a routine for debugging a condition, text versus number
comparison, syntax errors.

## Learn

### Recap

- `if COMMAND; then ... else ... fi` runs `then` when the command exits 0.
- `[[ ... ]]` is a command for tests: `-f` (regular file), `-z` (empty),
  `==` (same text), `-gt` and friends (numbers).
- `bash -x` prints each command, with values filled in, before it runs
  (*Shell variables: fix the quoting bugs*).

### Debugging a condition in three steps

1. **Run it by hand** with the same arguments. If bash prints an error,
   read it: the script name, the line number, then what went wrong.
2. **No error, wrong answer?** Run it again with `bash -x`. Each line that
   starts with `+` is a command bash ran, so you see the exact test it
   checked, with the values swapped in, and which branch it took.
3. **Compare** that line with what you meant.

Here `cleanup.sh` should print `delete` for logs older than 7 days:

```bash
bash -x cleanup.sh 30
# + set -euo pipefail
# + age=30
# + [[ 30 > 7 ]]
# + echo keep
# keep
```

The trace shows the test that ran, `[[ 30 > 7 ]]`, and that bash took the
`else` branch: it decided 30 is not more than 7. *Text or number?* below
explains why.

### Brackets are words

`[[` and `]]` aren't punctuation like Go's `(` and `)`. Bash reads them as
separate words, so they need spaces around them, just like a command and
its arguments:

| You write | Bash reads |
|---|---|
| `[[ -d "$dir" ]]` | the test `-d` on `$dir` |
| `[[-d "$dir"]]` | a command named `[[-d`, which doesn't exist |

A failing condition isn't fatal under `set -e`, so the script carries on
into the `else` branch, with a `command not found` line on stderr.

### Text or number?

Inside `[[ ]]`, `<` and `>` compare **text**, letter by letter, like words
in a dictionary. `"9" > "10"` is true, because `9` comes after `1`. The
letter operators compare **numbers**:

```bash
[[ "9" > "10" ]]      # true: text order
[[ "9" -gt "10" ]]    # false: 9 is not greater than 10
```

### A command is already a condition

`if` runs any command and checks its exit status. `[[ ]]` is just one such
command. Putting another command *inside* `[[ ]]` makes bash read its words
as a test, and they don't make one:

```text
check.sh: line 4: unexpected token `-q', conditional binary operator expected
```

That's a *syntax error*: bash can't understand the statement, so it doesn't
run it at all. Neither branch runs, and the script stops with status 2.
`bash -x` won't help here; the error message already names the line.

## Your task

Fix the three backup checks. The `testdata/` folder has a fake backup
archive and two backup logs; open them to see what the checks look at.

### 1. Fix the backup file check

**Goal:** make `1_backup_exists.sh` print `backup found: FILE` when `$1` is
a regular file, and otherwise print `no backup at FILE` to stderr and exit 1.

```bash
bash 1_backup_exists.sh testdata/backups/db_2026-10-12.tar.gz
# backup found: testdata/backups/db_2026-10-12.tar.gz

bash 1_backup_exists.sh testdata/backups/db_2026-10-13.tar.gz
# no backup at testdata/backups/db_2026-10-13.tar.gz
```

A directory doesn't count: `testdata/backups` gets `no backup at
testdata/backups`, exit 1. Right now the script always takes the `else`
branch, after `[[-f: command not found`.

**Think it through:**
- Which line does the error name?
- What does bash think the first word of the condition is?

### 2. Fix the number comparison

**Goal:** make `2_failure_alert.sh` print `ALERT: COUNT failed backups` when
the number `$1` is more than 5, and `ok: COUNT failed backups` otherwise.

```bash
bash 2_failure_alert.sh 10
# ALERT: 10 failed backups

bash 2_failure_alert.sh 5
# ok: 5 failed backups
```

Right now 10 prints `ok`, but 6 prints `ALERT`. There's no error message,
so this one is a job for `bash -x`.

**Think it through:**
- Run `bash -x 2_failure_alert.sh 10`. Which test did bash run, and which
  branch did it take?
- Is `>` comparing these as numbers or as text? Which operator means "more
  than" for numbers?

### 3. Fix the log check

**Goal:** make `3_backup_log.sh` print `LOG: complete` when the file `$1`
contains `backup complete`, and `LOG: INCOMPLETE` otherwise.

```bash
bash 3_backup_log.sh testdata/backup_ok.log
# testdata/backup_ok.log: complete

bash 3_backup_log.sh testdata/backup_failed.log
# testdata/backup_failed.log: INCOMPLETE
```

Right now bash stops with a syntax error, status 2, and prints neither
line.

**Think it through:**
- `grep -q` is a command with an exit status. What does `if` need around
  it?

## Learn more

- [Bash manual: Bash Conditional Expressions](https://www.gnu.org/software/bash/manual/html_node/Bash-Conditional-Expressions.html)
- [Greg's Wiki: Bash Pitfalls](https://mywiki.wooledge.org/BashPitfalls) (numbers 7, 9 and 10 are these three bugs)
- [ShellCheck](https://www.shellcheck.net/): catches all three of these bugs
