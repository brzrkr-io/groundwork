## Task 1: Revert the bad commit

Find the bad commit in `git log --oneline`: it's one below the newest. Revert
it by hash or by its position relative to `HEAD`.

### Closer

`git revert --no-edit <commit>`, where `<commit>` is the hash you saw, or
`HEAD~1`.

### Answer

```bash
git revert --no-edit HEAD~1
```

## Task 2: Throw away the experiment

Switch to `experiment` first: reset only moves the branch you're on. Then
move it back two commits with the mode that discards their changes, and
switch back.

### Closer

`git switch experiment`, `git reset --hard HEAD~2`, `git switch main`.

### Answer

```bash
git switch experiment
git reset --hard HEAD~2
git switch main
```

## Task 3: Restore a file from an old commit

`git restore` can copy a file from any commit, not just the last one. After
that it's an ordinary change: stage and commit.

### Closer

`git restore --source=v1.0 docs/runbook.md`, then `git add` and
`git commit -m "..."`.

### Answer

```bash
git restore --source=v1.0 docs/runbook.md
git add docs/runbook.md
git commit -m "Restore the rollback section of the runbook"
```
