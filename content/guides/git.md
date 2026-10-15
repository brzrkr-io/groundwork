# How Git works

**In one sentence:** Git records your project as a chain of snapshots, called commits, each pointing back to the one before, and lets you name, compare, combine and undo them.

**Why it matters:** every config file, Terraform module, Kubernetes manifest and CI pipeline you'll touch lives in Git. When production breaks at 3 a.m., the first questions are "what changed, who changed it, and how do we undo it?", and Git answers all three in seconds if you know how it thinks.

**You'll learn:**

- Why teams put everything under version control
- Git's mental model: snapshots, hashes, a graph of commits, and three places a file can live
- What branches and `HEAD` really are (just pointers), and what merge and rebase do to them
- How remotes, fetch, pull and push fit together
- The safe way to undo each kind of mistake, and why the reflog means you rarely lose work
- How to investigate history with log, diff, blame and bisect, and how it all maps to pull requests and GitOps

## Why version control

Someone edits `nginx.conf` on Friday afternoon, the site goes down, and nobody knows what the file looked like before. Soon there's `nginx.conf.bak`, `nginx.conf.old` and `nginx.conf.friday`. That's version control done badly.

*Version control* records every change to a set of files: what, who, when and why. You can see every past version, undo in one command, work in parallel without overwriting each other, and review changes before they go live.

*Git* is the version control system nearly everyone uses. It's *distributed*: every copy of a repository holds the full history, so you can commit and dig through history with no network at all.

## Snapshots, not diffs

You might picture Git storing a list of changes: "line 4 changed, line 9 added". It doesn't. Each *commit* is a **snapshot** of every tracked file, as it was at that moment, plus a little metadata:

```text
commit 7f3a9c1
  tree:    the full set of files and folders
  parent:  e2b8d40   (the commit before this one)
  author:  Priya Shah <priya@example.com>, 2026-09-14 10:02
  message: Raise nginx worker count to 8
```

When you ask "what did this commit change?", Git compares its snapshot with its parent's and computes the diff on the spot. Snapshots aren't wasteful: a file that didn't change is stored once and shared by every later snapshot.

### Hashes name everything

Everything Git stores (file contents, folder listings, commits) is named by a *hash*: a 40-character fingerprint computed from its contents, like `7f3a9c1e5b...`. You usually see the first 7 characters.

- **Same content, same hash.** Change one byte, the date or the parent, and the hash changes completely. A commit's hash covers its whole history.
- **Commits never change.** "Editing" a commit (amend, rebase) really makes a new commit with a new hash. The old one lingers until Git cleans it up.

## Commits form a graph

Each commit points to its *parent*. Follow the arrows back and you walk the history:

```text
A <-- B <-- C <-- D
```

The first commit has no parent; a *merge commit* has two. So history is a graph that can fork and join, but never loops, because a commit can only point at older commits. That's a *DAG*: a directed acyclic graph.

```text
A---B---C-------F     main
         \     /
          D---E       feature/tls
```

Here `D` and `E` were made on a branch, and `F` merges them back: `F` has two parents, `C` and `E`. `git log --graph --oneline --all` draws your repo like this.

## The three places a file lives

Before something becomes a commit, it passes through three places:

| Place | What it holds | You move things here with |
|---|---|---|
| *working tree* | the files you see and edit | your editor |
| *staging area* (the *index*) | exactly what the next commit will contain | `git add` |
| *repository* (`.git/`) | every commit ever made | `git commit` |

```text
working tree --git add--> staging area --git commit--> repository
working tree <--git restore-- staging area <--git restore --staged-- last commit
```

Why a middle step? A real session leaves a fix you want to ship next to a debug line you don't. Staging lets you choose exactly what goes in. `git status` shows each place; `git diff` compares working tree with staging, `git diff --staged` staging with the last commit.

You practise this in *Staging: choose what goes in a commit*.

## Branches and HEAD are pointers

A *branch* is a name that points at one commit: a tiny file in `.git/refs/heads/` holding a hash. Creating one is instant, because nothing is copied.

*HEAD* answers "where am I?". It normally points at a branch, and the branch points at a commit:

```text
              HEAD
               |
               v
              main
               |
               v
A---B---C---D
```

