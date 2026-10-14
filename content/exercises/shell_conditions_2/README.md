# if and test: fix the backup checks

The nightly backup job has three safety checks, and none of them work. One
never finds the backup, one thinks 10 is less than 5, and one won't even
start. Each is a small mistake that's easy to make in `if` conditions.

**New here:** reading test errors, text versus number comparison.

## Learn

### Recap

- `if COMMAND; then ... else ... fi` runs `then` when the command exits 0.
- `[[ ... ]]` is a command for tests: `-f` (regular file), `-z` (empty),
  `==` (same text), `-gt` and friends (numbers).
- Error messages name the script and the line: read them first.

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
as a test, which fails before the script even starts:

```text
unexpected token `-q', conditional binary operator expected
```

When bash can't make sense of a script, it runs **nothing**: the error comes
before your first line runs.

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

Right now it always takes the `else` branch, after
`[[-f: command not found`.

**Think it through:**
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

Right now 10 prints `ok`, but 6 prints `ALERT`.

**Think it through:**
- Is `>` comparing these as numbers or as text?
- Which operator means "more than" for numbers?

### 3. Fix the log check

**Goal:** make `3_backup_log.sh` print `LOG: complete` when the file `$1`
contains `backup complete`, and `LOG: INCOMPLETE` otherwise.

```bash
bash 3_backup_log.sh testdata/backup_ok.log
# testdata/backup_ok.log: complete

bash 3_backup_log.sh testdata/backup_failed.log
# testdata/backup_failed.log: INCOMPLETE
```

Right now bash refuses to run the script at all.

**Think it through:**
- `grep -q` is a command with an exit status. What does `if` need around
  it?

## Learn more

- [Bash manual: Bash Conditional Expressions](https://www.gnu.org/software/bash/manual/html_node/Bash-Conditional-Expressions.html)
- [Greg's Wiki: Bash Pitfalls](https://mywiki.wooledge.org/BashPitfalls) (numbers 7, 9 and 10 are these three bugs)
- [ShellCheck](https://www.shellcheck.net/): catches all three of these bugs
