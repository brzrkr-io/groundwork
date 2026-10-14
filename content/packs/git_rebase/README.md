# Rebase: catch a branch up with main

Your feature branch started a week ago, and `main` has moved on since. Before
you open a pull request, you want your commits to sit on top of the newest
`main`, as if you'd started today. That's what *rebase* does, and it keeps
history a straight line.

**New here:** `A..B` ranges, `git rebase`, `git merge --ff-only`, linear history.

## Learn

### Recap

- `bash setup.sh` rebuilds `sandbox/`; your scripts run inside it, in order
  (`cd sandbox && bash ../1_list_branch_commits.sh`). The tests do the same.
- A merge of two branches that both moved makes a merge commit, as in
  *Branches: resolve a merge conflict*.

### Ranges: what's on one branch but not the other

`A..B` means "commits reachable from `B` but not from `A`": what `B` has that
`A` lacks.

```bash
git log --oneline main..fix/tls     # commits on fix/tls that main doesn't have
git log --format=%s -3              # just the subject line (%s) of the newest 3 commits
```

`--format` picks what to print for each commit; `%s` is the subject, `%h`
the short hash, `%an` the author's name.

### Rebase: replay your commits

`git rebase main` (run on your branch) takes the commits only your branch
has, and re-applies them one by one on top of `main`'s newest commit:

```text
before:  A---B---D---E  main
              \
               C1--C2  fix/tls

after:   A---B---D---E  main
                      \
                       C1'--C2'  fix/tls
```

`C1'` has the same change and message as `C1`, but a new parent, so it's a
**new commit with a new hash**. The old `C1` and `C2` are dropped from the
branch.

If a replayed commit conflicts, rebase stops: fix the file, `git add` it,
then `git rebase --continue` (or `git rebase --abort` to give up).

> **Why is it written like this?** **The golden rule:** only rebase commits
> nobody else has. Rebasing a shared branch replaces commits your teammates
> already built on, and their next pull turns into a mess. Your own,
> unpushed (or unshared) feature branch is fine.

### Merge or rebase?

| | `git merge` | `git rebase` |
|---|---|---|
| History | true record, with merge commits | straight line, rewritten |
| Safe on shared branches | yes | no |
| Typical use | landing a PR on `main` | updating your own branch before review |

After a rebase, `main` can take the branch as a fast-forward.
`git merge --ff-only <branch>` does exactly that, and **refuses** (instead
of making a merge commit) if a fast-forward isn't possible.

## Your task

`feature/retries` adds a retry helper to the deploy script. `main` has two
new commits since it started. Run `bash setup.sh`, then `cd sandbox` and
`git log --graph --oneline --all`.

### 1. List the commits only on the branch

**Goal:** `1_list_branch_commits.sh` prints the subject of each commit on
`feature/retries` that `main` doesn't have, newest first, one per line.

```text
Retry failed deploys
Add retry helper
```

Print subjects only: no hashes. Stay on `main`.

**Think it through:**
- Which range syntax means "on this branch, not on that one"?
- Which `--format` placeholder is the subject?

### 2. Rebase the branch onto main

**Goal:** in `2_rebase_onto_main.sh`, switch to `feature/retries` and rebase
it onto `main`.

```text
$ git log --oneline --graph
* 8c7d6e5 Retry failed deploys
* 4f3e2d1 Add retry helper
* 185618f Add README
* 5c511a1 Pin base image to alpine 3.20
* ...
```

Rebased commits are new commits, so their hashes differ from before (and
from this example).

Afterwards the branch has main's newest commit as an ancestor, its own two
commits on top, and no merge commits.

**Think it through:**
- Which branch must you be on: the one being moved, or the one it moves onto?
- Do any of the replayed commits touch files `main` changed?

### 3. Fast-forward main and clean up

**Goal:** in `3_fast_forward_main.sh`, switch to `main`, fast-forward it to
`feature/retries` with a merge that refuses anything else, and delete
`feature/retries`.

Afterwards `main` has six commits in a straight line, ending with
`Retry failed deploys`, and only `main` is left.

**Think it through:**
- Which `git merge` option allows only a fast-forward?
- Once merged, which flag of `git branch` deletes safely?

## Learn more

- [Pro Git: Rebasing](https://git-scm.com/book/en/v2/Git-Branching-Rebasing)
- [git-rebase documentation](https://git-scm.com/docs/git-rebase)
- [Pro Git: Commit Ranges](https://git-scm.com/book/en/v2/Git-Tools-Revision-Selection#_commit_ranges)
