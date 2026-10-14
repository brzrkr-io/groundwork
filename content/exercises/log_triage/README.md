# Practice: log triage report

An alert fired and you're on call. The first thing you want is a quick
summary of the service log: how big it is, how many errors, which services
are failing and which requests were slowest. You build that report as four
small scripts, the last one putting the other three together.

**New here:** nothing; practice combining conditions, loops, functions and
pipelines.

## Learn

### Recap

- Check inputs first: `[[ -z "$x" ]]`, `[[ ! -f "$path" ]]`, a message to
  stderr with `>&2`, and `exit` with a non-zero status (*if and test*).
- `for level in A B C; do ... done` walks a fixed list (*Shell loops*).
- `grep -c` prints `0` but **exits 1** when nothing matches; add `|| true`
  where zero is a fine answer (*Grep and pipes*).
- `awk '$2 == "X" {print $3}'` picks columns from matching lines, and
  succeeds even when no line matches (*Columns with cut and awk*).
- `sort | uniq -c | sort -rn | head -n 3` builds a top-N list
  (*Sort and uniq: top-N reports*).
- With `set -o pipefail`, a pipeline fails if **any** command in it fails.
- A function prints, and `exit` inside a script stops just that script.
  Running `bash other.sh` from your script starts a separate script; if it
  fails, `set -e` stops yours too.

### The log format

Every line of the logs in `testdata/` has the same columns, separated by
single spaces:

```text
2026-10-13T14:00:12Z ERROR payments 5012 upstream timeout
└── 1: time ───────┘ └ 2 ┘ └─ 3 ──┘ └4─┘ └ 5 onward ────┘
```

| Column | Meaning |
|---|---|
| 1 | time (UTC) |
| 2 | level: `ERROR`, `WARN` or `INFO` |
| 3 | service |
| 4 | how long the request took, in milliseconds |
| 5 onward | the message (it has spaces in it) |

## Your task

Open `testdata/incident.log` (30 lines, a bad afternoon),
`testdata/rollout.log` (8 lines, a failed rollout) and `testdata/quiet.log`
(6 lines, a quiet night) and read them before you start.

### 1. Check the input

**Goal:** make `1_check_input.sh` print `LOG: N lines` for the log file
`$1`, after checking that the argument is there and the file exists.

| `$1` | Prints | Exit status |
|---|---|---|
| missing or empty | `usage: 1_check_input.sh LOG` to stderr | 2 |
| not a regular file | `cannot read LOG` to stderr | 1 |
| a regular file | `LOG: N lines` | 0 |

```bash
bash 1_check_input.sh testdata/quiet.log
# testdata/quiet.log: 6 lines

bash 1_check_input.sh testdata/missing.log
# cannot read testdata/missing.log
```

LOG is `$1` exactly as given; N is the number of lines.

**Think it through:**
- Which two checks come first, and in which order?
- How did you count lines and print only the number?

### 2. Count each level

**Goal:** make `2_levels.sh` print `LEVEL: N` for `ERROR`, `WARN` and `INFO`,
in that order, where N is the number of lines of the log `$1` at that level.

```bash
bash 2_levels.sh testdata/incident.log
# ERROR: 7
# WARN: 5
# INFO: 18

bash 2_levels.sh testdata/quiet.log
# ERROR: 0
# WARN: 1
# INFO: 5
```

A level with no lines prints `0`, and the script still exits 0.

**Think it through:**
- A loop over three fixed words, with one count per word.
- How do you match the level column only, not the same word inside a
  message? (Look at what's on each side of it.)
- What does `grep -c` do on `quiet.log` when there are no errors?

### 3. Rank the failing services

**Goal:** make `3_failing_services.sh` print `SERVICE: N` for every service
with `ERROR` lines in the log `$1`, the most errors first.

```bash
bash 3_failing_services.sh testdata/incident.log
# payments: 4
# checkout: 2
# auth: 1

bash 3_failing_services.sh testdata/rollout.log
# worker: 3
# api: 1
```

A log without errors (`quiet.log`) prints nothing and exits 0. The tests
compare exact text, so there are no spaces at the start of the lines.

**Think it through:**
- Which tool gives you column 3 of only the `ERROR` lines, and still
  succeeds when there are none?
- `uniq -c` prints the count **before** the name, with padding. How can you
  swap the two columns and add the `: ` at the end of the pipeline?

### 4. Write the triage report

**Goal:** make `4_triage.sh` print the full report for the log `$1`, using
your three scripts and a function named `section`.

```text
testdata/incident.log: 30 lines

== levels ==
ERROR: 7
WARN: 5
INFO: 18

== failing services ==
payments: 4
checkout: 2
auth: 1

== slowest requests ==
5012 ms payments
5004 ms payments
4870 ms payments
```

In order:

1. The output of `bash 1_check_input.sh` for the log.
2. `section "levels"`, then the output of `bash 2_levels.sh`.
3. `section "failing services"`, then the output of `bash 3_failing_services.sh`.
4. `section "slowest requests"`, then the three slowest lines of the log as
   `MS ms SERVICE` (columns 4 and 3), slowest first.

`section TITLE` prints an empty line, then `== TITLE ==`. When the log can't
be read, the report stops right after the check: it exits 1, with only
`cannot read LOG` on stderr and nothing on stdout. Run it from the exercise
folder, so `bash 1_check_input.sh` finds your script. Tasks 1 to 3 must pass
first.

**Think it through:**
- What does `set -e` do when `bash 1_check_input.sh` exits 1?
- For the slowest requests, which columns do you print, in which order, so
  `sort -rn` sorts by the time?

## Learn more

- [The Missing Semester: Data Wrangling](https://missing.csail.mit.edu/2020/data-wrangling/) (the same kind of log digging, step by step)
- [Google Shell Style Guide](https://google.github.io/styleguide/shellguide.html) (including when a script has outgrown bash)
