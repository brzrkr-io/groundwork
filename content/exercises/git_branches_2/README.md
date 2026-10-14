# Branches: resolve a merge conflict

Two people changed the same lines of the same config file on different
branches. Git can't know which version is right, so it stops and asks you.
That's a *merge conflict*: routine, not scary, once you've fixed a few.

**New here:** merge commits, `--no-edit`, conflict markers, `git merge --abort`, `|| true`.

## Learn

### Recap

- `bash setup.sh` rebuilds `sandbox/`; your scripts run inside it, in order
  (`cd sandbox && bash ../1_merge_metrics.sh`). The tests do the same.
- `git merge <branch>` brings `<branch>` into the current branch. If the
  current branch hasn't moved, that's a fast-forward, as in *Branches: create, switch and merge*.

### Merge commits

When **both** branches have new commits, Git can't just slide a pointer.
It makes a *merge commit*: a commit with two parents that combines both
sides.

```text
before:  A---B---D  main
              \
               C  fix/tls

after:   A---B---D---M  main
              \     /
               C----   fix/tls
```

A merge commit needs a message, so `git merge` opens an editor.
`git merge --no-edit fix/tls` accepts the default message,
`Merge branch 'fix/tls'`, instead. `git log --graph --oneline --all` draws
the picture above.

### Conflicts

If both sides changed the same lines (or lines right next to each other),
Git stops halfway and writes both versions into the file:

```text
port: 8080
<<<<<<< HEAD
workers: 4
=======
workers: 8
>>>>>>> tune-workers
```

Above `=======` is your side (`HEAD`), below is the other branch's.
`git status` lists the file under "Unmerged paths".

To resolve it:

1. Edit the file to what it **should** be, deleting all three marker lines.
   Often that's a mix of both sides.
2. `git add <file>` marks it resolved.
3. `git commit --no-edit` makes the merge commit with the default message.

Changed your mind? `git merge --abort` puts everything back as it was
before the merge.

### Exit codes and `set -e`

`git merge` exits with status `1` when it stops on a conflict. In a script
with `set -e`, that ends the script. When a failure is expected, `cmd || true`
turns it into success: if `cmd` fails, `true` runs and succeeds.

> **Why is it written like this?** **Git's choice:** a non-zero exit code
> tells scripts and CI "this needs a human". **Your choice:** `|| true` says
> "I know, carry on". Use it only where you really mean that.

### Writing a whole file from a script

A *heredoc* (from *Linux and the shell*) writes several lines at once. The
quoted `'EOF'` keeps bash from expanding anything inside:

```bash
cat > config/db.yaml <<'EOF'
host: db.internal
pool: 20
EOF
```

## Your task

Two branches are waiting to merge into `main`. `add-metrics` merges
cleanly. `raise-limits` changed `config/app.yaml` right where `main` did.
Run `bash setup.sh`, then `cd sandbox` and `git log --graph --oneline --all`.

### 1. Merge a branch with a merge commit

**Goal:** in `1_merge_metrics.sh`, merge `add-metrics` into `main` with the
default message, without opening an editor.

```text
$ git log --oneline -1
4e5f6a7 Merge branch 'add-metrics'
```

Your hash will differ.

The newest commit on `main` must have two parents and contain
`config/metrics.yaml`.

**Think it through:**
- Which branch are you on after `setup.sh`?
- Which option accepts the default message?

### 2. Start the conflicting merge

**Goal:** in `2_start_conflicting_merge.sh`, merge `raise-limits` into
`main`, and make the script succeed even though the merge stops on a conflict.

Afterwards the merge is still in progress and `config/app.yaml` is the
conflicted file:

```text
$ git status --short
UU config/app.yaml
```

**Think it through:**
- What exit code does `git merge` give on a conflict, and what does `set -e`
  do with it?
- What can follow `||` to turn a failure into success?

### 3. Resolve the conflict

**Goal:** in `3_resolve_conflict.sh`, write the resolved `config/app.yaml`,
mark it resolved, and finish the merge with the default message.

The resolved file keeps `main`'s timeout and the branch's connection limit:

```yaml
name: web
replicas: 2
timeout: 30s
max_connections: 200
log_level: info
```

Afterwards the newest commit is `Merge branch 'raise-limits'` with two
parents, and nothing is left uncommitted.

**Think it through:**
- Which tool writes a whole multi-line file in one go?
- Which two commands come after fixing the file?

## Learn more

- [Pro Git: Basic Merge Conflicts](https://git-scm.com/book/en/v2/Git-Branching-Basic-Branching-and-Merging#_basic_merge_conflicts)
- [git-merge documentation: How conflicts are presented](https://git-scm.com/docs/git-merge#_how_conflicts_are_presented)
- [GitHub Docs: Resolving a merge conflict using the command line](https://docs.github.com/en/pull-requests/collaborating-with-pull-requests/addressing-merge-conflicts/resolving-a-merge-conflict-using-the-command-line)
