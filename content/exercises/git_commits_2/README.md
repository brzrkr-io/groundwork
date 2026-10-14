# Commits: fix the last commit

Everyone commits too early sometimes: a typo in the message, a forgotten file,
three tiny commits that should be one. While the commits are still only on
your machine, Git lets you tidy them before anyone sees them.

**New here:** `git commit --amend`, `HEAD~n`, `git reset --soft` and `--mixed`.

## Learn

### Recap

- `bash setup.sh` rebuilds `sandbox/`; your scripts run inside it, in order
  (`cd sandbox && bash ../1_reword.sh`). The tests do the same.
- `git add` stages, `git commit -m` saves what's staged, `git log --oneline`
  lists commits.

### Amend: redo the newest commit

`git commit --amend` replaces the newest commit with a new one: whatever is
staged now, plus the old commit's contents.

```bash
git commit --amend -m "Rotate the TLS certificate"   # new message
git add nginx.conf
git commit --amend --no-edit                         # add a file, keep the message
```

Without `-m` or `--no-edit`, Git opens an editor for the message.

> **Why is it written like this?** **Git requires** `--amend` to make a
> *new* commit with a new hash: commits never change once made. The old one
> is simply no longer on the branch. That's why you only amend commits you
> haven't pushed: teammates may already have the old one.

### Naming older commits

`HEAD` is the commit you're on. `HEAD~1` is its parent, `HEAD~2` the one
before that:

```text
HEAD~2      HEAD~1      HEAD
  o ----------o-----------o
```

So `git log --oneline HEAD~2` lists history from two commits back.

### Reset: move the branch back

`git reset <commit>` moves the current branch back to `<commit>`. The commits
after it drop off the branch. What happens to their *changes* depends on the
mode:

| Command | Branch | Staging area | Working tree |
|---|---|---|---|
| `git reset --soft HEAD~1` | moved back | keeps the changes, staged | unchanged |
| `git reset HEAD~1` (`--mixed`) | moved back | reset | keeps the changes, unstaged |
| `git reset --hard HEAD~1` | moved back | reset | **changes thrown away** |

`--soft` is the "undo the commit, keep everything ready to commit again"
move:

```bash
git reset --soft HEAD~3     # drop the last three commits, keep their changes staged
git commit -m "Add the nginx role"
```

That turns three commits into one, which is called *squashing*.

> **Why is it written like this?** `--mixed` is **Git's default** when you
> give no mode. `--hard` is the only one that loses work, so think before
> typing it.

## Your task

You fixed the health check in a hurry. The newest commit has a typo-ridden
message and is missing its config file, and it really belongs with the
commit before it. Run `bash setup.sh` and look: `git log --oneline` and
`git status` in `sandbox/`.

### 1. Reword the last commit

**Goal:** in `1_reword.sh`, change the newest commit's message from
`fix teh helthcheck` to `Fix the health check timeout`.

```text
$ git log --oneline
a1b2c3d Fix the health check timeout
58c1f3e Add health check script
35309c9 Add deploy script
```

There must still be three commits: replace the commit, don't add one.

**Think it through:**
- Which commit option redoes the newest commit instead of adding one?
- How do you give the new message on the command line?

### 2. Add the forgotten file

**Goal:** in `2_add_forgotten_file.sh`, put `config/health.yaml` into the
newest commit, keeping its message `Fix the health check timeout`.

Afterwards, the newest commit changes `config/health.yaml` and
`healthcheck.sh`, there are still three commits, and nothing is left
uncommitted.

**Think it through:**
- An amended commit gets whatever is staged. What must happen first?
- Which option keeps the old message without opening an editor?

### 3. Squash the last two commits

**Goal:** in `3_squash.sh`, replace the newest two commits with one commit
whose message is `Add health check with a 5s timeout`.

```text
$ git log --oneline
f0e1d2c Add health check with a 5s timeout
35309c9 Add deploy script
```

The new commit must contain everything the two did: `healthcheck.sh` with
`--max-time 5`, and `config/health.yaml`.

**Think it through:**
- How do you write "two commits back" from `HEAD`?
- Which reset mode keeps the changes staged, ready to commit again?

## Learn more

- [Pro Git: Undoing Things](https://git-scm.com/book/en/v2/Git-Basics-Undoing-Things)
- [Pro Git: Reset Demystified](https://git-scm.com/book/en/v2/Git-Tools-Reset-Demystified)
- [git-reset documentation](https://git-scm.com/docs/git-reset)
