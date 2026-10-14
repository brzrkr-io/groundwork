## Task 1: Unstage your notes

You want the staging area to forget the file, while the file itself stays.
`git restore` can work on the staging area instead of the working tree.

### Closer

`git restore --staged <path>` copies the last commit's version into the
staging area; for a new file, that means "not staged". The working tree is
left alone.

### Answer

```bash
git restore --staged notes/todo.txt
```

## Task 2: Commit only the timeout fix

After task 1, the staging area holds just the timeout change. A plain commit
takes exactly that.

### Closer

`git commit -m "..."` without `-a`; don't `git add` anything first.

### Answer

```bash
git commit -m "Raise api timeout to 30s"
```

## Task 3: Throw away the debug line

The change to `deploy.sh` isn't staged, so the staging area still holds the
clean version. Copy it back over the working file.

### Closer

`git restore <path>` (no `--staged`) overwrites the working file with the
staged one.

### Answer

```bash
git restore deploy.sh
```
