# .gitignore: keep build output and secrets out

Some files must never be committed: compiled binaries (rebuilt every time),
logs (noise), and above all secrets. One leaked password in a Git repo is a
real incident. A `.gitignore` file tells Git which files to leave alone.

**New here:** `.gitignore` patterns, `!` exceptions, `git check-ignore`, `git rm --cached`.

## Learn

### Recap

- `bash setup.sh` rebuilds `sandbox/`; your scripts run inside it, in order
  (`cd sandbox && bash ../1_ignore_build_output.sh`). The tests do the same.
- `git status --short` marks untracked files with `??`.

### `.gitignore`

`.gitignore` is a plain text file in the repo, one *pattern* per line. You
commit it, so the whole team shares the rules. Untracked files that match a
pattern vanish from `git status`, and `git add .` skips them.

| Pattern | Ignores |
|---|---|
| `tmp/` | every folder named `tmp`, and everything in it |
| `*.tmp` | every file ending in `.tmp`, in any folder |
| `/notes.txt` | `notes.txt` at the top of the repo only |
| `secrets.yaml` | every file named `secrets.yaml`, in any folder |
| `!keep.tmp` | an exception: don't ignore `keep.tmp` after all |
| `# text` | a comment |

Order matters: later lines win, so an exception goes **after** the pattern it
excuses.

You can write the file from a script like any other:

```bash
printf '%s\n' 'tmp/' '*.tmp' > .gitignore    # one argument per line, replace the file
echo '*.bak' >> .gitignore                   # add one more line
```

> **Why is it written like this?** **Git requires** the name `.gitignore`
> exactly, with the dot: Git looks for it by name. Single quotes stop bash
> from expanding `*.tmp` into matching file names before `printf` sees it.

### Checking a rule

`git check-ignore -v <path>` tells you whether a path is ignored, and by
which line:

```text
$ git check-ignore -v build/app.tmp
.gitignore:2:*.tmp	build/app.tmp
```

No output (and exit code 1) means "not ignored".

### Ignoring doesn't untrack

`.gitignore` only affects *untracked* files. A file that's already committed
stays tracked, ignored or not. To stop tracking it, remove it from the
staging area only:

```bash
git rm --cached secrets.yaml    # untrack; the file stays on your disk
```

Commit that, and the file is gone from the *next* commit. It is still in
every older commit, so anyone with the repo can read it.

> **Why is it written like this?** A secret that ever reached a shared repo
> is leaked. Rotate it (change the password), then clean up. Rewriting
> history is a last resort and doesn't help copies people already have.

## Your task

The `web` repo has build output and logs lying around, a committed `.env`
with a password in it, and a production env file waiting to be committed by
accident. Run `bash setup.sh`, then `cd sandbox`, `git status --short` and
`ls -a`.

### 1. Ignore build output and logs

**Goal:** in `1_ignore_build_output.sh`, create `.gitignore` with patterns
for the `bin/` folder and every `.log` file, and commit it with the message
`Ignore build output and logs`.

The test checks these with `git check-ignore`:

| Path | Ignored? |
|---|---|
| `bin/web` | yes |
| `logs/deploy.log` | yes |
| `debug.log` | yes |
| `build.sh` | no |

The commit must contain `.gitignore`.

**Think it through:**
- Which pattern matches a folder, and which matches a file ending?
- `.gitignore` is a new file. What does Git need before it can commit it?

### 2. Stop tracking the committed secret

**Goal:** in `2_untrack_secret.sh`, untrack `.env` (keep it on disk), add a
`.env` line to `.gitignore`, and commit both with the message `Stop tracking .env`.

Afterwards `.env` is not in the newest commit, still exists in the sandbox,
and is ignored. Keep task 1's patterns: add to `.gitignore`, don't replace it.

**Think it through:**
- Which `git rm` option removes a file from the repo but not from disk?
- Which redirection adds a line without wiping the file?

### 3. Ignore every env file except the example

**Goal:** in `3_ignore_env_files.sh`, add patterns so every file ending in
`.env` is ignored except `example.env`, and commit with the message
`Ignore env files except the example`.

| Path | Ignored? |
|---|---|
| `config/prod.env` | yes |
| `config/example.env` | no (and it stays committed) |

Afterwards `git status --short` prints nothing: everything left over is
ignored.

**Think it through:**
- One pattern for "ends in `.env`", one exception. Which comes first?
- Which character starts an exception?

## Learn more

- [Pro Git: Ignoring Files](https://git-scm.com/book/en/v2/Git-Basics-Recording-Changes-to-the-Repository#_ignoring)
- [gitignore documentation](https://git-scm.com/docs/gitignore)
- [GitHub's .gitignore templates](https://github.com/github/gitignore)
