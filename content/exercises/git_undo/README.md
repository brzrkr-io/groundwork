# Undoing: revert a bad deploy

A commit on `main` took production down. You can't just delete it: `main` is
shared, and everyone already has it. Git's answer is a new commit that does
the exact opposite. This exercise covers that and two more ways to undo:
throw away local commits, and bring back one file from the past.

**New here:** `git revert`, `git reset --hard`, `git restore --source`, tags as names for commits.

## Learn

### Recap

- `bash setup.sh` rebuilds `sandbox/`; your scripts run inside it, in order
  (`cd sandbox && bash ../1_revert_bad_deploy.sh`). The tests do the same.
- `HEAD~1` is the commit before `HEAD`, as in *Commits: fix the last commit*.

### Naming a commit

Anywhere Git wants a commit, you can give:

| Name | Example |
|---|---|
| a hash, or its first 7+ characters | `3f9c2e1` |
| a branch | `main` |
| a *tag* (a fixed name for one commit, often a release) | `v2.3.0` |
| a position relative to one of those | `HEAD~2`, `main~1`, `v2.3.0~1` |

`git log --oneline` shows you the hashes. Because `setup.sh` uses fixed dates,
the sandbox's hashes are the same every time you rebuild it.

### Revert: undo with a new commit

`git revert <commit>` makes a **new** commit that applies the opposite of
`<commit>`'s change. History only grows, so it's safe on shared branches.

```bash
git revert --no-edit 3f9c2e1
```

```text
$ git log --oneline -2
c4d5e6f Revert "Enable HTTP/2 on the load balancer"
b8a9c0d Bump nginx to 1.27
```

`--no-edit` keeps the default message, `Revert "<original subject>"`.

### Reset --hard: throw local commits away

`git reset --hard <commit>` moves the current branch to `<commit>` and makes
the staging area and working tree match it. The commits after it drop off
the branch, and so do any uncommitted changes.

| Situation | Use |
|---|---|
| the commit is on a shared branch (pushed) | `git revert` |
| the commits are local only, and you want them gone | `git reset --hard` |

> **Why is it written like this?** **Git's rule:** reset moves the branch
> you're **on**, so switch to the branch first. Rewriting a pushed branch
> forces everyone else to repair theirs; reverting doesn't. And `--hard`
> discards uncommitted work for good, so check `git status` first.

### Restore one file from the past

`git restore --source=<commit> <path>` puts an old version of one file into
your working tree, leaving everything else alone. Then commit it like any
change:

```bash
git restore --source=v2.3.0 nginx/nginx.conf
```

## Your task

Production is down after the last deploy of `web`. Run `bash setup.sh`, then
`cd sandbox`, `git log --oneline --all --graph` and `git tag`.

### 1. Revert the bad commit

**Goal:** in `1_revert_bad_deploy.sh`, revert the commit `Point web at the new database cluster` on `main`, with the default message.

```text
$ git log --oneline -3
d1e2f3a Revert "Point web at the new database cluster"
72744f3 Update deploy steps in the runbook
aa0b2d9 Point web at the new database cluster
```

`config/db.yaml` goes back to `host: db-old.internal`, and every older commit
stays in history.

**Think it through:**
- How far back from `HEAD` is the bad commit? Or use its hash.
- Which option avoids the editor?

### 2. Throw away the experiment

**Goal:** in `2_drop_experiment.sh`, move the `experiment` branch back two
commits, discarding them, then switch back to `main`.

Afterwards `experiment` points at `Update deploy steps in the runbook` (where
it started), you're on `main`, and nothing is left uncommitted.

**Think it through:**
- Reset moves the current branch. Which branch must you be on?
- Which reset mode also throws away the files' changes?

### 3. Restore a file from an old commit

**Goal:** in `3_restore_runbook.sh`, restore `docs/runbook.md` exactly as it
was at tag `v1.0`, and commit it with the message
`Restore the rollback section of the runbook`.

Afterwards the newest commit on `main` changes only `docs/runbook.md`, and
the file matches `v1.0`'s version (with its `## Roll back` section).

**Think it through:**
- Which option of `git restore` picks the commit to copy from?
- A restored file is just a change. What makes it a commit?

## Learn more

- [Pro Git: Undoing Things](https://git-scm.com/book/en/v2/Git-Basics-Undoing-Things)
- [git-revert documentation](https://git-scm.com/docs/git-revert)
- [Atlassian: Resetting, checking out and reverting](https://www.atlassian.com/git/tutorials/resetting-checking-out-and-reverting)
