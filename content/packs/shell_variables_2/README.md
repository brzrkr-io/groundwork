# Shell variables: fix the quoting bugs

A teammate's deploy helpers break in four different ways: one won't start,
one prints the wrong thing, one chokes on a file name, and one trips over a
variable name. All four are classic bash mistakes, and all four are one-line
fixes.

**New here:** how bash reads a line (expand, split, remove quotes), reading
bash's error messages, and watching a script run with `bash -x`.

## Learn

### Recap

- `name=value` sets a variable: no spaces around `=`.
- `"$name"` uses it. Double quotes swap in values; single quotes don't.
- `${name}` marks where a name ends: `"${name}_old"`.
- `set -u` stops the script when it meets a variable that was never set.

### How bash reads a line

Before running a command, bash rewrites the line in steps. Take this line,
where `file` holds `release notes.txt`:

```bash
wc -l $file
```

| Step | What bash does | Result |
|---|---|---|
| 1. Expand | swap `$file` for its value | `wc -l release notes.txt` |
| 2. Split | cut the result into words at spaces | `wc`, `-l`, `release`, `notes.txt` |
| 3. Remove quotes | drop the quote characters | (there were none) |
| 4. Run | start `wc` with the words as arguments | `wc` looks for two files |

`wc` gets **two** file names, `release` and `notes.txt`, and neither exists.
Now with quotes:

```bash
wc -l "$file"
```

The quotes protect the value in step 2, so it stays one word:
`release notes.txt`. Step 3 removes the quotes, and `wc` gets exactly one
file name.

> **Why is it written like this?** **Bash requires** nothing here; it does
> exactly what you typed. **Convention:** quote every `"$variable"`,
> always. Skipping quotes is never *needed* for a plain variable, and it
> breaks the day a value contains a space or a `*`.

### Reading bash's errors

Bash names the script, the line, and what went wrong:

```text
deploy.sh: line 7: replicas: command not found
```

| Message | Usual cause |
|---|---|
| `NAME: command not found` | spaces around `=` (bash ran `NAME` as a command), or a typo |
| `$1: ambiguous redirect` | a `< $var` whose value has spaces: quote it |
| `NAME: unbound variable` | `set -u` met an unset variable: a typo, a missing argument, or a name that ran on into the next letters |

Single quotes give no error at all: they print `$name` as plain text. Watch
the output, not just the errors.

### Watch bash work: `bash -x`

When the message doesn't make it clear, run the script with `bash -x`. Bash
prints each command after step 3, just before running it, on a line starting
with `+`. A word that holds a space is shown in `'...'`. Here `count.sh`
sets `file="release notes.txt"`, then runs both `wc` lines from above:

```bash
bash -x count.sh
# + file='release notes.txt'
# + wc -l release notes.txt          <- two words: the bug
# wc: release: No such file or directory
# ...
# + wc -l 'release notes.txt'        <- one word: correct
# 2 release notes.txt
```

The `+` lines are what the command **really** got, not what you typed.
Single-quote mistakes show up too: `+ echo 'Hello $name'` means bash never
swapped `$name` in.

Your routine: run the script, read the error (script, line, message), check
the table above. If the output is wrong but there's no error, or the error
doesn't say why, run it again with `bash -x`.

## Your task

Fix the four helpers. Run each one by hand first and read what bash says;
add `-x` (`bash -x 1_region.sh`) to watch it. Each `TODO` sits right above
the broken line.

### 1. Fix the region assignment

**Goal:** make `1_region.sh` print `Deploying to REGION`, with the region
from `$1`, or `eu-west-1` when there's no argument.

```bash
bash 1_region.sh
# Deploying to eu-west-1

bash 1_region.sh us-east-2
# Deploying to us-east-2
```

Right now it stops with `region: command not found`.

**Think it through:**
- Bash thinks `region` is a command. Which characters on that line make it
  think so?

### 2. Fix the restart message

**Goal:** make `2_restart.sh` print `Restarting SERVICE on all nodes`, with
the service's name from `$1`.

```bash
bash 2_restart.sh payments
# Restarting payments on all nodes
```

Right now it prints `Restarting $service on all nodes`, word for word.

**Think it through:**
- Which kind of quotes swaps in a variable's value, and which leaves `$`
  alone?

### 3. Fix the file name with a space

**Goal:** make `3_line_count.sh` print the file name from `$1`, a colon and a
space, the number of lines in that file, then ` lines`, even when the name
has a space.

```bash
bash 3_line_count.sh testdata/changelog.txt
# testdata/changelog.txt: 5 lines

bash 3_line_count.sh "testdata/release notes.txt"
# testdata/release notes.txt: 3 lines
```

Notice the quotes on the second command: they make the name one argument
for the script. Inside the script, bash splits it again unless you stop it.

**Think it through:**
- In "How bash reads a line", which step breaks the name in two?
- What protects a value from that step?

### 4. Fix the backup name

**Goal:** make `4_backup_name.sh` print the service from `$1`, an
underscore, the date from `$2`, then `.tar.gz`.

```bash
bash 4_backup_name.sh api 2026-10-13
# api_2026-10-13.tar.gz

bash 4_backup_name.sh billing-db 2026-01-31
# billing-db_2026-01-31.tar.gz
```

Right now it stops with `service_: unbound variable`.

**Think it through:**
- Which variable name does bash read in `$service_$date`? Where does it think
  the name ends?
- How do you show bash exactly where a name ends?

## Learn more

- [Bash manual: Word Splitting](https://www.gnu.org/software/bash/manual/html_node/Word-Splitting.html)
- [Greg's Wiki: Quotes](https://mywiki.wooledge.org/Quotes)
- [ShellCheck](https://www.shellcheck.net/): paste a script and it points at quoting bugs like these
