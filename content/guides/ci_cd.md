# Continuous integration and delivery

**In one sentence:** *continuous integration* means merging changes often, each one built and tested automatically; *continuous delivery* means every change that passes is packaged once and can ship to production safely at the push of a button.

**Why it matters:** CI/CD is the conveyor belt a platform team builds and looks after. When main is red, nobody can ship. When a deploy goes wrong at 5pm, the pipeline decides whether rolling back takes one minute or one evening. When a token leaks from a build log, its permissions decide how bad that is.

**You'll learn:**
- what continuous integration is, and why small, frequent merges beat big ones
- how a pipeline is built from stages, jobs and steps
- how to read a GitHub Actions workflow: triggers, jobs, steps, `needs:`, `if:`, matrices, caches, artifacts
- why you build once and promote the same image from dev to staging to production
- rolling, blue/green and canary deploys, rollbacks, and GitOps
- how to keep secrets and the build itself safe: least privilege, OIDC, pinned actions, signed images

## Continuous integration

Picture four people on one Go service. Each works alone on a branch for three weeks, then they all merge on Friday. Nothing compiles, two people renamed the same function, and tests that passed on every branch fail together.

*Continuous integration* (CI) is the cure: everyone merges small changes into main at least daily, and an automated system builds and tests every push and every pull request. A conflict found after one hour of work is a five-minute fix; after three weeks, it's a project.

```text
  big-bang merges                       continuous integration
  ───────────────                       ──────────────────────
  1 huge merge every 3 weeks            many small merges a day
  conflicts pile up, hidden             each conflict tiny, found at once
  "which of 400 changes broke it?"      "the last commit broke it"
```

### The red build rule

A *red build* is a failed pipeline run on the main branch. The rule: **when main is red, fixing it is the team's top job**, either with a quick fix or by reverting the commit that broke it.

A red main blocks everyone, and the next failure hides behind the first. Teams that leave main red for days stop trusting CI, and then it catches nothing. Reverting first and investigating later is normal, and nobody's fault.

## A pipeline: stages, jobs, steps

A *pipeline* is the automated path from a commit to running software. It's usually drawn as stages, each one a gate the change must pass before the next:

```text
  push / pull request
        │
        ▼
  ┌────────────┐   ┌────────────┐   ┌────────────┐   ┌──────────────┐   ┌────────────┐
  │ lint, vet  │──►│    test    │──►│   build    │──►│ package image│──►│   deploy   │
  │ gofmt, vet │   │ go test    │   │ go build   │   │ docker build │   │ staging,   │
  │            │   │ -race      │   │ -ldflags   │   │ push, sign   │   │ production │
  └────────────┘   └────────────┘   └────────────┘   └──────────────┘   └────────────┘
     seconds         a minute         a minute          a few minutes      on main only
```

**Cheap checks first.** `gofmt` and `go vet` take seconds, so they run before the slow stages: the sooner a pipeline fails, the fresher the change is in the author's head.

Most CI systems share the same three levels:

| Level | What it is | Example |
|---|---|---|
| workflow / pipeline | the whole thing, started by an event | "ci", run on every push |
| job | a group of steps on one fresh machine | `test`, `build`, `deploy` |
| step | one command or one reusable action | `go test ./...` |

Jobs run in parallel unless one depends on another. Each starts on a clean machine, so nothing carries over unless you pass it on as an artifact.

## GitHub Actions, the running example

GitHub Actions is GitHub's CI system; GitLab CI and Jenkins use different syntax for the same ideas. A *workflow* is a YAML file in `.github/workflows/` that GitHub runs whenever one of its triggers fires.

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
      - uses: actions/checkout@v7
      - uses: actions/setup-go@v7
        with:
          go-version: ${{ matrix.go-version }}
          cache-dependency-path: go.sum
      - run: go vet ./...
      - run: go test -race ./...

  build:
    needs: test
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v7
      - uses: actions/setup-go@v7
        with:
          go-version: '1.27'
          cache-dependency-path: go.sum
      - run: go build -ldflags "-X main.version=${{ github.sha }}" -o bin/deployctl ./cmd/deployctl
      - uses: actions/upload-artifact@v7
        with:
          name: deployctl
          path: bin/deployctl

  deploy:
    needs: build
    if: github.ref == 'refs/heads/main'
    runs-on: ubuntu-latest
    environment: production
    steps:
      - uses: actions/download-artifact@v8
        with:
          name: deployctl
      - run: ls -l deployctl   # your real deploy command goes here
