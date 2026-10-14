# Undoing: rescue lost work with reflog

Someone ran `git reset --hard` one commit too far, and someone else deleted a
branch that was never merged. Both look like lost work. They aren't: Git
quietly logs every place `HEAD` has been, and the commits are still there.
The *reflog* is Git's safety net.

**New here:** `git reflog`, `HEAD@{n}`, `git branch <name> <commit>`, `git cherry-pick`, `git branch -D`.

## Learn

### Recap

- `bash setup.sh` rebuilds `sandbox/`; your scripts run inside it, in order
  (`cd sandbox && bash ../1_rescue_tls_commit.sh`). The tests do the same.
- A branch is just a pointer to a commit. Moving or deleting the pointer
  doesn't delete commits, as in *Branches: create, switch and merge*.

### Unreachable, not deleted

`git log` only shows commits you can reach from a branch by following
parents. After a `reset --hard` or a branch delete, some commits can't be
reached any more, so they vanish from `git log`. But they're still in
`.git/`. Git only throws unreachable commits away after weeks (30 to 90
days by default).

### The reflog

`git reflog` lists every move of `HEAD`, newest first: commits, switches,
resets, merges.

```text
$ git reflog
7a8b9c0 HEAD@{0}: commit: Bump nginx to 1.27
3d4e5f6 HEAD@{1}: checkout: moving from fix/tls to main
c1d2e3f HEAD@{2}: commit: Renew the wildcard certificate
3d4e5f6 HEAD@{3}: checkout: moving from main to fix/tls
```

`HEAD@{2}` means "where `HEAD` was two moves ago". You can use it, or the
hash next to it, anywhere Git wants a commit:

```bash
git branch rescue 'HEAD@{2}'        # a new branch pointing at that commit
git log --oneline -1 c1d2e3f        # look before you act
```

> **Why is it written like this?** **Convention:** quote `'HEAD@{2}'`. Some
> shells treat `{...}` specially, and quotes stop that. **Git's rule:**
> the reflog is local only. It's never pushed, and it only covers what
> happened in this copy of the repo.

### What the reflog can't save

Only committed work is safe. Uncommitted edits wiped by `git reset --hard`
or `git restore` were never in a commit, so there's nothing to go back to.
When in doubt, commit (even a "WIP" commit), and tidy up later.

### Cherry-pick: copy one commit

`git cherry-pick <commit>` applies that commit's change on top of the
current branch as a **new** commit, with the same message and author. It's
the usual way to carry one fix across: a hotfix from `main` onto a release
branch, or a rescued commit back onto `main`.

### Deleting unmerged branches

`git branch -d` refuses to delete a branch with commits that the branch
you're on doesn't have. `git branch -D` deletes it anyway. That's exactly how
`feature/alerts` got lost.

## Your task

Two pieces of work went missing from the `web` repo this week. Run
`bash setup.sh`, then `cd sandbox`, `git log --oneline --all` (they're not
there) and `git reflog` (they are).

### 1. Rescue the reset commit

**Goal:** in `1_rescue_tls_commit.sh`, create a branch named `rescue-tls`
pointing at the lost commit `Add TLS cert rotation`.

```text
$ git log --oneline -1 rescue-tls
7291b38 Add TLS cert rotation
```

Stay on `main`; `main` must not change.

`HEAD@{n}` numbers shift every time `HEAD` moves, including when you
explore with `git switch`. The tests use a fresh sandbox, so read the
numbers right after `bash setup.sh`, or use the hash: it never shifts.

**Think it through:**
- Which reflog line was that commit made in?
- `git branch` takes an optional second argument: what is it?

### 2. Bring back the deleted branch

**Goal:** in `2_restore_deleted_branch.sh`, recreate the branch
`feature/alerts` pointing at its last commit, `Add disk space alert`.

Its parent must be `Add app config`, as before. Stay on `main`.

**Think it through:**
- The branch is gone, but commits made on it are in `HEAD`'s reflog. Which line?
- Did task 1 add a reflog line? (Does `git branch` move `HEAD`?)

### 3. Copy the rescued commit onto main

**Goal:** in `3_cherry_pick_tls.sh`, cherry-pick `rescue-tls` onto `main`,
then delete the `rescue-tls` branch.

```text
$ git log --oneline -2
9f0a1b2 Add TLS cert rotation
eb1a91e Scale web to 3 replicas
```

Your new commit's hash will differ.

The new commit has one parent (it's a copy, not a merge) and adds
`scripts/rotate-certs.sh`. Afterwards `rescue-tls` is gone.

**Think it through:**
- Which branch must you be on to add a commit to `main`?
- The original commit is still on `rescue-tls`, which isn't merged. Which
  delete flag do you need?

## Learn more

- [Pro Git: Data Recovery](https://git-scm.com/book/en/v2/Git-Internals-Maintenance-and-Data-Recovery#_data_recovery)
- [git-reflog documentation](https://git-scm.com/docs/git-reflog)
- [git-cherry-pick documentation](https://git-scm.com/docs/git-cherry-pick)
