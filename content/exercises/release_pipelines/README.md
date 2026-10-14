# Release pipelines

CI tells you a commit works. A *release pipeline* turns a commit you chose
into a versioned binary and image, then walks that one image through
staging into production. It's the part of CI/CD you touch on every release
and every rollback.

**New here:** tag triggers, `${{ }}` expressions, `github.ref_name`,
`github.sha`, `if:`, `permissions:`, `secrets.GITHUB_TOKEN`,
`docker/login-action`, `docker/build-push-action`, `environment:`,
`concurrency:`.

## Learn

### A release starts with a tag

You mark a commit as a release with a Git tag, then push the tag:

```bash
git tag v1.4.2
git push origin v1.4.2
```

Pushing a tag is a `push` event, like pushing a branch. A `tags:` filter
picks which tags count:

```yaml
on:
  push:
    tags: ['nightly-*']     # tags such as nightly-2026-10-13
```

| Settings under `push:` | The workflow runs for |
|---|---|
| only `branches:` | pushes to matching branches; never for tags |
| only `tags:` | pushes of matching tags; never for branches |
| neither | every push, branch or tag |

In a pattern, `*` matches any characters except `/`.

### Expressions and the github context

`${{ ... }}` is an *expression*. GitHub replaces it with a value before the
step runs. The values come from *contexts*; the `github` context describes
the current run:

| Expression | After pushing tag `v1.4.2` | After pushing to `main` |
|---|---|---|
| `github.event_name` | `push` | `push` |
| `github.ref` | `refs/tags/v1.4.2` | `refs/heads/main` |
| `github.ref_name` | `v1.4.2` | `main` |
| `github.sha` | the 40-character commit hash | the 40-character commit hash |
| `github.actor` | the user who pushed | the user who pushed |

```yaml
- run: echo "built ${{ github.sha }} for ${{ github.actor }}"
```

### Stamping the version

In *Version info and ldflags* you set a variable at build time with
`-ldflags "-X main.release=2025.10"`. In a workflow, the fixed text becomes
an expression, so every tag stamps its own version into the binary. The
double quotes keep `-X main.version=...` together as one argument.

### Running a job only sometimes: if

```yaml
  announce:
    if: github.ref_name == 'main'
    runs-on: ubuntu-latest
```

`if:` holds an expression; there you may leave out the `${{ }}`. Text goes
in single quotes. `==`, `!=`, `&&` (and) and `||` (or) work as in Go. A job
whose `if` is false is *skipped*, not failed.

### The workflow's token and its permissions

Every run gets a token, `secrets.GITHUB_TOKEN`, that works only for this
repository and only until the run ends. You never create it. `permissions:`
decides what it may do:

| Permission | `read` lets the token... | `write` also lets it... |
|---|---|---|
| `contents` | clone the code | push commits and tags |
| `packages` | pull images from ghcr.io | push images to ghcr.io |
| `issues` | read issues | open and comment on issues |

```yaml
jobs:
  triage:
    runs-on: ubuntu-latest
    permissions:
      contents: read
      issues: write
```

> **Why is it written like this?**
> - **GitHub requires:** once you name one permission, every permission you
>   don't name becomes `none`. A job's own `permissions:` replaces the
>   workflow's, so a job that writes something still lists `contents: read`,
>   or checkout fails.
> - **Convention:** give each job the least it needs. A leaked token can do
>   only what its permissions allow.

### Logging in and pushing an image

GitHub's container registry is `ghcr.io`; images there are named
`ghcr.io/<owner>/<name>:<tag>`. Docker publishes two actions for the job:

| Action | Inputs | What it does |
|---|---|---|
| `docker/login-action` | `registry`, `username`, `password` | logs the runner in to a registry |
| `docker/build-push-action` | `context`, `push`, `tags`, `build-args` | builds the Dockerfile, then pushes the image |

Here is the pair for another registry:

```yaml
- uses: docker/login-action@v4
  with:
    registry: registry.example.com
    username: deploy-bot
    password: ${{ secrets.REGISTRY_PASSWORD }}
- uses: docker/build-push-action@v7
  with:
    context: .
    push: true
    tags: registry.example.com/billing:2.0.0
    build-args: |
      VERSION=2.0.0
      REGION=eu
```

- `context: .` builds from the checked-out folder, the files your other
  steps see. Without it the action fetches the commit itself.
- `push: true` pushes after building; without it the image stays on the
  runner and disappears with it.
- `tags` and `build-args` take one item per line (`|` starts a block of
  lines). A build arg fills the Dockerfile's `ARG` of the same name:
  `ARG VERSION` in the builder stage, then `-ldflags "-X main.version=$VERSION"`.

> **Why is it written like this?**
> - **GitHub requires:** a token as the password for ghcr.io, with
>   permission to push packages. The run's own `secrets.GITHUB_TOKEN`
>   works once the job has `packages: write`.
> - **Convention:** `github.actor` as the username, as the login-action
>   docs show.
> - **Docker recommends:** a `docker/setup-buildx-action` step before
>   building, for caching and multi-platform images. It's optional here.
> - **Author's choice:** tagging the image with the version, so the
>   image, the binary and the Git tag all say `v1.4.2`.

### Environments and approvals

`environment: staging` on a job says "this job deploys to staging". Each
environment is set up in the repository's settings (Settings, then
Environments), not in the YAML:

