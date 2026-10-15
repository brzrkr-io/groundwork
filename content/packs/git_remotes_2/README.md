# Remotes: fix a rejected push

You commit, you push, and Git says no: `! [rejected]`. It happens daily on
any team, and it isn't an error in your work. Git is refusing to throw away
commits it would lose. This exercise reads the two common rejections and
fixes each the safe way.

**New here:** rejected pushes (`fetch first`, `non-fast-forward`), diverged
branches in `git status`, `git pull --rebase`, `git push --force-with-lease`.

## Learn

### Recap

- `bash setup.sh` rebuilds `sandbox/`, which holds `origin.git` (the shared
  repo), `alex/` (a teammate's clone) and `web/` (yours). This time your
  scripts run **inside** `web`: `cd sandbox/web && bash ../../1_whats_waiting.sh`
  (two folders up, back to the exercise folder).
- `git fetch` updates `origin/main` without touching your branches;
  `git log main..origin/main` lists what arrived, as in *Remotes: clone,
  push and fetch*.
- Rebase replays your commits on top of another branch as new commits, as
  in *Rebase: catch a branch up with main*.

### Why a push gets rejected

A push only ever **fast-forwards** the remote's branch: it moves it to a
commit that already contains everything the branch had. If the remote has
a commit you don't, moving its branch to yours would drop that commit, so
Git refuses:

```text
 ! [rejected]        main -> main (fetch first)
hint: Updates were rejected because the remote contains work that you do not
hint: have locally.
```

After a `git fetch`, `git status` says the same thing in other words:

```text
Your branch and 'origin/main' have diverged,
and have 1 and 1 different commits each, respectively.
```

```text
        C  your main          (1 commit origin doesn't have)
       /
  A---B
       \
        D  origin/main        (1 commit you don't have)
```

### Fix 1: catch up, then push

Combine the two lines of work locally, then push the result, which now
contains everything:

| Command | Result |
|---|---|
| `git pull --rebase` | fetch, then replay your unpushed commits on top of `origin/main`: `A-B-D-C'`, a straight line |
| `git pull --no-rebase` | fetch, then merge: a merge commit joining C and D |

Plain `git pull` with diverged branches stops and asks you to pick one of
these (or set `git config pull.rebase true` once). Rebasing your own
**unpushed** commits is safe: nobody else has them yet. Then `git push`
works, because it's a fast-forward again.

```bash
git pull --rebase    # on a branch with an upstream
git push
```

### Fix 2: you rewrote your own pushed branch

Rebasing a branch you've already pushed (to bring it up to date before
review) makes new commits. The remote still has the old ones, so the push
is rejected with a different reason:

```text
 ! [rejected]        fix/dns -> fix/dns (non-fast-forward)
hint: Updates were rejected because the tip of your current branch is behind
hint: its remote counterpart.
```

Here `git pull` is the **wrong** fix: it would merge the old commits back
in, and you'd have every change twice. You *meant* to replace the branch,
so you force the push:

| Command | Replaces the remote branch... |
|---|---|
| `git push --force` | always, even if a teammate pushed to it since your last fetch, deleting their work |
| `git push --force-with-lease` | only if it's still where your `origin/<branch>` says; otherwise it's rejected |

> **Why is it written like this?** **Convention:** never force-push a shared
> branch like `main`; `git revert` is the shared-branch undo, as in
> *Undoing: revert a bad deploy*. On your own feature branch, force-pushing
> after a rebase is normal, and `--force-with-lease` is the habit that
> keeps it safe.

## Your task

You pushed `feature/retries` for review yesterday. This morning you
committed `Raise web replicas to 4` on `main` without fetching, while Alex
pushed `Pin base image to alpine 3.20`. Run `bash setup.sh`, then
`cd sandbox/web` and try `git push` yourself: it's rejected, and nothing
changes. Look around with `git log --oneline --graph --all`.

### 1. See what's waiting on origin

**Goal:** in `1_whats_waiting.sh`, fetch from origin, then print the subject
of each commit on `origin/main` that your `main` doesn't have, newest
first, one per line.

```text
$ bash ../../1_whats_waiting.sh
Pin base image to alpine 3.20
```

Print nothing else, and leave your branches as they are.

**Think it through:**
- Which command updates `origin/main` without touching `main`?
- Which range means "on `origin/main`, not on `main`"?

### 2. Rebase your commit and push main

**Goal:** in `2_rebase_and_push.sh`, put your unpushed commit on top of
Alex's by pulling with rebase, then push `main`.

Afterwards origin's `main` is a straight line, with no merge commit, and
your `main` is the same commit:

```text
$ git log --oneline main
6de9379 Raise web replicas to 4
f2333b6 Pin base image to alpine 3.20
f9538b6 Add deploy script
5521496 Add web config
```

Your newest hash will differ.

**Think it through:**
- Which `git pull` option replays your commit instead of merging?
- After that, is the push a fast-forward?

### 3. Push the rebased feature branch

**Goal:** in `3_push_rebased_branch.sh`, switch to `feature/retries`,
rebase it onto `main`, and push it with the force option that refuses if
someone else pushed to the branch.

Afterwards origin's `feature/retries` starts from the newest `main` and
holds the two rebased commits, `Add retry helper` and
`Retry failed deploys`. Your `feature/retries` is the same commit.

**Think it through:**
- Why would a plain `git push` be rejected after the rebase?
- Which of the two force options checks before it overwrites?

## Learn more

- [git-push documentation](https://git-scm.com/docs/git-push) (see "Note about fast-forwards" and `--force-with-lease`)
- [git-pull documentation](https://git-scm.com/docs/git-pull)
- [Pro Git: The Perils of Rebasing](https://git-scm.com/book/en/v2/Git-Branching-Rebasing#_rebase_peril)
