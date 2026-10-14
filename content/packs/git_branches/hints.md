## Task 1: Create a feature branch

One command creates the branch and moves you onto it. Look at the
*Creating and switching* examples.

### Closer

`git switch -c <name>` creates and switches. The name is
`feature/healthcheck`.

### Answer

```bash
git switch -c feature/healthcheck
```

## Task 2: Commit on the branch

You're already on the branch after task 1. Commit the file the usual way:
stage, then commit.

### Closer

`git add healthcheck.sh`, then `git commit -m "..."`. Only the current
branch moves.

### Answer

```bash
git add healthcheck.sh
git commit -m "Add health check script"
```

## Task 3: Merge the branch into main

Merging pulls the named branch into the branch you're on, so get onto
`main` first. Delete the feature branch last.

### Closer

`git switch main`, then `git merge feature/healthcheck`, then
`git branch -d feature/healthcheck`.

### Answer

```bash
git switch main
git merge feature/healthcheck
git branch -d feature/healthcheck
```
