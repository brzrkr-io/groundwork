# Release pipelines: fix the risky deploy

A pipeline that can deploy to production can also leak a token, hand out
too much power, or ship the wrong thing. Here you fix four such risks, each
found in a real-looking post-incident review.

**New here:** your own secrets (`secrets.NAME`), `write-all`, conditions on
`github.event_name` and `github.ref`, tags that move.

## Learn

### Secrets

A *secret* is a value you store in the repository's settings (Settings,
then Secrets and variables, then Actions), never in a file. A workflow
reads it with an expression:

```yaml
env:
  SLACK_WEBHOOK: ${{ secrets.SLACK_WEBHOOK }}
```

`env:` turns it into an environment variable, for the workflow, one job or
one step. The narrower, the fewer steps can see it.

- GitHub replaces a secret's value with `***` when it shows up in a log,
  but only an exact match. Encoded, split or changed, it slips through.
  Never print a secret, even "just to check".
- A token that was ever committed stays in the Git history and in every
  clone. Deleting the line isn't enough: *rotate* it (revoke it and make a
  new one).

This is the same rule as in *Image security: non-root and no secrets*:
secrets come from outside, at run time.

### Permissions from the top down

`permissions:` at the top of a workflow is the default for every job. A
job's own `permissions:` replaces it for that job, as in
*Release pipelines*.

| Top-level setting | Every job's token may... |
|---|---|
| `write-all` | write nearly everything: code, releases, packages, issues |
| `read-all` | read everything, write nothing |
| `contents: read` | read the code, and nothing else |

```yaml
permissions:
  contents: read          # the default for all jobs

jobs:
  triage:
    permissions:          # this job only
      contents: read
      issues: write
```

### Conditions on the event

One workflow often runs for several events. `github.event_name` and
`github.ref` tell them apart:

| What happened | `github.event_name` | `github.ref` |
|---|---|---|
| push to `main` | `push` | `refs/heads/main` |
| push to `release/2.1` | `push` | `refs/heads/release/2.1` |
| pull request #42 | `pull_request` | `refs/pull/42/merge` |
| push of tag `v1.4.2` | `push` | `refs/tags/v1.4.2` |

```yaml
  preview:                # a preview site for every pull request
    if: github.event_name == 'pull_request'
  backup:
    if: github.ref == 'refs/heads/main' || github.ref == 'refs/heads/release/2.1'
```

A pull request from a branch in the same repository runs with the
repository's secrets. A job without a condition runs for it too.

### Tags that move

An image tag is a label, and a push can move it. `:latest` moves on every
push:

```text
10:00  run A builds commit a1, pushes :latest  -> a1
10:01  run B builds commit b2, pushes :latest  -> b2   (label moved)
10:03  run A deploys :latest                   -> production runs b2,
                                                  which run A never tested
```

`github.sha` is the commit's hash: the same in every job of one run, and
different for every commit. An image tagged with it names exactly one
build:

```yaml
tags: ghcr.io/acme/billing:build-${{ github.run_id }}
```

> **Why is it written like this?**
> - **GitHub requires:** `secrets.NAME` matching the secret's name in the
>   settings, and single quotes for text inside an expression.
> - **Convention:** least privilege. `contents: read` at the top, and more
>   only on the job that needs it.
> - **Convention:** deploy by an exact tag (or, stricter, by digest), never
>   by `:latest`.

## Your task

After a bad Friday deploy of `statusd`, the review listed four findings.
Each file holds one of them, with a `TODO` above the risky part. Fix each
so the workflow still does its job, minus the risk.

### 1. Stop leaking the deploy token

**Goal:** in `1_secrets.yml`, read `DEPLOY_TOKEN` from the repository's
secrets, and stop printing it.

- No `ghp_` anywhere in the file (a GitHub token starts with it).
- Every place that sets `DEPLOY_TOKEN` (the job's `env:` now, or a step's)
  sets it to exactly `${{ secrets.DEPLOY_TOKEN }}`.
- No command that mentions `DEPLOY_TOKEN` uses `echo` or `printf`.
- Keep the step that runs `./deploy.sh production`; the script reads
  `DEPLOY_TOKEN` from its environment.

The build log then shows `./deploy.sh production`, and nothing of the
token.

**Think it through:**
- Which secret name goes after `secrets.`?
- In a real repository, what must also happen to the token that was in the
  file?

### 2. Grant only what each job needs

**Goal:** in `2_permissions.yml`, replace `write-all` so `test` can only
read the code and only `image` can push packages.

- The top-level `permissions:` is exactly `contents: read`.
- The `image` job has its own `permissions:`, exactly `contents: read` and
  `packages: write`.
- The `test` job has no `permissions:` of its own (or only
  `contents: read`).

**Think it through:**
- Which job writes something, and what does it write?
- Why does `image` list `contents: read` again?

### 3. Don't deploy pull requests

**Goal:** in `3_pull_requests.yml`, make `deploy` run only for pushes to
`main`.

- Keep both triggers: pull requests still need their tests.
- `test` has no `if:`.
- `deploy` still needs `test`, and gets exactly this condition (with or
  without `${{ }}` around it):

  ```yaml
  if: github.event_name == 'push' && github.ref == 'refs/heads/main'
  ```

| What happened | Jobs that run |
|---|---|
| push to `main` | `test`, then `deploy` |
| push to `release/2.1` | `test` |
| pull request | `test` |

**Think it through:**
- Which row of the event table is the only one that should deploy?
- Where does `if:` go: on the job or on a step?

### 4. Deploy exactly what you tested

**Goal:** in `4_latest.yml`, tag the image with the commit, and deploy
that tag.

- No step uses `:latest`.
- The build step's `tags` is exactly
  `ghcr.io/acme/statusd:${{ github.sha }}`.
- The `deploy` job runs
  `./deploy.sh production ghcr.io/acme/statusd:${{ github.sha }}`.

Commit `3f9c2ab...` then builds `ghcr.io/acme/statusd:3f9c2ab...`, and the
same run deploys that very image, whatever other runs push meanwhile.

**Think it through:**
- What is the same in the build job and the deploy job of one run, but
  different in the next run?

## Learn more

- [Secure use reference](https://docs.github.com/en/actions/reference/security/secure-use)
- [Using secrets in GitHub Actions](https://docs.github.com/en/actions/how-tos/write-workflows/choose-what-workflows-do/use-secrets)
- [Contexts reference](https://docs.github.com/en/actions/reference/workflows-and-actions/contexts)
- [Use GITHUB_TOKEN for authentication in workflows](https://docs.github.com/en/actions/tutorials/authenticate-with-github_token)