When you commit, the new commit's parent is the current one, and the branch `HEAD` points at moves forward. Other branches stay put. `git switch -c fix/timeout` creates a branch and moves `HEAD` to it; `git branch -d fix/timeout` deletes only the pointer.

If `HEAD` points straight at a commit, you're in *detached HEAD* state, as during a bisect. Fine for looking around; create a branch before committing anything you want to keep.

A *tag* is a pointer too, but one that never moves: `v1.4.0` names one release commit for good. Tags are how you say "this exact commit is what we shipped", so rollbacks, bisects and release notes start from them.

## Merging

`git merge fix/timeout` brings the commits from `fix/timeout` into the branch you're on. There are two cases.

### Fast-forward

If your branch hasn't moved since the other one started, there's nothing to combine. Git slides your branch's pointer forward:

```text
before:  A---B  main
              \
               C---D  fix/timeout

after:   A---B---C---D  main, fix/timeout
```

No new commit is made.

### Merge commit

If both branches have new commits, Git combines both sets of changes since their *merge base* (the last commit they share) in a merge commit with two parents:

```text
before:  A---B---E  main
              \
               C---D  fix/timeout

after:   A---B---E-------M  main
              \         /
               C---D----   fix/timeout
```

### Conflicts

If both sides changed the same lines, Git can't know which is right. It stops, marks the file, and leaves the decision to you:

```text
<<<<<<< HEAD
timeout: 30s
=======
timeout: 15s
>>>>>>> fix/timeout
```

Edit the file to what it should be (no markers), `git add` it, and `git commit`; or `git merge --abort` to back out. A conflict isn't an error; it's Git refusing to guess. You resolve one in *Branches: resolve a merge conflict*.

## Rebasing, and when not to

A *rebase* is the other way to bring a branch up to date. Instead of a merge commit, `git rebase main` (run on your branch) replays each of your commits on top of the newest `main`:

```text
before:  A---B---E  main
              \
               C---D  fix/timeout

after:   A---B---E  main
                  \
                   C'---D'  fix/timeout
```

`C'` and `D'` have the same changes as `C` and `D`, but new parents, so new hashes. The straight line is easy to read and bisect. But if a teammate already built on `C`, your rebase pulls the floor out from under them. Hence the golden rule:

> **Only rebase commits that nobody else has.** Rebase your own feature branch before review; never rebase `main` or any branch others pull from.

| | Merge | Rebase |
|---|---|---|
| History | what really happened, with merge commits | a tidy straight line, rewritten |
| Safe on shared branches | yes | no |
| Typical use | landing a pull request on `main` | catching up your own branch before review |

You try it in *Rebase: catch a branch up with main*.

## Remotes: fetch, pull and push

A *remote* is another copy of the repository, usually on GitHub or GitLab, named `origin` by default. They sync only when you ask.

| Command | What it does |
|---|---|
| `git clone <url>` | copy a whole repository, history and all |
| `git fetch` | download new commits from the remote; update `origin/main` and friends, but **not** your branches |
| `git pull` | `fetch`, then merge (or rebase) the remote's copy of your current branch into it, e.g. `origin/main` into `main` |
| `git push` | upload your commits and move the remote's branch to match |

`origin/main` is a *remote-tracking branch*: your memory of where the remote's `main` was when you last fetched.

`git push` refuses if the remote has commits you don't: fetch, combine (`git pull --rebase` replays your unpushed commits on top), push again. `git push --force` skips that check, which is how teammates' work gets destroyed. After rebasing your own branch, `git push --force-with-lease` at least refuses if someone else pushed meanwhile.

The remote doesn't have to be a website: a *bare* repository (history only, no working files) in a folder on your disk works the same way. You clone one, push a branch and play a teammate in *Remotes: clone, push and fetch*, then fix both kinds of rejected push in *Remotes: fix a rejected push*.

## Undoing things

Which undo to use depends on two questions: **is the change committed?** and **has anyone else got it?**

| Situation | Command | What happens |
|---|---|---|
| unstage a file | `git restore --staged <file>` | staging area gets the last commit's version; your edit stays |
| throw away an uncommitted edit | `git restore <file>` | working file gets the staged version; **the edit is gone for good** |
| fix the last commit (not pushed) | `git commit --amend` | the last commit is replaced with a new one |
| undo commits, keep changes staged | `git reset --soft <commit>` | the branch moves back; the changes stay staged |
| undo commits, keep changes unstaged | `git reset <commit>` (`--mixed`) | the branch moves back; the changes stay in your files |
| throw local commits away | `git reset --hard <commit>` | the branch, staging area and files all move back |
| undo a commit that's already shared | `git revert <commit>` | a **new** commit applies the opposite change |

