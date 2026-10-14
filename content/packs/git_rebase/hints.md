## Task 1: List the commits only on the branch

`git log` takes a range of two branch names with `..` between them, and
`--format` chooses what to print per commit.

### Closer

`git log --format=%s A..B` prints the subjects of commits on `B` but not on
`A`. Which branch is `A` here?

### Answer

```bash
git log --format=%s main..feature/retries
```

## Task 2: Rebase the branch onto main

Rebase moves the branch you're on. Get onto the feature branch, then rebase
it onto `main`.

### Closer

`git switch feature/retries`, then `git rebase <base>`.

### Answer

```bash
git switch feature/retries
git rebase main
```

## Task 3: Fast-forward main and clean up

After the rebase, `main` is an ancestor of the branch, so `main` can just
slide forward. Ask `git merge` to do only that.

### Closer

`git switch main`, `git merge --ff-only feature/retries`, then
`git branch -d feature/retries`.

### Answer

```bash
git switch main
git merge --ff-only feature/retries
git branch -d feature/retries
```
