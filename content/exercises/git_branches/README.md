# Branches: create, switch and merge

Teams don't commit half-finished work straight to `main`, because `main` is
what gets deployed. Each change grows on its own *branch* and joins `main`
when it's ready, usually through a pull request. Here you do the whole
round trip by hand.

**New here:** `git branch`, `git switch`, `git switch -c`, `git merge` (fast-forward), `git branch -d`.

## Learn

### Recap

- `bash setup.sh` rebuilds `sandbox/`; your scripts run inside it, in order
  (`cd sandbox && bash ../1_create_branch.sh`). The tests do the same.
- Each commit points to its *parent*, the commit before it.

### A branch is a pointer

A *branch* is a name that points at one commit. That's all: a tiny file in
`.git/refs/heads/` holding a commit hash. When you commit on a branch, Git
moves that pointer to the new commit.

*HEAD* says which branch you're on. `git branch` lists branches and stars
the current one:

```text
$ git branch
  main
* fix/nginx-timeout
```

### Creating and switching

```bash
git branch fix/nginx-timeout       # create a branch here; stay where you are
git switch fix/nginx-timeout       # move HEAD to it; your files change to match
git switch -c fix/nginx-timeout    # both at once: create and switch
```

Uncommitted changes come with you when you switch, as long as they don't
clash with the other branch.

> **Why is it written like this?** **Convention:** branch names like
> `feature/...` and `fix/...` group work by kind; the `/` is just part of the
> name. `git checkout` and `git checkout -b` are the older commands for the
> same moves; you'll see them in many guides.

### Merging: fast-forward

`git merge <branch>` brings another branch's commits into the current one.
If the current branch hasn't moved since the other branch started, there is
nothing to combine: Git just slides the pointer forward. That's a
*fast-forward*:

```text
before:  A---B  main
              \
               C  fix/nginx-timeout

after:   A---B---C  main, fix/nginx-timeout
```

No new commit is made. `git log --oneline` on `main` now shows `C`.

### Cleaning up

`git branch -d <branch>` deletes a branch that is fully merged into the
branch you're on. It refuses if the branch has commits yours lacks (`-D`
forces it; careful). Deleting a branch only deletes the pointer, never
commits that other branches still reach.

## Your task

You've written `healthcheck.sh` for the `web` service. Ship it the team way:
on a branch, then merged into `main`. Run `bash setup.sh`, then `cd sandbox`,
`git log --oneline` and `git status`.

### 1. Create a feature branch

**Goal:** in `1_create_branch.sh`, create a branch named `feature/healthcheck`
and switch to it.

```text
$ git branch
* feature/healthcheck
  main
```

It must start at the same commit as `main`.

**Think it through:**
- Which `git switch` option creates the branch too?
- Does `healthcheck.sh` (untracked) come with you?

### 2. Commit on the branch

**Goal:** in `2_commit_on_branch.sh`, commit `healthcheck.sh` on
`feature/healthcheck` with the message `Add health check script`.

Afterwards `feature/healthcheck` has three commits and `main` still has two.

**Think it through:**
- Which branch does a commit move: the current one or all of them?
- Same two steps as any commit: which?

### 3. Merge the branch into main

**Goal:** in `3_merge_back.sh`, switch to `main`, merge `feature/healthcheck`
into it, and delete `feature/healthcheck`.

```text
$ git log --oneline
9b8a7c6 Add health check script
dccae83 Add app config
35309c9 Add deploy script
```

Your hash for the new commit will differ.

It must be a fast-forward: no merge commit. Afterwards you're on `main` and
`git branch` lists only `main`.

**Think it through:**
- `git merge` brings the named branch into which branch?
- Why can't Git delete the branch you're standing on?

## Learn more

- [Pro Git: Branches in a Nutshell](https://git-scm.com/book/en/v2/Git-Branching-Branches-in-a-Nutshell)
- [Pro Git: Basic Branching and Merging](https://git-scm.com/book/en/v2/Git-Branching-Basic-Branching-and-Merging)
- [Learn Git Branching](https://learngitbranching.js.org/) (interactive, in the browser)
