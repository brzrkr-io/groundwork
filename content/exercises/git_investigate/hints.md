## Task 1: Find the commits that touched a value

`git log -S` finds commits that added or removed some text. Combine it with
`--format` to print the author and subject with `: ` between them.

### Closer

`git log -S '<text>' --format='%an: %s'`. Quote the text: it has a space.

### Answer

```bash
git log -S 'max_connections: 500' --format='%an: %s'
```

## Task 2: Find who last changed a line

Two steps: get the hash from `git blame --porcelain -L 4,4`, then print that
commit with `git log -1 --format=...`.

### Closer

`git blame --porcelain -L 4,4 config/app.yaml | head -n 1 | cut -d ' ' -f 1`
prints the hash. Capture it with `commit=$(...)`, then use `"$commit"`.

### Answer

```bash
commit=$(git blame --porcelain -L 4,4 config/app.yaml | head -n 1 | cut -d ' ' -f 1)
git log -1 --format='%an: %s' "$commit"
```

## Task 3: List the files changed between two releases

`git diff` compares any two commits, and tags name commits. One option makes
it print paths instead of the full change.

### Closer

`git diff --name-only <old> <new>`.

### Answer

```bash
git diff --name-only v1.0.0 v1.1.0
```
