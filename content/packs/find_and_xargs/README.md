# Find and xargs: clean up old logs

A disk fills up at 3 a.m., and it's always logs. `find` lists the files that
match your rules (name, age, size) anywhere under a folder, and `xargs` runs
a command on all of them. Together they're the classic cleanup job.

**New here:** `find` with `-type`, `-name`, `-mtime`, `-size` and `-delete`;
`xargs`; `-print0` and `xargs -0` for names with spaces.

## Learn

### find walks a folder tree

`find` starts at a folder and visits everything under it: files, folders,
and the folders inside those. It prints each path it visits:

```bash
find backups
# backups
# backups/db
# backups/db/monday.sql
# backups/notes.txt
```

Every path starts with the folder **as you typed it**. Give it `backups`
and you get `backups/...`; give it `/srv/backups` and you get
`/srv/backups/...`.

### Tests: which paths to print

After the folder come *tests*. find prints only the paths that pass all of
them:

| Test | Passes for |
|---|---|
| `-type f` | regular files (not folders) |
| `-type d` | folders |
| `-name '*.sql'` | names that match the pattern |
| `-mtime +30` | files last modified more than 30 whole days ago (see below) |
| `-size +100M` | files bigger than 100 megabytes (`k` and `G` work too) |

```bash
find backups -type f -name '*.sql' -mtime +30
# backups/db/monday.sql     (if it is 31 or more days old)
```

In `-name`, `*` means "any characters", so `'*.sql'` matches names that
**end** in `.sql`. It doesn't match `monday.sql.gz`.

> **Why is it written like this?** **find requires** the folder first and
> the tests after it, each test a word starting with `-`. **Convention:**
> quote the pattern, `'*.sql'`. Without quotes, bash may swap `*.sql` for
> the matching files in your current folder before find ever sees it.

### -mtime: counting in whole days

`-mtime` counts a file's age in whole days and drops the fraction: a file
modified 7 days and 20 hours ago is 7 days old.

| Test | Passes for |
|---|---|
| `-mtime +7` | more than 7 whole days old: 8 days or older |
| `-mtime 7` | exactly 7 whole days old |
| `-mtime -7` | less than 7 whole days old: 0 to 6 |

So `+7` skips a file that is 7 days and 20 hours old. For a cleanup job
that's the safe side: a file is never deleted early.

### sort: a stable order

find prints paths in the order the disk hands them over, which can differ
between machines. Pipe the list into `sort` to put it in order, one path per
line, as in *Sort and uniq: top-N reports*:

```bash
find backups -type f | sort
```

### xargs: run a command on every path

A pipe sends find's output to the next command's **input**. But commands
like `cat`, `rm` and `wc` want file names as **arguments**. `xargs` bridges
that: it reads words from its input and adds them as arguments to a command.

```bash
find backups -type f -name '*.txt' | xargs wc -c
# runs: wc -c backups/notes.txt backups/todo.txt
```

### Spaces in names break plain xargs

Plain `xargs` splits its input at **every** space, not only at line ends. A
file called `release notes.txt` becomes two arguments, `release` and
`notes.txt`, and neither exists:

```text
cat: backups/release: No such file or directory
cat: notes.txt: No such file or directory
```

