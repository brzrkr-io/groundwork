## Task 1: Rescue the reset commit

Run `git reflog` in the sandbox and find the line that says
`commit: Add TLS cert rotation`. Its `HEAD@{n}` or its hash names the commit.

### Closer

`git branch <name> <commit>` creates a branch at any commit without
switching to it. The commit is `'HEAD@{5}'`, or the hash on that line.

### Answer

```bash
git branch rescue-tls 'HEAD@{5}'
```

## Task 2: Bring back the deleted branch

The same trick: find `commit: Add disk space alert` in the reflog. Creating a
branch doesn't move `HEAD`, so task 1 didn't shift the numbers.

### Closer

`git branch feature/alerts <commit>`, where the commit is `'HEAD@{2}'` or the
hash on that reflog line.

### Answer

```bash
git branch feature/alerts 'HEAD@{2}'
```

## Task 3: Copy the rescued commit onto main

You're on `main`. Copy the commit from `rescue-tls` onto it, then delete the
branch with the flag that doesn't check for merges.

### Closer

`git cherry-pick <branch-or-commit>`, then `git branch -D rescue-tls`.

### Answer

```bash
git cherry-pick rescue-tls
git branch -D rescue-tls
```
