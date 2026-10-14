## Task 1: See what's waiting on origin

Fetching only updates the `origin/...` bookmarks. Then list the commits in a
range, subjects only.

### Closer

`git fetch`, then `git log --format=%s A..B`, where `A` is your branch and
`B` the remote-tracking branch.

### Answer

```bash
git fetch
git log --format=%s main..origin/main
```

## Task 2: Rebase your commit and push main

Pull, but tell it to replay your commit on top instead of merging. Once
your `main` contains Alex's commit, the push is a fast-forward.

### Closer

`git pull --rebase`, then plain `git push`: `main` already has its
upstream.

### Answer

```bash
git pull --rebase
git push
```

## Task 3: Push the rebased feature branch

Switch, rebase onto `main` as in the rebase exercise, then push. The plain
push is rejected because origin still has the old commits, and you meant to
replace them.

### Closer

`git switch feature/retries`, `git rebase main`, then `git push` with the
force option that has "lease" in its name.

### Answer

```bash
git switch feature/retries
git rebase main
git push --force-with-lease
```
