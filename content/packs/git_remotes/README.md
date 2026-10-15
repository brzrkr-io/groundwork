# Remotes: clone, push and fetch

So far every repo lived in one folder on your machine. A team shares one
repository on GitHub or GitLab, and everyone keeps a full copy. Here you
clone a shared repo, push a branch, and play a teammate who fetches and
pulls it, all on your own machine and without an account anywhere.

**New here:** bare repositories, `git clone`, remotes and `origin`,
`git remote -v`, `git push` and `git push -u`, remote-tracking branches like
`origin/main`, `git fetch`, `git pull --ff-only`, `git branch -a`.

## Learn

### Recap

- `bash setup.sh` rebuilds `sandbox/`. This time the sandbox is a folder
  **holding** repositories, and your scripts run in that folder:
  `cd sandbox && bash ../1_clone.sh`. Each script starts with `cd` into the
  repo it works on, as in *Files and folders: paths, mkdir, cp, mv and rm*.
- `git merge --ff-only` moves a branch forward and refuses anything else,
  as in *Rebase: catch a branch up with main*.

### A shared repository is just another copy

A *remote* is another copy of the same repository that yours can sync
with. On a team it lives on GitHub or GitLab, but Git doesn't care where:
a folder on the same disk works the same way, and that's what this
sandbox uses.

```text
sandbox/
├── origin.git/   the shared repo: plays GitHub. A *bare* repo: history, no files to edit
├── alex/         Alex's clone. Alex is a teammate; you'll play Alex too
└── web/          your clone (you make it in task 1)
```

A *bare* repository has only the history (what `.git/` holds) and no
working tree, because nobody edits files on the server. Its name ends in
`.git` by convention.

### Clone: your own copy

`git clone <where> <folder>` copies the whole history into a new folder,
checks out the default branch, and remembers where it came from under the
name `origin`:

```bash
git clone /srv/git/infra.git infra   # any path or URL: https://..., git@...
cd infra
git remote -v
# origin  /srv/git/infra.git (fetch)
# origin  /srv/git/infra.git (push)
```

A clone copies commits, not settings: `git config user.name` and
`user.email` (without `--global`) must be set again in each clone.

### Remote-tracking branches: `origin/main`

Your clone keeps a read-only bookmark for each branch on the remote:
`origin/main`, `origin/fix/dns`. They show where the remote's branches were
**the last time you talked to it**. `git branch -a` lists them next to your
own branches.

```text
your main     ──► C
origin/main   ──► C     (as of your last fetch; origin may have moved since)
```

### Push: send your commits

`git push <remote> <branch>` uploads your branch's new commits and moves
the remote's branch to match. For a new branch, add `-u` (short for
`--set-upstream`): it also records `origin/<branch>` as the branch's
*upstream*, so from then on a plain `git push` or `git pull` on that branch
knows where to go.

```bash
git switch -c fix/dns
# ... edit, add, commit ...
git push -u origin fix/dns   # first push of a new branch
git push                     # later pushes from the same branch
```

A clone's `main` already has `origin/main` as its upstream, so `git push`
alone works there.

> **Why is it written like this?** **Git requires** you to name the remote
> and branch the first time: it never guesses where to publish. **Convention:**
> the remote is called `origin`, and your branch keeps the same name on the
> remote.

### Fetch and pull: get other people's commits

| Command | Downloads new commits | Moves `origin/*` | Changes your branches and files |
|---|---|---|---|
| `git fetch` | yes | yes | **no** |
| `git pull` | yes | yes | yes: fetch, then merge the upstream into your branch |

`git fetch` is always safe: look first, then decide. After fetching, a
range shows what arrived:

```bash
git fetch
git log --oneline main..origin/main   # on origin, not yet in my main
```

`git pull --ff-only` is a pull that only fast-forwards. If your branch has
commits of its own, it stops instead of making a merge you didn't plan
(*Remotes: fix a rejected push* handles that case).

> **Why is it written like this?** **Git requires** nothing extra for a
> fast-forward. **Convention:** `--ff-only` on pulls you expect to be
> simple, so a surprise stops you instead of quietly creating a merge
> commit.

## Your task

Alex has pushed the web service to the team repo, `origin.git`. You'll
clone it, push a fix on a branch, then switch hats: Alex fetches your
branch to review it, and finally pulls the merged result. Run
`bash setup.sh`, then `cd sandbox`, `ls`, and `git -C alex log --oneline`
(`-C` runs git as if inside that folder).

Every script runs in `sandbox/`, so each one starts with `cd web` or
`cd alex`.

### 1. Clone the repository

**Goal:** in `1_clone.sh`, clone `origin.git` into a folder named `web`.
Then, inside `web`, set the name `Sam Rivera` and the email
`sam@example.com` for that repository only.

```text
$ cd web && git remote -v
origin  /home/sam/.../sandbox/origin.git (fetch)
origin  /home/sam/.../sandbox/origin.git (push)
$ git log --oneline
045c8c3 Add README
...
```

`web`'s `main` is the same commit as origin's `main`.

**Think it through:**
- Which two arguments does `git clone` need here?
- Where must you be when you run `git config` without `--global`?

### 2. Push a branch

**Goal:** in `2_push_branch.sh`, inside `web`, create and switch to the
branch `fix/timeout`, add the line `request_timeout: 30s` to the end of
`config/app.yaml`, commit it with the message `Add a 30s request timeout`,
and push the branch to `origin` with its upstream set.

Afterwards `origin.git` has a branch `fix/timeout` with your one commit on
top of `main`, and `web` has nothing left to commit:

```text
$ git branch -vv
* fix/timeout 9d8e7f6 [origin/fix/timeout] Add a 30s request timeout
  main        045c8c3 [origin/main] Add README
```

Your hash will differ.

**Think it through:**
- Which redirection adds a line to the end of a file?
- Which push option records the upstream?

### 3. Fetch as Alex

**Goal:** in `3_review_as_alex.sh`, play Alex: inside `alex`, fetch from
origin, then print the subject of each commit on `origin/fix/timeout` that
Alex's `main` doesn't have, newest first, one per line.

```text
$ bash ../3_review_as_alex.sh
Add a 30s request timeout
```

Print nothing else. Alex's `main` and files must stay as they were.

**Think it through:**
- Which command brings in new commits without touching Alex's branches?
- Which range means "on `origin/fix/timeout`, not on `main`"? Which
  `--format` placeholder is the subject?

### 4. Land the fix and pull it

**Goal:** in `4_land_and_pull.sh`, inside `web`, switch to `main`,
fast-forward it to `fix/timeout` (refusing any other merge) and push
`main`. Then, inside `alex`, pull `main`, again only as a fast-forward.

Afterwards `main` is the same commit in all three repos, and
`alex/config/app.yaml` ends with `request_timeout: 30s`:

```text
$ git -C origin.git log --oneline -2 main
9d8e7f6 Add a 30s request timeout
045c8c3 Add README
```

The newest hash will differ.

**Think it through:**
- `main` already has an upstream. Does `git push` need any arguments?
- From inside `web`, what's the relative path to `alex`?

## Learn more

- [Pro Git: Working with Remotes](https://git-scm.com/book/en/v2/Git-Basics-Working-with-Remotes)
- [Pro Git: Remote Branches](https://git-scm.com/book/en/v2/Git-Branching-Remote-Branches)
- [Pro Git: Getting Git on a Server](https://git-scm.com/book/en/v2/Git-on-the-Server-Getting-Git-on-a-Server) (bare repositories)
