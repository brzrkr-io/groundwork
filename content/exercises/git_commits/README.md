# Commits: init, add and commit

Every change to a server's config should be recorded: who changed what, when
and why. Git does that for any folder of files. Here you put a folder of
deploy files under version control and make your first commits.

**New here:** `git init`, `git config`, `git status`, `git add`, `git commit`, `git log`.

## Learn

### Your sandbox

Git exercises work on a practice repository, the *sandbox*. Build it from
the exercise folder, then look around:

```bash
bash setup.sh         # build (or rebuild, from scratch) sandbox/
cd sandbox
ls
```

You write your answers as git commands in the numbered scripts (`1_init_repo.sh`
and so on). Each script runs **inside** the sandbox, so try one like this:

```bash
cd sandbox
bash ../1_init_repo.sh
```

Each test does the same: it runs `setup.sh` in a fresh folder, then your
scripts **in order**, from 1 up to the one it checks, and inspects the
result with git. To start over by hand, run `bash setup.sh` again.

> **Why is it written like this?** **Git requires** nothing here: these are
> plain bash scripts with one git command per line, as you'd type them.
> `set -euo pipefail` (from *Linux and the shell*) stops the script at the
> first failing command, so a mistake shows up as an error, not a mess.

### A repository

A *repository* (repo) is a folder whose history Git tracks. `git init` creates
the hidden `.git/` folder where that history lives. Delete `.git/` and the
history is gone; your files stay.

```bash
git init -b trunk      # start a repo whose first branch is called trunk
```

`-b` names the first *branch*. Most teams use `main`; without `-b`, Git
picks its configured default, which may be `master`.

### Who you are

Every commit records an author name and email. `git config` sets them:

```bash
git config user.name "Ada Lovelace"            # this repo only
git config --global user.name "Ada Lovelace"   # every repo on this machine
```

Without `--global`, the setting is saved in `.git/config` and only applies to
this repo.

### The three places a file lives

| Place | What it is | Command that moves files here |
|---|---|---|
| *working tree* | the files you see and edit | (you edit them) |
| *staging area* (the *index*) | what the next commit will contain | `git add <path>` |
| *repository* | the saved commits | `git commit -m "message"` |

`git add .` stages everything in the current folder; `git add config/app.yaml`
stages one file. `git status` shows what is changed, staged or untracked.

### Commits and messages

A *commit* is a snapshot of every staged file plus a message, an author and
a date. `git log --oneline` lists commits, newest first:

```text
3f9c2e1 Rotate the TLS certificate
a41b07d Add nginx config
```

A good message has a short subject (about 50 characters) in the *imperative*:
"Add nginx config", not "added stuff". Read it as "this commit will ...".

## Your task

You've been handed a folder of deploy files for the `web` service. Put it under
version control. Run `bash setup.sh`, then look at `sandbox/`: `deploy.sh`,
`config/app.yaml` and `README.md`.

### 1. Create the repository

**Goal:** in `1_init_repo.sh`, make the sandbox a Git repo on branch `main`,
with name `Sam Rivera` and email `sam@example.com` set for this repo only.

The test checks:

```text
current branch:  main
user.name:       Sam Rivera
user.email:      sam@example.com
```

**Think it through:**
- Which `git init` option names the first branch?
- Which config keys hold the name and the email? Leave out `--global`.

### 2. Make the first commit

**Goal:** in `2_first_commit.sh`, stage all three files and commit them with
the message `Add deploy script and app config`.

```text
$ git log --oneline
4d1a0c2 Add deploy script and app config
```

Your hash will differ. Afterwards, `git status` must show nothing left to
commit.

**Think it through:**
- Staging comes first, then committing. Which command does each?
- How do you stage every file in the folder at once?

### 3. Commit a change

**Goal:** in `3_add_health_check.sh`, add the line `healthcheck: /healthz` to
the end of `config/app.yaml` and commit it with the message `Add health check path`.

`config/app.yaml` must end up as:

```yaml
name: web
replicas: 2
port: 8080
healthcheck: /healthz
```

The new commit must change only `config/app.yaml`, and nothing may be left
uncommitted.

**Think it through:**
- Which bash redirection appends a line to a file instead of replacing it?
- Git only commits what's staged. What must happen between editing and
  committing?

## Learn more

- [Pro Git: Getting a Git Repository](https://git-scm.com/book/en/v2/Git-Basics-Getting-a-Git-Repository)
- [Pro Git: Recording Changes to the Repository](https://git-scm.com/book/en/v2/Git-Basics-Recording-Changes-to-the-Repository)
- [How to Write a Git Commit Message](https://cbea.ms/git-commit/)