`git revert` is the shared-branch undo: history only grows, so nobody else's copy breaks. That's how you roll back a bad deploy on `main`, as in *Undoing: revert a bad deploy*.

### The reflog is your safety net

`git reflog` lists every place `HEAD` has been: each commit, switch, reset, merge and rebase, newest first.

```text
$ git reflog
4c5d6e7 HEAD@{0}: reset: moving to HEAD~2
9a8b7c6 HEAD@{1}: commit: Add cert rotation
```

Commits dropped by `reset --hard`, a bad rebase or a deleted branch still exist for weeks, and the reflog has their hashes: `git branch rescue 'HEAD@{1}'` gets one back. To copy just that one commit onto your current branch instead, *cherry-pick* it: `git cherry-pick <hash>` makes a new commit with the same change and message.

It can't save what was never committed, so **commit early**: a messy commit can be tidied later. You rescue lost work in *Undoing: rescue lost work with reflog*.

## Investigating history

During an incident, history is evidence.

| Question | Command |
|---|---|
| What happened recently? | `git log --oneline --graph` |
| What exactly did this commit change? | `git show <commit>` |
| What changed between two releases? | `git diff v1.4.0 v1.5.0` (add `--stat` or `--name-only` for a summary) |
| When did this text appear or disappear? | `git log -S 'max_connections: 500'` |
| Who last changed this line, and in which commit? | `git blame -L 40,40 nginx.conf` |
| Which commit broke it? | `git bisect` |

### Bisect: binary search on history

The check passed at the last release and fails now, 200 commits later. *Bisecting* tests the commit halfway: if it's bad, the culprit is in the older half, otherwise the newer half. Each test halves the suspects, so 200 commits take about 8 tests.

```bash
git bisect start main v1.4.0      # bad first, then good
git bisect run ./smoke-test.sh    # exit 0 = good, 1-127 = bad (125 = skip); Git does the rest
git bisect reset                  # back to where you started
```

That's *binary search*, from the *Algorithms and Big-O* guide, applied to history. Small commits that each pass the tests make it precise. You use it in *Bisect: find the commit that broke the check*.

## Good commits

History is only useful if it's readable. Two habits make it so.

**One change per commit.** "Raise timeout to 30s" is easy to review, revert and bisect; "Fix stuff" touching twelve files isn't. **A message that explains:**

```text
Raise api timeout to 30s

The payments provider now takes up to 20s under load, so the 10s
timeout failed about 2% of checkouts on Monday (INC-4821).
```

A short subject (about 50 characters) in the *imperative* ("Raise", not "Raised"; read it as "this commit will raise..."), a blank line, then a body explaining **why**. The diff already shows what.

### `.gitignore`

Build output (`bin/`), logs (`*.log`) and above all secrets (`.env`, keys) must never be committed. A committed `.gitignore` file lists patterns Git leaves untracked.

It only affects untracked files. A secret already committed stays in every old commit even after `git rm --cached`, so treat it as leaked: rotate it first. You practise this in *.gitignore: keep build output and secrets out*.

## From Git to pull requests and GitOps

### Pull requests

A *pull request* (GitLab: *merge request*) is GitHub's and GitLab's layer on top of Git: "please merge my branch into `main`". You push a branch, open the pull request, CI tests it, teammates review the diff, you push fixes, and someone merges it (as a merge commit, a squash or a rebase). Protected branches make that the only way into `main`. The *Code review and working on a team* guide covers the human side.

### GitOps

*GitOps* goes further: **the Git repository is the source of truth for what should be running.** Kubernetes manifests or Terraform live in Git, and an agent such as Argo CD or Flux keeps the cluster matching it, as in the *Declarative systems and control loops* guide.

- **Deploy** = merge a pull request that changes the image tag.
- **Roll back** = `git revert` that commit.
- **Audit** = `git log`: who changed production, when, why, and who approved it.

Small commits, clear messages and revert over reset become how you operate production.

## Common mistakes

