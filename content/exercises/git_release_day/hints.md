## Task 1: Revert the debug logging

Look at `git log --oneline`: the debug logging commit is one below the
newest. Revert it by hash or by its position.

### Closer

`git revert --no-edit <commit>`; the commit is `HEAD~1` or its hash.

### Answer

```bash
git revert --no-edit HEAD~1
```

## Task 2: Merge the rate-limit feature

You're on `main`. Merge the branch with the default message, then delete the
merged branch with the safe flag.

### Closer

`git merge --no-edit <branch>`, then `git branch -d <branch>`.

### Answer

```bash
git merge --no-edit feature/rate-limit
git branch -d feature/rate-limit
```

## Task 3: Bump the version and tag the release

Three steps: replace the file's contents, commit it, tag the new commit.

### Closer

`echo "1.4.0" > VERSION`, then `git add` and `git commit -m "..."`, then
`git tag <name>`, which tags `HEAD`.

### Answer

```bash
echo "1.4.0" > VERSION
git add VERSION
git commit -m "Release v1.4.0"
git tag v1.4.0
```

## Task 4: Print the release notes

`git log` with a range of two tags, a format with fixed text before the
subject, and the option that skips merge commits.

### Closer

`git log --no-merges --format='- %s' <old>..<new>`.

### Answer

```bash
git log --no-merges --format='- %s' v1.3.0..v1.4.0
```