```

Trace a pull request through it:

1. `on:` lists the triggers: a push to `main`, or any pull request.
2. `permissions:` lets the workflow's automatic token read the code and nothing else.
3. `runs-on:` picks the *runner*, the machine for the job: a fresh virtual machine, thrown away afterwards.
4. `test` has a `strategy.matrix`, so GitHub runs **two** copies in parallel, with `${{ matrix.go-version }}` set to `1.26` in one and `1.27` in the other.
5. Each step either `uses:` a published *action* (reusable code; `actions/checkout` clones the repository) or `run:`s a shell command.
6. `build` has `needs: test`: it waits for both test copies and is skipped if either fails. It stamps the commit into the binary with `-ldflags -X`, as in *Packages and project layout*, and uploads it as an artifact.
7. `deploy` runs only `if:` the ref is `refs/heads/main`. On a pull request the ref is like `refs/pull/42/merge`, so it's skipped; after the merge, the push to `main` runs all three jobs.
8. `environment: production` can make the job wait for a person to approve it.

### YAML gotchas

YAML guesses types, and sometimes guesses wrong.

| You write | YAML reads | Fix |
|---|---|---|
| `go-version: 1.20` | the number `1.2`, so you get Go 1.2 or an error | `'1.20'` |
| `country: NO` | in older YAML, the boolean `false` | `'NO'` |
| a tab for indentation | a parse error | spaces only |
| a step indented one space too far | a different structure, or an error | let your editor show indentation |

**Quote every version number.** `1.26` survives as a number, but `1.30` won't. Quote them all and you never have to think about it.

## Caches and artifacts

Both save files from a job, but for different reasons.

A *cache* makes jobs faster. It stores the downloaded Go modules under a *cache key*, and the next run with the same key restores them instead of downloading. The key includes a hash of `go.sum`, so it changes exactly when your dependencies do:

```text
  key: Linux-go-<hash of go.sum>
  go.sum unchanged  ─► same key ─► cache hit,  modules restored in seconds
  go.sum changed    ─► new key  ─► cache miss, download, save under new key
```

`actions/setup-go` does this for you; `cache-dependency-path: go.sum` tells it which file to hash. With the general `actions/cache` action you write the key yourself: `${{ runner.os }}-go-${{ hashFiles('**/go.sum') }}`.

An *artifact* is an output you keep: a binary, a test report. Artifacts pass results to a later job (as `build` hands `deployctl` to `deploy` above), and stay downloadable from the run's page for days.

| | Cache | Artifact |
|---|---|---|
| purpose | speed | keep or pass on an output |
| if it's missing | the job is slower, still correct | the next job fails |

**A cache must never change the result.** A build that only passes with a warm cache is broken.

## Build once, promote everywhere

A common anti-pattern: build the image for staging, test it, then build again for production. The second build pulls a newer base image, and production runs code nobody tested.

Instead, build **one** image per commit and *promote* that exact image from one environment to the next. Only configuration differs, read from environment variables as in *Container config: settings from the environment*.

```text
  commit 3f9c2ab ─► build ─► registry: api@sha256:91e4...   (one image)
                                    │
                    ┌───────────────┼────────────────┐
                    ▼               ▼                ▼
                   dev  ───────►  staging ──────►  production
                 (auto)          (auto)         (approval needed)
               same digest     same digest       same digest
```

Refer to the image by its *digest* (`sha256:...`). A tag like `:latest` can be moved to another image; a digest can't.

A GitHub *environment* is a named target like `production` with its own secrets and rules. A *required reviewer* rule pauses the deploy job until a named person approves: your last chance to say "not at 5pm on a Friday".

### Delivery vs deployment

The "CD" in CI/CD means one of two things:

| Term | Every green commit on main... | A person... |
|---|---|---|
| *continuous delivery* | is built, tested and ready to release | decides when it goes to production (one click) |
| *continuous deployment* | goes to production automatically | is not in the loop for normal changes |

Continuous deployment needs excellent tests and monitoring, because nobody looks before production does. Many teams start with delivery.

## Deployment strategies

Version 1 is serving traffic and version 2 is in the registry. How do you swap them?

- **Rolling update:** replace instances a few at a time. Kubernetes Deployments do this by default.
- **Blue/green:** run a full copy of the new version (green) beside the old (blue), test it, then switch all traffic at once.
- **Canary:** send a small slice of traffic, say 5%, to the new version and watch error rates. Healthy? Widen the slice. Not? Send everything back.

| Strategy | Extra capacity | Rollback speed | Users hit by a bad release | Complexity |
|---|---|---|---|---|
| rolling | a little | minutes (roll back the same way) | growing share during the roll | low |
| blue/green | double, during the switch | seconds (switch back) | all, until you switch back | medium |
| canary | a little | seconds (send the slice back) | only the canary slice | high: needs metrics and traffic splitting |

During a rolling or canary deploy, **two versions run at once**, against the same database. Remember that whenever you change a table or a message format.

### Rollbacks

It's 5:10pm, the deploy finished at 5:00, and errors have tripled. **Roll back first, investigate after.**

A rollback is a deploy of the previous image digest, still in the registry because you built once and kept it. It works only if the old version still runs against today's database. So make *backward-compatible migrations*: never rename or drop, in one release, something the previous release still reads.

```text
  release 1: add column email_v2, write both columns   ◄─ previous release still works
  release 2: read only email_v2                         ◄─ release 1 still works
  release 3: drop the old column                        ◄─ nothing reads it any more
