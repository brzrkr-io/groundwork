# Processes and permissions: an incident on the box

The server is slow, a job won't stop, a secret is readable by everyone and
a cron job says `command not found`. This is the hands-on side of *The
operating system for programmers*: you find a process, stop it, look inside
it, fix file modes and work out why a script fails under cron.

**New here:** `ps`, `pgrep`, `kill` with `-TERM`, `-KILL` and `-0`, `seq`,
`$$`, `/proc/<pid>/fd` and `/proc/<pid>/cwd`, `readlink`, `chmod` with
octal modes, `umask`, how bash searches `PATH`, `$0` and `dirname`. It
recaps `command -v` from *Shell functions* and `pwd` from *Shell commands
and quoting*.

## Learn

The ideas are in the guide *The operating system for programmers*, in its
sections "Processes", "Permissions", "File descriptors and the standard
streams", "Exit codes" and "Signals". Here are the commands.

### Finding a process: ps and pgrep

Every running program is a *process* with a *PID* and a parent PID
(*PPID*). `ps -e` lists every process, and `-o` picks the columns:

```bash
ps -e -o pid,ppid,comm
#     PID    PPID COMMAND
#       1       0 systemd
#     812       1 sshd
#    4410     812 bash
```

`comm` is the process *name*: the file name of the program it runs, cut to
15 characters. `pgrep -x NAME` prints the PID of every process with exactly
that name. `ps -p PID` shows just that process, and a `=` after a column
name drops the header line:

```bash
pgrep -x sshd
# 812
ps -o pid=,comm= -p 812
#     812 sshd
```

> **Why is it written like this?** **ps requires** the columns after `-o`
> separated by commas, without spaces. **Convention:** `pgrep -x` in
> scripts, because without `-x` a part of the name is enough: `pgrep ssh`
> finds `sshd` and `ssh-agent` too.

`pgrep -f` searches whole command lines instead of names. Careful: a script
started as `bash find.sh sshd` has `sshd` in its own command line, so
`pgrep -f sshd` finds the script as well.

### Stopping a process: kill

`kill` sends a signal to a PID. Ask politely first, and force it only if the
process won't go:

| Command | Sends | The process |
|---|---|---|
| `kill -TERM 812` | SIGTERM (15) | may clean up and exit, or ignore it |
| `kill -KILL 812` | SIGKILL (9) | is ended by the kernel at once; it can't refuse |
| `kill -0 812` | nothing | exit status 0 while the process exists, 1 after |

A process stopped by a signal has no exit code of its own: its parent sees
128 plus the signal number. That's the 143 (128 + 15) and 137 (128 + 9)
you meet in `systemctl status`, CI logs and `kubectl describe pod`.

To give something time, check a few times with a short `sleep` in between.
`seq 5` prints 1 to 5, and `sleep` takes fractions of a second:

```bash
for _ in $(seq 5); do
	if [[ -f /tmp/ready ]]; then
		echo "ready"
		exit 0
	fi
	sleep 0.2
done
echo "not ready after 1 second" >&2
exit 1
```

> **Why is it written like this?** **Convention:** plain `kill` sends TERM
> too, but `kill -TERM` tells the reader you meant it. Old scripts write
> numbers, `kill -9`; names are easier to read.

### Looking inside a process: /proc

Linux shows every process as a folder, `/proc/<pid>/`. `$$` is your own
shell's PID, so you can look at yourself:

```bash
ls -l /proc/$$/fd
# lrwx------ 1 you users 64 Oct 13 09:15 0 -> /dev/pts/0
# lrwx------ 1 you users 64 Oct 13 09:15 1 -> /dev/pts/0
# lrwx------ 1 you users 64 Oct 13 09:15 2 -> /dev/pts/0
```

- `/proc/<pid>/fd/` has one entry per open file, named after its *file
  descriptor* number: 0, 1 and 2 are stdin, stdout and stderr.
- `/proc/<pid>/cwd` is the process's working directory.

These are *symbolic links* (symlinks): tiny files that point at another
path. `readlink` prints where one points:

```bash
readlink /proc/$$/cwd
# /home/you/groundwork
```

You can look inside your own processes (root can look inside all). `lsof -p
PID` shows the same, but it isn't installed everywhere; `/proc` always is.