- **Committing secrets.** Add `.env` and keys to `.gitignore` before the first commit. If one slips through and is pushed, rotate it: deleting it later doesn't unleak it.
- **Rewriting shared history.** Amending, rebasing or `reset --hard` on commits others have, followed by `push --force`, breaks everyone's copy. Use `git revert` on shared branches.
- **Reaching for `reset --hard` or `git restore` too fast.** They throw away uncommitted work for good. Commit (even "WIP") or run `git status` first.
- **`git add .` without looking.** It sweeps in debug lines, logs and secrets. Read `git diff --staged` before every commit.
- **Panicking after a "lost" commit.** It's almost always in `git reflog`.
- **Giant commits with vague messages.** They're hard to review, revert and bisect. One change per commit; say why.
- **Resolving a conflict by deleting the other side.** Read both versions; the right answer is often a mix of the two.

## Check yourself

1. Git stores each commit as a snapshot of all files, not as a list of changes. How can it then show you "what changed" in a commit?
2. Why does amending a commit, or rebasing a branch, change the commit's hash?
3. What is a Git branch, physically, and why is creating one instant?
4. You run `git merge feature` on `main`. When does Git fast-forward, and when does it make a merge commit?
5. Why should you never rebase a branch, such as `main`, that teammates have already pulled?
6. A bad commit is already on the shared `main` branch and deployed. Should you use `git reset --hard` or `git revert` to undo it, and why?
7. What do `git reset --soft HEAD~1`, `git reset HEAD~1` and `git reset --hard HEAD~1` each do with the undone commit's changes?
8. You ran `git reset --hard HEAD~3` and lost a commit you needed. How do you get it back?
9. What is the difference between `git fetch` and `git pull`?
10. A check passed at the last release and fails today, 500 commits later. Roughly how many tests does `git bisect` need to find the first bad commit, and why?
11. A teammate committed a `.env` file with a database password and pushed it. Is adding `.env` to `.gitignore` and running `git rm --cached .env` enough?
12. You edited `config/app.yaml` and `deploy.sh`, but only the config change should go into the next commit. Which commands do you run, and how do you check what the commit will contain before making it?
13. A merge stops with `<<<<<<<`, `=======` and `>>>>>>>` lines in a config file. What do the two parts mean, and which three steps finish the merge?

## Answers

