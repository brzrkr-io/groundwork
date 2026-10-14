# Bisect: find the commit that broke the check

The check passed at the last release. Today it fails, and there are a dozen
commits in between. Reading them all is slow. *Bisecting* halves the
suspects with each test, so even 1,000 commits take about ten checks, and
Git can run the check for you.

**New here:** `git bisect start`, `git bisect run`, `refs/bisect/bad`, `git tag`, `git bisect reset`.

## Learn

### Recap

- `bash setup.sh` rebuilds `sandbox/`; your scripts run inside it, in order
  (`cd sandbox && bash ../1_start_bisect.sh`). The tests do the same.
- A tag names one commit, as in *Undoing: revert a bad deploy*. `git revert --no-edit <commit>`
  undoes a commit with a new one.

### Binary search on history

You know one *good* commit (the check passed) and one *bad* commit (it
fails). Somewhere between them is the first bad commit. Test the commit in
the middle: if it's bad, the culprit is in the older half; if it's good,
the newer half. Repeat.

```text
good                                     bad
 o----o----o----o----o----o----o----o----o
                     ^ test this one first
```

That's *binary search*: each test halves the suspects, so 16 commits take 4
tests.

### `git bisect`

```bash
git bisect start <bad> <good>    # e.g. git bisect start HEAD v3.1.0
```

Git checks out a commit in the middle (you're on no branch while
bisecting; that's normal). You test it and answer `git bisect good` or
`git bisect bad`, and Git checks out the next one, until it prints:

```text
4d5e6f7a... is the first bad commit
```

### Let Git run the check

`git bisect run <command>` does the answering for you. It runs `<command>`
on each commit: exit code `0` means good, `1` to `127` (except `125`) means
bad.

```bash
git bisect run bash smoke-test.sh
```

When it's done, the first bad commit is at `refs/bisect/bad`, a ref you can
use like a branch name until the bisect ends.

### Tags and finishing up

`git tag <name> <commit>` gives a commit a permanent name (`<commit>`
defaults to `HEAD`):

```bash
git tag broke-login 4d5e6f7
```

`git bisect reset` ends the bisect and puts you back on the branch you
started from. It also deletes `refs/bisect/bad`, so tag first.

> **Why is it written like this?** **Git requires** a check whose exit code
> tells the truth: `0` for good, non-zero for bad. That's why a check script
> must end with the command whose exit code matters; here `grep -q`
> returns `0` when it finds the line.

## Your task

`check.sh` makes sure the web service listens on port `8080`. It passed at
tag `v2.0.0` and fails on `main`. Run `bash setup.sh`, then `cd sandbox`,
`bash check.sh; echo $?` and `git log --oneline v2.0.0..main`.

### 1. Start bisecting

**Goal:** in `1_start_bisect.sh`, start a bisect with `main` as the bad
commit and `v2.0.0` as the good one.

Afterwards Git is bisecting: `refs/bisect/bad` is `main`'s commit, and
`v2.0.0`'s commit is marked good.

**Think it through:**
- `git bisect start` takes the bad commit first, then the good one.
- Where is `HEAD` after the script? Try `git status` by hand.

### 2. Let Git run the check

**Goal:** in `2_run_bisect.sh`, run `bash check.sh` on every commit with
`git bisect run`, tag the first bad commit `first-bad`, then end the bisect.

```text
$ git log --oneline -1 first-bad
cd6419b Tidy up config formatting
```

Afterwards you're back on `main`, and no bisect is in progress.

**Think it through:**
- What does `git bisect run` need: a command, and its arguments?
- Which ref holds the first bad commit, and when does it disappear?

### 3. Revert the culprit

**Goal:** in `3_revert_culprit.sh`, revert the commit tagged `first-bad` on
`main`, with the default message.

```text
$ git log --oneline -1
2b3c4d5 Revert "Tidy up config formatting"
```

Afterwards `bash check.sh` exits `0` again.

**Think it through:**
- A tag works anywhere a commit hash does.
- Which option skips the editor?

## Learn more

- [Pro Git: Debugging with Git (binary search)](https://git-scm.com/book/en/v2/Git-Tools-Debugging-with-Git#_binary_search)
- [git-bisect documentation](https://git-scm.com/docs/git-bisect)
- [Binary search algorithm](https://en.wikipedia.org/wiki/Binary_search_algorithm) on Wikipedia
