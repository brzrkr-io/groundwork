## Task 1: Start bisecting

`git bisect start` can take both endpoints at once: the bad commit first,
then the good one.

### Closer

`git bisect start <bad> <good>`, with a branch name and a tag name.

### Answer

```bash
git bisect start main v2.0.0
```

## Task 2: Let Git run the check

Three commands: `git bisect run` with the check command, a `git tag` that
names the commit at `refs/bisect/bad`, and `git bisect reset`. Order matters:
reset deletes the ref.

### Closer

`git bisect run bash check.sh`, then `git tag first-bad refs/bisect/bad`,
then `git bisect reset`.

### Answer

```bash
git bisect run bash check.sh
git tag first-bad refs/bisect/bad
git bisect reset
```

## Task 3: Revert the culprit

The tag names the culprit, so you can revert it by name. You're on `main`
after task 2's reset.

### Closer

`git revert --no-edit <commit>`, with the tag as `<commit>`.

### Answer

```bash
git revert --no-edit first-bad
```
