# Staging: choose what goes in a commit

A real working session leaves a mix behind: one fix you want to ship, a
debug line you don't, some personal notes. The staging area is how you pick
exactly what goes into the next commit, so a reviewer (or a 3 a.m. you)
sees one clean change.

**New here:** `git status --short`, `git diff`, `git diff --staged`, `git restore`, `git restore --staged`.

## Learn

### Recap

- `bash setup.sh` rebuilds `sandbox/`; your scripts run inside it, in order
  (`cd sandbox && bash ../1_unstage_notes.sh`). The tests do the same.
- A file lives in three places: the *working tree* (what you edit), the
  *staging area* (what the next commit gets) and the *repository* (commits).

### Reading `git status --short`

`git status --short` (or `-s`) prints two columns per file: the staging area
on the left, the working tree on the right.

```text
M  nginx.conf        staged change, ready to commit
 M deploy.sh         changed, not staged
A  certs/ca.pem      new file, staged
?? scratch.txt       untracked: Git has never seen it
```

### Two diffs

| Command | Compares | Shows |
|---|---|---|
| `git diff` | staging area vs working tree | what you changed but haven't staged |
| `git diff --staged` | last commit vs staging area | what the next commit will contain |

Run `git diff --staged` before every commit. It's the cheapest review
there is.

### Moving files back

`git restore` copies a file *backwards*, from a safer place over a less safe
one:

```bash
git restore --staged certs/ca.pem   # staging area <- last commit: unstage, keep your edits
git restore nginx.conf              # working tree <- staging area: discard your edits
```

> **Why is it written like this?** **Git's choice:** `--staged` picks which
> place gets overwritten. Without it, `git restore` overwrites your working
> file, and **those edits are gone for good**: they were never committed, so
> Git has no copy. Older guides use `git reset HEAD <file>` and
> `git checkout -- <file>` for the same two moves.

### Committing what's staged

`git commit -m "..."` commits the staging area only. Changes that are not
staged stay in your working tree, untouched. (`git commit -a` would stage
every tracked change first; that's how debug lines sneak into production.)

## Your task

You fixed the api timeout, staged it, and accidentally staged your personal
notes too. `deploy.sh` still has a debug line. Run `bash setup.sh`, then
`cd sandbox` and look at `git status --short`, `git diff` and
`git diff --staged`.

### 1. Unstage your notes

**Goal:** in `1_unstage_notes.sh`, take `notes/todo.txt` out of the staging
area, keeping the file on disk.

Afterwards, only `config/app.yaml` is staged:

```text
$ git status --short
M  config/app.yaml
 M deploy.sh
?? notes/
```

**Think it through:**
- Which `git restore` option touches the staging area?
- Which place must stay untouched so the notes survive?

### 2. Commit only the timeout fix

**Goal:** in `2_commit_fix.sh`, commit the staged change with the message
`Raise api timeout to 30s`.

The new commit changes only `config/app.yaml`. The debug line in `deploy.sh`
stays uncommitted, and `notes/` stays untracked.

**Think it through:**
- After task 1, is the staging area already exactly what you want?
- Which commit option would wrongly sweep `deploy.sh` in?

### 3. Throw away the debug line

**Goal:** in `3_discard_debug.sh`, put `deploy.sh` back the way it is in the
last commit.

Afterwards the only thing `git status --short` shows is your untracked notes:

```text
$ git status --short
?? notes/
```

**Think it through:**
- `deploy.sh` isn't staged. Which place does the clean copy come from?
- Is this undoable once done? (Read the note under *Moving files back*.)

## Learn more

- [Pro Git: Recording Changes to the Repository](https://git-scm.com/book/en/v2/Git-Basics-Recording-Changes-to-the-Repository)
- [git-restore documentation](https://git-scm.com/docs/git-restore)
- [git-status documentation: short format](https://git-scm.com/docs/git-status#_short_format)