- **Required reviewers:** up to six people or teams. The job waits with
  "Waiting for review" until one of them approves.
- **Environment secrets:** only jobs that name the environment can read
  them, and only after approval.

The YAML only names the environment. That's on purpose: someone who can
edit the workflow still can't delete the approval. (On free plans,
required reviewers work in public repositories only.)

### Build once, promote the same image

```text
  tag v1.4.2 ─► image job ─► ghcr.io/acme/alertd:v1.4.2   (built once)
                                   │
                    deploy-staging ┘ ─► deploy-production
                    (same image)        (same image, after approval)
```

Production runs exactly the image that staging tested. Building again for
production would pull newer base images and dependencies, so production
would run something nobody tested. The guide explains it in
*Continuous integration and delivery*.

### One deploy at a time: concurrency

```yaml
  migrate-db:
    concurrency: database
```

Jobs with the same `concurrency` group never run at the same time in the
repository. If one is running, the next waits (*pending*); if another one
was already pending, the newer replaces it. The long form says the same:

```yaml
    concurrency:
      group: database
```

> **Why is it written like this?**
> - **GitHub requires:** `environment` and `concurrency` sit on the job,
>   beside `runs-on`.
> - **Convention:** no `cancel-in-progress: true` on deploys. It would stop
>   a running deploy halfway when a newer one starts.

## Your task

Releases of `statusd` and `alertd` are built on a laptop and copied to
servers by hand. You'll automate them: a version tag builds a stamped
binary, pushes an image, and promotes it from staging to production.

### 1. Release on version tags

**Goal:** in `1_release.yml`, run the workflow only for version tags, and
add a `release` job that builds `bin/statusd` stamped with the tag.

- `on:` has only `push`, with `tags: ['v*']` and no `branches`.
- Keep the `test` job.
- Add a job with the id `release`:
  - it `needs` the `test` job and runs on `ubuntu-latest`,
  - its steps: `actions/checkout` and `actions/setup-go` (pinned, with
    `go-version-file: go.mod`, as in *CI workflows*), then a step that runs
    exactly:

    ```bash
    go build -ldflags "-X main.version=${{ github.ref_name }}" -o bin/statusd .
    ```

Pushing tag `v1.4.2` runs `test`, then `release`, and the binary's
`main.version` is `v1.4.2`. Pushing a commit to `main` runs nothing.

**Think it through:**
- Which setting under `push:` limits it to tags?
- Which row of the github context table holds `v1.4.2`?

### 2. Build and push the image

**Goal:** in `2_image.yml`, give the `image` job the token permissions it
needs, and steps that push `ghcr.io/acme/statusd` tagged with the version.

- The job's own `permissions:` are exactly `contents: read` and
  `packages: write`.
- Its steps, in this order, each pinned:
  1. `actions/checkout`,
  2. `docker/login-action` with `registry: ghcr.io`,
     `username: ${{ github.actor }}` and
     `password: ${{ secrets.GITHUB_TOKEN }}`,
  3. `docker/build-push-action` with `context: .`, `push: true`,
     `tags` including `ghcr.io/acme/statusd:${{ github.ref_name }}` and
     `build-args` including `VERSION=${{ github.ref_name }}`.

You may add a `docker/setup-buildx-action` step before the build; the test
doesn't look for it. Pushing tag `v1.4.2` pushes the image
`ghcr.io/acme/statusd:v1.4.2`, with `v1.4.2` stamped into the binary
inside.

**Think it through:**
- What does the job need to read, and what does it need to write?
- The login and the push both need the token: which one gets it through
  an input?

### 3. Promote from staging to production

**Goal:** in `3_deploy.yml`, add two jobs that deploy the image job's
image to `staging`, then the same image to `production`.

The repository has a `deploy.sh`: `./deploy.sh staging v1.4.2` rolls out
`ghcr.io/acme/alertd:v1.4.2` to staging. Add:

- a job `deploy-staging`:
  - it needs `image`, runs on `ubuntu-latest`, and has
    `environment: staging`,
  - its steps: a pinned `actions/checkout`, then a step that runs
    `./deploy.sh staging ${{ github.ref_name }}`;
- a job `deploy-production`:
  - it needs `deploy-staging`, runs on `ubuntu-latest`, and has
    `environment: production`,
  - `concurrency: production` (or the long form with
    `group: production`), without `cancel-in-progress: true`,
  - its steps: a pinned `actions/checkout`, then a step that runs
    `./deploy.sh production ${{ github.ref_name }}`.
- Neither deploy job builds an image (no `docker/build-push-action`, no
  `docker build`).

With required reviewers set on `production` in the repository settings,
tag `v1.4.2` goes to staging on its own, then waits for an approval before
the same image reaches production.

**Think it through:**
- Which job does each deploy wait for, and why does that order matter?
- Where do the approval rules live, if not in this file?

## Learn more

- [Deployments and environments](https://docs.github.com/en/actions/reference/workflows-and-actions/deployments-and-environments)
- [Use GITHUB_TOKEN for authentication in workflows](https://docs.github.com/en/actions/tutorials/authenticate-with-github_token)
- [Publishing Docker images](https://docs.github.com/en/actions/tutorials/publish-packages/publish-docker-images)
- [docker/build-push-action](https://github.com/docker/build-push-action)
