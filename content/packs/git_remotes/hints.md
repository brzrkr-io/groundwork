## Task 1: Clone the repository

`git clone` takes where to copy from and the folder to create. Settings
without `--global` go into the repo you're in, so `cd` into the clone first.

### Closer

`git clone origin.git web`, then `cd web`, then `git config user.name "..."`
and `git config user.email "..."`.

### Answer

```bash
git clone origin.git web
cd web
git config user.name "Sam Rivera"
git config user.email "sam@example.com"
```

## Task 2: Push a branch

Everything happens in `web`: make the branch, change the file, commit as
usual, then push. A new branch needs the remote, the branch name and the
option that sets the upstream.

### Closer

`git switch -c fix/timeout`, then `echo "..." >> config/app.yaml`, `git add`
and `git commit -m "..."`. Finish with `git push -u origin fix/timeout`.

### Answer

```bash
cd web
git switch -c fix/timeout
echo "request_timeout: 30s" >> config/app.yaml
git add config/app.yaml
git commit -m "Add a 30s request timeout"
git push -u origin fix/timeout
```

## Task 3: Fetch as Alex

Go into `alex`. Fetch updates only the `origin/...` bookmarks, which is
exactly what a reviewer wants. Then list a range of commits, subjects only.

### Closer

`cd alex`, `git fetch`, then `git log --format=%s A..B`, where `A` is
Alex's branch and `B` the remote-tracking branch.

### Answer

```bash
cd alex
git fetch
git log --format=%s main..origin/fix/timeout
```

## Task 4: Land the fix and pull it

In `web`: switch to `main`, merge with the option that only fast-forwards,
push. `main` already knows its upstream. Then go next door to `alex` and
pull, with the same only-fast-forward option.

### Closer

`git merge --ff-only fix/timeout` and plain `git push` in `web`; then
`cd ../alex` and `git pull --ff-only`.

### Answer

```bash
cd web
git switch main
git merge --ff-only fix/timeout
git push

cd ../alex
git pull --ff-only
```