```

## GitOps

So far a CI job pushes changes into the cluster. *GitOps* flips this: the desired state of every environment (which image digest, how many replicas) lives as files in Git, and an agent inside the cluster, such as Argo CD or Flux, keeps comparing the cluster to Git and fixes any difference.

```text
  CI builds image ─► commit "api: digest sha256:91e4..." to the config repo
                                       │
                                       ▼
                      agent in the cluster: Git says X, cluster has Y
                                       │
                                       ▼
                              change the cluster to X, repeat forever
```

This is the control loop from *Declarative systems and control loops*: observe, compare, act. In return:

- **Git is the audit log.** Who deployed what, and when, is in `git log`.
- **Rollback is `git revert`.** The agent sees the old digest and puts it back.
- **Drift heals itself.** A hand edit to the cluster gets changed back.
- **CI needs no cluster credentials.** It only writes to Git.

## Secrets in CI

A CI runner holds the keys to production. Most leaks are accidents: a debug `echo`, a verbose flag, a token committed "just for a test".

| Practice | Why |
|---|---|
| store tokens as *repository secrets*, read as `${{ secrets.NAME }}` | never in the workflow file or the repo |
| put production secrets on the `production` environment | only approved deploy jobs can read them |
| never `echo` a secret | GitHub *masks* secret values as `***` in logs, but an encoded copy slips through |
| set `permissions:` to the minimum, per job | a compromised step can do only what the token allows |
| use OIDC instead of long-lived cloud keys | nothing durable to leak |

**OIDC, short-lived credentials.** The old way: paste an AWS access key into a secret and hope it never leaks; it stays valid for years. With *OIDC* (OpenID Connect), the job gets a signed identity token from GitHub ("I'm repo X, branch main") and the cloud swaps it for credentials that expire within the hour. The job needs `permissions: id-token: write`, and the cloud role trusts only that repository and branch.

**If a token leaks:** revoke and rotate it first, then delete the log, then find out how. Assume it has already been copied.

## Supply chain security

Your pipeline runs other people's code: actions, base images, Go modules. Four habits keep that safe:

- **Pin actions to a full commit SHA.** The tag in `actions/checkout@v7` can be moved to new code by its owner, or by whoever steals their account. `actions/checkout@3d3c42e5aac5ba805825da76410c181273ba90b1  # v7.0.1` can't change. Dependabot keeps SHAs up to date.
- **Produce an SBOM.** A *software bill of materials* lists every package inside an image. When the next big vulnerability hits the news, you search SBOMs instead of guessing.
- **Sign images.** Tools like cosign, from the Sigstore project, attach a signature to the image digest in the registry. The cluster can refuse any image your pipeline didn't sign.
- **Record provenance.** *SLSA* ("salsa") is a framework for how trustworthy a build is. Its key idea is *provenance*: a signed record of which commit, workflow and builder produced an artifact.

These close the gap *Image security: non-root and no secrets* leaves open: a hardened image is still dangerous if someone else built it.

## Common mistakes

- **Leaving main red.** Every later commit lands on a broken base. Fix it or revert it now.
- **Long-lived branches.** Weeks of drift mean a painful merge. Merge small changes daily.
- **Unquoted versions in YAML.** `1.20` becomes `1.2`. Quote every version.
- **Rebuilding per environment, or deploying `:latest`.** Production runs something untested, and you can't name what to roll back to. Build once, deploy by digest.
- **A migration that breaks the previous release.** The rollback fails too. Add first, remove in a later release.
- **Echoing secrets while debugging.** Masking is a safety net, not a guarantee. Rotate anything that hits a log.
- **Broad permissions and long-lived cloud keys.** Start each job from `contents: read`, and use OIDC.
- **Actions pinned to a tag.** Their owner can change the code under you. Pin to a full commit SHA.

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

## Answers