### Permissions: chmod and umask

`ls -l` shows a file's *mode*, as the guide explains: `-rw-r--r--` is `644`.
`chmod` with three octal digits sets all three groups at once, whatever they
were before. `chmod +x` only adds the x bits:

```bash
chmod 640 app.yaml      # rw-r-----
chmod +x build.sh       # adds x, keeps the rest
```

A new file starts as 666 (`rw-rw-rw-`), and the *umask* removes bits from
it. The usual umask is 022, so new files are 644: readable by everyone.
After `umask 077`, every file the script creates starts as 600:

```bash
umask 077
date > private-notes.txt    # rw------- from the moment it exists
```

Creating the file first and running `chmod` after leaves a moment when
anyone could open it. The umask closes that gap.

> **Why is it written like this?** **chmod requires** octal digits, and
> always reads them as octal, so no leading `0` is needed (Go needs `0640`).
> **Your choice:** digits when you want an exact mode, `+x` when you want to
> add one thing.

### "command not found": PATH and the working directory

When you type `healthcheck`, bash looks for a file of that name in each
folder of `PATH`, in order. If none has it, bash says `command not found`
and the status is 127. `command -v` shows which file would run:

```bash
echo "$PATH"
# /home/you/bin:/usr/local/bin:/usr/bin:/bin
command -v kubectl
# /usr/local/bin/kubectl
PATH="/opt/tools/bin:$PATH"   # from now on, look in /opt/tools/bin first
```

Cron, systemd and CI start your script with their own `PATH` and their own
working directory (often your home folder). So a script that works by hand
fails there:

| Symptom | Check | Usual cause |
|---|---|---|
| `tool: command not found` | `echo "$PATH"`, `command -v tool` | the tool's folder isn't on `PATH` |
| `app.conf: No such file or directory` | `pwd` | relative paths start from a different folder |

### Where is my script? $0 and dirname

`$0` is the script's path as it was started: `bash /srv/jobs/backup.sh`
gives `/srv/jobs/backup.sh`. `dirname` drops the last part of a path:

```bash
dirname /srv/jobs/backup.sh
# /srv/jobs
```

After a `cd`, `$PWD` holds the new working directory as an absolute path,
the same one `pwd` prints, even if you `cd`'d to a relative one.

## Your task

You're on call for a box that's misbehaving. Write the five scripts you'd
reach for: find, stop, inspect, lock down, fix the cron job.

To try them by hand, run `bash setup.sh` in the exercise folder. It builds
`sandbox/`, makes `bin/healthcheck` runnable, and starts two practice
processes in the background:

| Name | Behaves like |
|---|---|
| `report-export` | a stuck job: ignores TERM, works in `sandbox/reports`, keeps `export.csv` open |
| `cache-warmer` | a well-behaved worker: stops on TERM |

Run `bash setup.sh` again for a fresh start; it stops the old practice
processes first. The tests don't use them: they start their own, with names
like `export-48213`, and stop them afterwards. Tasks 1–3 read `/proc`, so
they need Linux; elsewhere their tests are skipped.

### 1. Find the runaway

**Goal:** make `1_find_runaway.sh` print the PID and the parent PID of the
process whose name is `$1`.

Print both numbers on one line, PID first. Spaces around them don't matter,
so `ps` padding is fine. Exactly one process has the name.

```bash
bash 1_find_runaway.sh report-export
# 48213       1
```

Your numbers differ. The parent is 1, or your login session's service
manager: setup.sh has exited, and orphans get adopted. In the tests, the
parent is the test program.

**Think it through:**
- Which command finds a PID from an exact name? Which one prints chosen
  columns for one PID?
- Why would `pgrep -f "$1"` find two processes here?

### 2. Stop it politely, then by force

**Goal:** make `2_stop.sh` stop the process whose PID is `$1`: send TERM,
wait up to 1 second, send KILL if it's still there, and print which signal
stopped it.

Check whether it's gone every 0.1 seconds, at most 10 times. Then print
exactly one of these lines, with the exit code its parent sees:

```text
stopped by TERM (exit code 143)
killed by KILL (exit code 137)
```

```bash
bash 2_stop.sh "$(pgrep -x cache-warmer)"
# stopped by TERM (exit code 143)
bash 2_stop.sh "$(pgrep -x report-export)"
# killed by KILL (exit code 137)
```

