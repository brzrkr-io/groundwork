## Task 1: Reword the last commit

You don't want a new commit; you want to redo the newest one with a different
message. `git commit` has an option for exactly that.

### Closer

`git commit --amend` redoes the newest commit, and `-m "..."` gives the new
message.

### Answer

```bash
git commit --amend -m "Fix the health check timeout"
```

## Task 2: Add the forgotten file

Stage the file first, then amend. Amending puts what's staged into the
newest commit.

### Closer

`git add config/health.yaml`, then `git commit --amend` with the option that
keeps the existing message.

### Answer

```bash
git add config/health.yaml
git commit --amend --no-edit
```

## Task 3: Squash the last two commits

Move the branch back two commits, keeping the changes staged, then commit
them once with the new message.

### Closer

`git reset --soft HEAD~2` drops two commits from the branch and leaves their
changes staged. Then `git commit -m "..."`.

### Answer

```bash
git reset --soft HEAD~2
git commit -m "Add health check with a 5s timeout"
```