1. Fixing a red main comes first: fix it quickly or revert the commit that broke it. While main is red nobody can ship, new failures hide behind the first, and people stop trusting CI.
2. It's skipped. On a pull request the ref is like `refs/pull/42/merge`, so the `if:` is false. Only a push to main, after `test` passes, runs the deploy.
3. YAML reads unquoted `1.20` as the number 1.2, so Go 1.2 is requested. Writing `'1.20'` keeps it a string.
4. A cache (like Go modules keyed by a hash of `go.sum`) only saves time: without it the job is slower but still correct. An artifact is an output passed on or kept, like a binary: without it, the job that needs it fails.
5. A second build can pick up a newer base image or different dependencies, so production would run something never tested. Promoting the same image digest means production runs exactly what passed staging.
6. With a canary, only the small slice of traffic, say 5%, until metrics show the problem and traffic goes back. With blue/green, the switch moves everyone at once, though switching back takes seconds.
7. A rollback runs the previous image against today's database. If the migration renamed or dropped something that release reads, it crashes too, and there's no way back.
8. CI commits the new image digest to a Git repository. An agent in the cluster, like Argo CD or Flux, sees Git and the cluster differ and changes the cluster to match. To roll back, `git revert` that commit.
9. A long-lived key stays valid for months or years, so a leaked copy works until someone notices. With OIDC the job swaps a signed identity token for credentials that expire within the hour and work only for that repository and branch. There's no stored key to steal.

## Key terms

| Term | Meaning |
|---|---|
| continuous integration (CI) | merging small changes into main often, each built and tested automatically |
| red build | a failed pipeline run on main; fixing or reverting it comes first |
| pipeline | the automated path from a commit to running software: lint, test, build, package, deploy |
| workflow | a GitHub Actions pipeline: a YAML file in `.github/workflows/` |
| job | steps that run together on one fresh machine; jobs run in parallel unless one `needs` another |
| runner | the machine that runs a CI job |
| matrix build | one job run in parallel with different settings, like several Go versions |
| cache (CI) | files kept between runs to save time, under a key like a hash of `go.sum` |
| artifact | a job's output, kept or passed to a later job, like a binary |
| image digest | an image's `sha256:` content hash; unlike a tag, it can't be moved |
| promotion | moving the same built image from dev to staging to production, never rebuilding |
| continuous delivery | every passing change is ready to release; a person decides when |
| continuous deployment | every passing change goes to production automatically |
| rolling update | replacing instances with the new version a few at a time |
| blue/green deployment | running the new version beside the old, then switching all traffic at once |
| canary deployment | sending a small slice of traffic to the new version, widening it if healthy |
| GitOps | desired state lives in Git; an agent in the cluster keeps the cluster matching it |
| OIDC (in CI) | a job proves who it is with a signed token and gets cloud credentials that expire soon |
| SBOM | software bill of materials: the list of every package inside an image |
| provenance | a signed record of which source, workflow and builder produced an artifact |

## Watch

- [Continuous Integration vs Feature Branch Workflow](https://www.youtube.com/watch?v=v4Ijkq6Myfc) · Modern Software Engineering · 18 min: Dave Farley, who co-wrote the book on continuous delivery, on why merging daily beats long-lived branches.
- [GitHub Actions Tutorial - Basic Concepts and CI/CD Pipeline with Docker](https://www.youtube.com/watch?v=R8_veQiYBjI) · TechWorld with Nana · 33 min: workflows, jobs, runners and actions, then a pipeline that builds and pushes a Docker image.
- [Top 5 Most-Used Deployment Strategies](https://www.youtube.com/watch?v=AWVTKBUnoIg) · ByteByteGo · 10 min: animated rolling, blue/green and canary deploys with their trade-offs.
- [What is GitOps, How GitOps works and Why it's so useful](https://www.youtube.com/watch?v=f5EpcWp0THw) · TechWorld with Nana · 12 min: push vs pull deployments, and why desired state in Git helps.

## Go deeper

- [Continuous Integration](https://martinfowler.com/articles/continuousIntegration.html): Martin Fowler's classic on the practices behind CI and why they work.
- [Workflow syntax for GitHub Actions](https://docs.github.com/en/actions/reference/workflows-and-actions/workflow-syntax): the reference for every key used here.
- [Secure use reference](https://docs.github.com/en/actions/reference/security/secure-use): GitHub's checklist for secrets, token permissions, OIDC and pinning actions.
- [Blue Green Deployment](https://martinfowler.com/bliki/BlueGreenDeployment.html): the short original description of blue/green, database included.
- [SLSA](https://slsa.dev/): the framework's site, explaining build levels and provenance.