1. It compares the commit's snapshot with its parent's snapshot and computes the diff on demand. Unchanged files are stored once and shared between snapshots, so snapshots stay cheap.
2. A commit's hash is computed from everything in it: the files, the message, the author, the date and the parent's hash. Amend changes the content or message; rebase changes the parent. Either way it's a new commit with a new hash, and the old one is simply no longer on the branch.
3. A tiny file in `.git/refs/heads/` that holds one commit hash: a pointer. Creating one writes that file; nothing is copied.
4. Fast-forward when `main` has no commits that `feature` lacks (`main` hasn't moved since `feature` started): Git just moves `main`'s pointer forward. If both have new commits, it makes a merge commit with two parents.
5. Rebasing replaces the commits with new ones (new hashes). Teammates' work is built on the old commits, so their next pull duplicates commits or conflicts. Only rebase commits nobody else has.
6. `git revert`. It adds a new commit that undoes the bad one, so history only grows and everyone's copy stays valid. `reset --hard` rewrites the shared branch and needs a force-push that breaks teammates' copies.
7. `--soft` moves the branch back and keeps the changes staged. The default `--mixed` keeps them in your files, unstaged. `--hard` throws them away, from the staging area and the working tree.
8. Run `git reflog`, find the line for that commit (or for where `HEAD` was before the reset), and point a branch at it: `git branch rescue 'HEAD@{1}'`, or `git reset --hard <hash>` to put your branch back.
9. `git fetch` downloads new commits and updates remote-tracking branches like `origin/main`, without touching your branches. `git pull` does a fetch and then merges (or rebases) the remote branch into your current branch.
10. About 9, because each test halves the remaining suspects, and 2 to the power 9 is 512. That's binary search.
11. No. It stops tracking the file from the next commit on, but the password is still in the old commits, and everyone who cloned or forked the repo has it. Rotate the password first, then clean up.
12. `git add config/app.yaml`, then `git diff --staged` to read exactly what's staged, then `git commit -m "..."` without `-a`. `deploy.sh` stays changed but unstaged, so it's left out.
13. Between `<<<<<<<` and `=======` is your branch's version (`HEAD`); between `=======` and `>>>>>>>` is the other branch's. Edit the file to the right final version with no marker lines, `git add` it to mark it resolved, then `git commit` to make the merge commit.

## Key terms

| Term | Meaning |
|---|---|
| version control | a system that records every change to a set of files: what, who, when and why |
| repository | a project folder plus its full history, stored in `.git/` |
| commit | a snapshot of all tracked files, with a message, an author, a date and a parent |
| hash | a fingerprint of an object's contents that Git uses as its name, like `7f3a9c1` |
| parent | the commit a commit was made on top of; a merge commit has two |
| DAG | directed acyclic graph: the shape of Git history, where arrows point back and never loop |
| working tree | the files you see and edit |
| staging area (index) | what the next commit will contain; filled with `git add` |
| branch | a movable name that points at one commit |
| tag | a fixed name for one commit, such as the release `v1.4.0`; unlike a branch, it never moves |
| HEAD | the pointer to where you are: usually a branch, sometimes a commit |
| detached HEAD | `HEAD` pointing straight at a commit instead of a branch |
| fast-forward | a merge that only moves a branch pointer forward, making no new commit |
| merge commit | a commit with two parents that joins two lines of history |
| merge conflict | both sides changed the same lines; Git stops and asks you to choose |
| rebase | replaying commits on top of another commit, making new commits |
| remote | another copy of the repository, such as `origin` on GitHub |
| remote-tracking branch | your local record of a remote's branch, like `origin/main` |
| bare repository | a repository with history but no working files, like the copy on a Git server |
| revert | a new commit that undoes an earlier one |
| cherry-pick | copying one commit's change onto the current branch as a new commit |
| reset | moving the current branch to another commit (soft, mixed or hard) |
| reflog | the local log of everywhere `HEAD` has been; the way back to "lost" commits |
| blame | showing, per line, the commit that last changed it |
| bisect | binary search through history to find the first bad commit |
| `.gitignore` | a committed file of patterns for files Git should leave untracked |
| pull request | a request, on GitHub or GitLab, to merge a branch, with review and CI |
| GitOps | running infrastructure from a Git repo, with agents keeping the live system matching it |

## Watch

- [How Git Works: Explained in 4 Minutes](https://www.youtube.com/watch?v=e9lnsKot_SQ) · ByteByteGo · 4 min: the working tree, staging area, local and remote repositories in one animated diagram; a quick picture of *The three places a file lives* and *Remotes: fetch, pull and push*.
- [Git MERGE vs REBASE: Everything You Need to Know](https://www.youtube.com/watch?v=0chZFIZLR_0) · ByteByteGo · 5 min: a short animated comparison of merging and rebasing, to go with *Rebasing, and when not to*.
- [Git For Ages 4 And Up](https://www.youtube.com/watch?v=3m7BgIvC-uQ) · Linux.conf.au 2013 -- Canberra, Australia · 100 min: the classic; Michael Schwern builds the commit graph out of Tinkertoys and shows what commit, branch, merge and rebase do to it. Long, so watch it in parts.
- [So You Think You Know Git - FOSDEM 2024](https://www.youtube.com/watch?v=aolI_Rz0ZqY) · GitButler · 47 min: GitHub co-founder Scott Chacon on Git features most people never find. Save it for after this module.

## Go deeper

- [Pro Git, chapter 10.2: Git Objects](https://git-scm.com/book/en/v2/Git-Internals-Git-Objects): the free official book shows blobs, trees and commits being built by hand, the *Snapshots, not diffs* idea made concrete.
- [Pro Git, chapter 3.5: Remote Branches](https://git-scm.com/book/en/v2/Git-Branching-Remote-Branches): `origin/main`, fetch, pull and push with diagrams.
- [Think Like (a) Git](https://think-like-a-git.net/): a short guide for the perplexed that explains Git as a graph and branches as labels.
- [How to Write a Git Commit Message](https://cbea.ms/git-commit/) (Chris Beams): the seven rules behind *Good commits*, with examples.
- [OpenGitOps](https://opengitops.dev/): the four GitOps principles (declarative, versioned and immutable, pulled automatically, continuously reconciled) in one page.