The tests check that a process that stops on TERM never gets KILL, and that
one that ignores TERM gets at least half a second before KILL. Real scripts
wait longer (Kubernetes gives a pod 30 seconds); 1 second keeps the tests
fast.

**Think it through:**
- Which `kill` checks whether a PID still exists, without hurting it?
- What should the script do as soon as the process is gone? And when the
  loop ends and it's still there?

### 3. Look inside it

**Goal:** make `3_open_files.sh` print the working directory of the process
whose PID is `$1`, then every file it has open.

The first line is `cwd`, a space and the directory. Then comes one line per
entry in `/proc/$1/fd`: the fd number, a space, and where it points. The fd
lines may come in any order.

```bash
bash 3_open_files.sh "$(pgrep -x report-export)"
# cwd /home/you/groundwork/.../sandbox/reports
# 0 /dev/null
# 1 /dev/null
# 2 /dev/null
# 255 /home/you/groundwork/.../sandbox/report-export
# 3 /home/you/groundwork/.../sandbox/reports/export.csv
```

fd 255 is bash keeping its own script open. In the tests, fds 0 and 1 are
pipes, which show as something like `pipe:[18807555]`: print them like any
other.

**Think it through:**
- Which command prints where a link points?
- How do you loop over the entries in a folder, and get `3` from
  `/proc/48213/fd/3`?

### 4. Lock down the secret

**Goal:** make `4_fix_permissions.sh` set the modes of two files in the
folder `$1`, and write `$2` to a new file `$1/token` that only you can read.

| File | Mode afterwards |
|---|---|
| `$1/secret.env` | `600` (`rw-------`), whatever it was before |
| `$1/deploy.sh` | `755` (`rwxr-xr-x`), whatever it was before |
| `$1/token` | `600`, holding `$2` and a newline |

The tests run your script with umask 022, the usual default. Create the
token so that it is `600` from the moment it exists.

```bash
bash 4_fix_permissions.sh sandbox/app tok-123
ls -l sandbox/app
# -rwxr-xr-x 1 you users 41 Oct 13 09:15 deploy.sh
# -rw------- 1 you users 20 Oct 13 09:15 secret.env
# -rw------- 1 you users  8 Oct 13 09:15 token
```

**Think it through:**
- One test case starts with `secret.env` at 666. Which kind of `chmod` gets
  it to exactly 600?
- What has to happen before the token file is created?

### 5. Fix the cron job

**Goal:** make `5_run_checks.sh` work from any folder: go to the script's
own folder, put its `bin` folder on `PATH`, print `using` and the full path
of the `healthcheck` that will run to stderr, then run the checks.

`bin/healthcheck` and `checks.conf` sit next to the script. A teammate tested
it from the exercise folder with `PATH="$PWD/bin:$PATH" bash
5_run_checks.sh`. Cron runs it from your home folder with its own `PATH`:

```bash
cd ~ && bash ~/groundwork/.../5_run_checks.sh
# .../5_run_checks.sh: line 14: healthcheck: command not found
```

Fixed, the same command prints:

```bash
cd ~ && bash ~/groundwork/.../5_run_checks.sh
# using /home/you/groundwork/.../bin/healthcheck     (this line to stderr)
# ok api.internal:8080
# ok db.internal:5432
```

The `using` path is what `command -v healthcheck` prints once the `bin`
folder is on `PATH`, so add that folder as an absolute path. The tests copy
your script and `bin/healthcheck` into a temporary folder (one with a space
in its name), with their own `checks.conf`, and run it from another folder.

**Think it through:**
- Which variable holds the script's path, and which command turns that
  into its folder?
- After you `cd` there, which variable holds that folder as an absolute
  path?

## Learn more

- [The Linux man-pages project: proc(5)](https://man7.org/linux/man-pages/man5/proc.5.html)
- [The Linux man-pages project: signal(7)](https://man7.org/linux/man-pages/man7/signal.7.html)
- [GNU Bash manual: Command Search and Execution](https://www.gnu.org/software/bash/manual/html_node/Command-Search-and-Execution.html)
- [The Missing Semester: Command-line Environment](https://missing.csail.mit.edu/2020/command-line/)
