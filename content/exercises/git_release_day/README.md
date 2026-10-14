# Practice: release day

It's release day for the `web` service. `main` has a change that must not
ship, a reviewed feature branch is waiting, and the release needs a version,
a tag and release notes. Everything in this module, in one sitting.

**New here:** `git log --no-merges`; the rest is practice.

## Learn

### Recap

- `bash setup.sh` rebuilds `sandbox/`; your scripts run inside it, in order
  (`cd sandbox && bash ../1_revert_debug_logging.sh`). The tests do the same.
- `git revert --no-edit <commit>` undoes a shared commit with a new one.
- `git merge --no-edit <branch>` merges with the default message;
  `git branch -d` deletes a merged branch.
- `git tag <name>` names the current commit.
- `git log A..B` lists commits in `B` but not in `A`; `--format='%s'`
  prints subjects; `--no-merges` skips merge commits.

## Your task

Run `bash setup.sh`, then `cd sandbox` and `git log --graph --oneline --all`.

### 1. Revert the debug logging

**Goal:** in `1_revert_debug_logging.sh`, revert `Enable debug logging in prod` on `main`, with the default message.

Afterwards `config/app.yaml` on `main` has `log_level: info` again and still
has `request_timeout: 15s`.

**Think it through:**
- Where is the bad commit relative to `HEAD`?
- Why revert and not reset?

### 2. Merge the rate-limit feature

**Goal:** in `2_merge_rate_limit.sh`, merge `feature/rate-limit` into `main`
with a merge commit and the default message, then delete
`feature/rate-limit`.

```text
$ git log --oneline -1
7c8d9e0 Merge branch 'feature/rate-limit'
```

Your hash will differ.

**Think it through:**
- Both branches have new commits. What kind of merge does that make?
- Which delete flag refuses unmerged work?

### 3. Bump the version and tag the release

**Goal:** in `3_tag_release.sh`, set the contents of `VERSION` to `1.4.0`,
commit it with the message `Release v1.4.0`, and tag that commit `v1.4.0`.

`VERSION` holds one line: `1.4.0`. Nothing may be left uncommitted.

**Think it through:**
- Which redirection replaces a file's contents?
- Tag before or after committing?

### 4. Print the release notes

**Goal:** `4_release_notes.sh` prints the subject of every commit after
`v1.3.0` up to and including `v1.4.0`, newest first, as `- <subject>`,
skipping merge commits.

```text
- Release v1.4.0
- Revert "Enable debug logging in prod"
- Add request timeout setting
- Enable debug logging in prod
- Document rate limits
- Add rate limiter config
```

One line per commit, nothing else.

**Think it through:**
- Which range syntax means "after this tag, up to that one"?
- The `- ` is fixed text in the format. Which option drops the merge commit?

## Learn more

- [Pro Git: Tagging](https://git-scm.com/book/en/v2/Git-Basics-Tagging)
- [Semantic Versioning](https://semver.org/)
- [Keep a Changelog](https://keepachangelog.com/en/1.1.0/)
