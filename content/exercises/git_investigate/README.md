# Investigating history: log -S, blame and diff

During an incident the questions come fast: when did this value change, who
changed it, and what went out in the last release? The answers are already
in Git's history, if you know how to ask.

**New here:** `git log -S`, `--format` placeholders, `git blame`, `git show`, `git diff A B --name-only`.

## Learn

### Recap

- `bash setup.sh` rebuilds `sandbox/`; your scripts run inside it, in order
  (`cd sandbox && bash ../1_find_connection_limit.sh`). The tests do the same.
- `git log --format=%s` prints subjects, as in *Rebase: catch a branch up with main*.
- A pipe (`|`) feeds one command's output to the next; `$(...)` captures
  output into a variable (from *Linux and the shell*).

### Choosing what `git log` prints

| Placeholder | Prints |
|---|---|
| `%h` / `%H` | short / full hash |
| `%an` | author name |
| `%ad` | author date |
| `%s` | subject |

```bash
git log --format='%h %an %s' -3
```

### Searching history: `git log -S`

`git log -S '<text>'` lists the commits where the number of times `<text>`
appears changed: the commits that **added or removed** it. It's the fastest
way to answer "when did this value appear, and when did it go away?"

```bash
git log -S 'worker_processes 8' --oneline
```

Add `-p` to see each change, or `-- <path>` to search one file only.

### Who changed this line: `git blame`

`git blame <file>` prints every line with the commit that last changed it.
`-L 10,12` limits it to lines 10 to 12:

```text
$ git blame -L 2,2 nginx.conf
3f9c2e1d (Priya Shah 2026-09-01 09:04:00 +0000 2) worker_processes 8;
```

For scripts, `--porcelain` is the machine-readable form. Its first line
starts with the full hash, then a space:

```text
3f9c2e1d4b5a69788796a5b4c3d2e1f0a9b8c7d6 2 2 1
```

`git show <commit>` then prints that commit's message and change.

> **Why is it written like this?** **Git's choice:** blame finds the
> **last** change only. If that was a reformat, blame the line again starting
> from the commit before it (`git blame <commit>~1 -L ...`) to dig deeper.

### What changed between two releases

`git diff <old> <new>` shows the change between any two commits, often two
tags. `--name-only` prints just the file paths, `--stat` a summary.

```bash
git diff --stat v2.3.0 v2.4.0
```

## Your task

The `web` service has been timing out under load. Before the postmortem,
you dig through the config history. Run `bash setup.sh`, then `cd sandbox`
and `git log --format='%h %an %s'`.

### 1. Find the commits that touched a value

**Goal:** `1_find_connection_limit.sh` prints every commit that added or
removed the text `max_connections: 500`, newest first, as `<author name>: <subject>`.

```text
Priya Shah: Restore normal connection limit
Alex Kim: Raise connection limit for the sale
```

One commit per line, nothing else.

**Think it through:**
- Which `git log` option searches for added or removed text?
- Which placeholders are the author name and the subject? What fixed text goes between them?

### 2. Find who last changed a line

**Goal:** `2_blame_timeout.sh` prints the commit that last changed line 4 of
`config/app.yaml` (the `timeout:` line) as `<author name>: <subject>`.

```text
Alex Kim: Lower timeout to 5s
```

Use `git blame` to find it: the newest commit that touched the **file** is a
different one.

**Think it through:**
- Which `git blame` options pick one line and give script-friendly output?
- How do you cut the first word (the hash) from the first line, and feed it to `git log`?

### 3. List the files changed between two releases

**Goal:** `3_files_changed_in_release.sh` prints the paths of the files that
differ between tags `v1.0.0` and `v1.1.0`, one per line.

```text
config/app.yaml
deploy.sh
docs/runbook.md
```

**Think it through:**
- Which command compares two commits?
- Which option prints paths only?

## Learn more

- [Pro Git: Viewing the Commit History](https://git-scm.com/book/en/v2/Git-Basics-Viewing-the-Commit-History)
- [Pro Git: Debugging with Git (blame)](https://git-scm.com/book/en/v2/Git-Tools-Debugging-with-Git)
- [git-log documentation: pretty formats](https://git-scm.com/docs/git-log#_pretty_formats)