The fix comes in a pair. `-print0` makes find end each path with a *null
byte* (a character that can't appear in a file name) instead of a newline,
and `xargs -0` splits only at null bytes:

```bash
find backups -type f -name '*.txt' -print0 | xargs -0 wc -c
```

> **Why is it written like this?** **Both tools require** the pair:
> `-print0` without `-0` (or the other way round) mixes up the paths.
> **Convention:** whenever find feeds xargs, use `-print0 | xargs -0`, even
> if today's names have no spaces. Tomorrow's will.

### Deleting: -delete or xargs rm

find can delete what it finds. Put `-delete` **last**, after the tests:

```bash
find /tmp/build-cache -type f -name '*.tmp' -delete
```

`find ... -print0 | xargs -0 rm --` does the same (`--` tells rm that what
follows are file names, even one that starts with `-`). `-delete` is
shorter, and spaces can't trip it up. It also copes with finding nothing:
plain `xargs` still runs `rm` once with no names, and `rm` fails with
`missing operand`. There is no undo either way, so run the command without
`-delete` first and read the list.

> **Why is it written like this?** **find requires** `-delete` after the
> tests: find works left to right, and a `-delete` before the tests deletes
> everything it visits.

## Your task

The log folders on a worker node keep growing. Write the scripts for the
cleanup job: first look, then count, then delete.

Every script takes the folder as `$1` and prints paths starting with `$1`,
just as find does. To try them by hand, run `bash setup.sh` in the exercise
folder. It builds this tree in `sandbox/`:

```text
sandbox/
├── app.conf                    3 lines
├── api/
│   ├── access.log
│   ├── access.log.1            old
│   ├── api.conf                old, 4 lines
│   └── error.log               old
├── nginx/
│   ├── access.log
│   └── nginx.conf              5 lines
└── worker/
    ├── jobs/
    │   └── cron.log            old
    ├── old worker.log          old
    ├── queue settings.conf     2 lines
    └── worker.log
```

*Old* files were last modified more than 7 whole days ago (setup.sh dates
them 1 January 2024); the others are from today. Run `bash setup.sh` again
for a fresh copy whenever you like.

The tests build the same tree in a temporary folder (with the old files 10
days old and the others 2 days old), plus a second, different tree called
`srv`, and run your script from there: `bash 1_list_logs.sh sandbox`, then
`bash 1_list_logs.sh srv`.

### 1. List the log files

**Goal:** make `1_list_logs.sh` print every `.log` file under `$1`, sorted,
one path per line.

A `.log` file is a file whose name ends in `.log`. Rotated logs like
`access.log.1` don't count.

```bash
bash 1_list_logs.sh sandbox
# sandbox/api/access.log
# sandbox/api/error.log
# sandbox/nginx/access.log
# sandbox/worker/jobs/cron.log
# sandbox/worker/old worker.log
# sandbox/worker/worker.log
```

**Think it through:**
- Which two tests pick out files whose names end in `.log`?
- Which command after a pipe puts the list in order?

### 2. List the old log files

**Goal:** make `2_list_old_logs.sh` print every `.log` file under `$1` that
was last modified more than 7 whole days ago, sorted, one path per line.

```bash
bash 2_list_old_logs.sh sandbox
# sandbox/api/error.log
# sandbox/worker/jobs/cron.log
# sandbox/worker/old worker.log
```

`access.log.1` is old too, but it isn't a `.log` file.

**Think it through:**
- Which test is about age, and which sign means "more than"?
- Look at the `-mtime` table: which number goes after the sign?

### 3. Count the config lines

**Goal:** make `3_count_conf_lines.sh` print the total number of lines in
all the `.conf` files under `$1`, as one number.

```bash
bash 3_count_conf_lines.sh sandbox
# 14
```

That's 3 + 4 + 5 + 2. Two of the names have a space (`queue settings.conf`
here, `deploy notes.conf` in `srv`), and their lines count too. Spaces
around the number don't matter: on macOS, `wc` puts some in front.

**Think it through:**
- `cat` with several files prints them one after another. What counts the
  lines of all that?
- How do the paths from find become arguments to `cat`?
- What happens to `queue settings.conf` on the way, and which pair of
  flags prevents it?

### 4. Delete the old log files

**Goal:** make `4_delete_old_logs.sh` delete every `.log` file under `$1`
that was last modified more than 7 whole days ago.

Leave every other file alone: new logs, rotated logs like `access.log.1`,
and `.conf` files, however old. Folders stay too, even when they end up
empty. Nothing needs to be printed.

```bash
bash 4_delete_old_logs.sh sandbox
find sandbox -type f | sort
# sandbox/api/access.log
# sandbox/api/access.log.1
# sandbox/api/api.conf
# sandbox/app.conf
# sandbox/nginx/access.log
# sandbox/nginx/nginx.conf
# sandbox/worker/queue settings.conf
# sandbox/worker/worker.log
```

Then `bash setup.sh` brings the deleted files back.

**Think it through:**
- Task 2 already finds exactly the files to delete. What do you add?
- Where does `-delete` go, and why does that matter?

## Learn more

- [GNU findutils manual](https://www.gnu.org/software/findutils/manual/find.html)
- [The Missing Semester: Shell Tools and Scripting](https://missing.csail.mit.edu/2020/shell-tools/)
