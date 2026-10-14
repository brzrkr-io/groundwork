# Practice: containerize and ship a service

`inventoryd` is a small Go service with no image and no pipeline. You'll
give it everything it needs before its first deploy: a Dockerfile, an
ignore file, a CI workflow and a release workflow.

**New here:** nothing; practice.

## Learn

Each recap is a reminder, not the full lesson. If a piece feels new, go back
to the exercise it names.

### Recap: a multi-stage Dockerfile

From *Dockerfiles* and *Multi-stage builds*:

| Piece | What it does |
|---|---|
| `FROM image:tag AS name` | starts a stage and gives it a name |
| `WORKDIR /path` | the folder every later instruction works in |
| copy the module files, `RUN go mod download`, then `COPY . .` | the download gets its own cached layer, rebuilt only when `go.mod` or `go.sum` change |
| `ARG NAME=default` | a build setting; `docker build --build-arg NAME=value` overrides it |
| `CGO_ENABLED=0` | Go builds a static binary that needs no C libraries from the image |
| `-ldflags "-X main.version=..."` | sets the `version` variable in package `main` at build time |
| `COPY --from=name src dest` | takes one file from an earlier stage |
| `ENTRYPOINT ["/prog"]` | the program the container runs; words after the image name in `docker run` become its arguments |

`${NAME}` inside a `RUN` line is replaced by the `ARG`'s value.

### Recap: image security

From *Image security*:

- Run from a *distroless* final stage: no shell and no package manager, so
  there is less to attack. It also means `RUN` and shell-form commands
  can't work there.
- Switch to a non-root user with numbers, `USER 65532:65532`. Kubernetes'
  `runAsNonRoot` check can only verify a numeric user.
- Secrets never go into an image. `COPY . .` copies whatever is in the
  build context, so keep `.env` files out of the context itself.

### Recap: the build context and ignore files

`docker build .` sends the folder `.` (the *build context*) to the builder.
An ignore file lists what to leave out, one pattern per line:

| Pattern | Leaves out |
|---|---|
| `tmp` | `tmp` at the top of the context; if it's a folder, everything in it too |
| `/tmp/` | the same: leading and trailing slashes don't matter |
| `*.log` | `.log` files at the top only: `*` never crosses a `/` |
| `**/*.log` | `.log` files in any folder, the top included |
| `tmp*` | `tmp`, `tmp.db`, `tmp-old/`, ... at the top |
| `!keep.log` | an exception: `keep.log` stays, though an earlier line left it out |
| `# text` | a comment, when `#` is the first character of the line |

When several lines match a path, the **last** one decides.

The usual name is `.dockerignore`, at the top of the context. BuildKit also
reads `<Dockerfile name>.dockerignore` next to the Dockerfile, and that one
wins. Files starting with a dot can't be part of a Groundwork exercise, so
here you write `Dockerfile.dockerignore`, the ignore file for `Dockerfile`.

### Recap: CI workflows

From *CI workflows*: a workflow says when it runs (`on`) and what runs
(`jobs`). Every job starts on an empty machine, so it checks out the code
and installs Go first.

```yaml
on:
  push:
    branches: [release]    # pushes to the release branch only
  pull_request:            # and every pull request
```

- Pin every action: `@v7`, `@v7.0.1` or a full commit hash, never `@main`.
- `actions/setup-go` with `go-version-file: go.mod` uses the same Go
  version as your `go.mod`.
- `go test -race` also looks for *data races*, a bug in concurrent code
  you'll meet later in the course. CI is where it's cheap to run.

### Recap: release pipelines

From *Release pipelines*:

| Piece | Meaning |
|---|---|
| `on: push: tags: ['deploy-*']` | runs when a matching Git tag is pushed |
| `${{ github.ref_name }}` | the tag (or branch) that started the run, such as `v1.4.0` |
| `permissions:` on a job | what that job's `GITHUB_TOKEN` may do; anything not listed is denied |
| `${{ secrets.GITHUB_TOKEN }}` | a token GitHub creates for every run; nothing to set up |
| `needs: other-job` | wait for that job, and skip this one if it fails |

The two Docker actions and the inputs you need:

| Action | Input | Value |
|---|---|---|
| `docker/login-action` | `registry` | the registry host, such as `ghcr.io` |
| | `username`, `password` | who logs in, and their token |
| `docker/build-push-action` | `context` | the build context: `.` is the checked-out folder |
| | `push` | `true` to push the image after building it |
| | `tags` | the image references, one per line |
| | `build-args` | `NAME=value` settings for `ARG`s, one per line |

`docker/setup-buildx-action`, before the build, is optional but recommended:
it gives the build its own BuildKit builder with more features.

> **Why is it written like this?**
> - **Docker requires:** an `ARG` declared inside the stage that uses it.
>   An `ARG` above the first `FROM` is visible only to `FROM` lines.
> - **GitHub requires:** `packages: write` for `GITHUB_TOKEN` to push to
>   ghcr.io. Once a job lists any permission, every one it doesn't list is
>   `none`, so it also needs `contents: read` to check out the code.
> - **Convention:** the image tag is the Git tag. The commit tagged
>   `v1.4.0` becomes `ghcr.io/acme/inventoryd:v1.4.0`, and its `-version`
>   prints `v1.4.0`.
> - **Convention:** patterns in YAML are quoted (`['v*']`). A value that
>   *starts* with `*` means something else in YAML, so the habit is to
>   quote them all.
> - **Author's choice:** the workflow and job names, and the paths
>   `/out/inventoryd` and `/inventoryd`.

## Your task

The warehouse team wants `inventoryd` in the cluster by Friday. Before that
it needs an image, a pipeline that tests every change, and a release that
publishes a versioned image whenever someone tags a commit.

### File map

```text
.
├── main.go                    inventoryd itself (given; don't change it)
├── go.mod, go.sum             the module (go.sum is there because the tests use sigs.k8s.io/yaml)
├── Dockerfile                 task 1: the image
├── Dockerfile.dockerignore    task 2: what the build leaves out
├── ci.yml                     task 3: in a real repository, .github/workflows/ci.yml
├── release.yml                task 4: in a real repository, .github/workflows/release.yml
├── exercise_test.go           one test per task
├── dockerfile_test.go         helpers that read Dockerfiles
└── workflow_test.go           helpers that read workflows
```

`main.go` listens on the port in `$PORT` (default `8080`) and answers
`/healthz` with `ok <version>`. `inventoryd -version` prints the version and
exits. `version` is `"dev"` unless the build sets it with `-ldflags`.

You don't need Docker: the tests read your files. If you have Docker, the
commands in each task work for real.

### 1. Write the Dockerfile

**Goal:** in `Dockerfile`, build `inventoryd` in a `build` stage and run it
from a distroless final stage as a non-root user.

The build stage, in this order:

1. Start from `golang:1.26`, and name the stage `build`.
2. Copy `go.mod` and `go.sum` (both: this folder has a `go.sum`), then
   `RUN go mod download`. A `WORKDIR` such as `/src` first is a good idea.
3. Copy the whole build context: `COPY . .`, or `COPY . /src/` with the
   folder spelled out.
4. Declare `ARG VERSION=dev`, anywhere in the stage before the build.
5. Build with exactly this line:
   `RUN CGO_ENABLED=0 go build -ldflags "-X main.version=${VERSION}" -o /out/inventoryd .`
   (`ENV CGO_ENABLED=0` in the stage, then the `RUN` without the
   `CGO_ENABLED=0 ` part, works too.)

The final stage:

6. Start from `gcr.io/distroless/static-debian13` (its `:nonroot` tag is
   fine too).
7. Copy `/out/inventoryd` from the `build` stage with `COPY --from=build`,
   to any path you like, such as `/inventoryd`.
8. `USER 65532:65532`.
9. `EXPOSE 8080`.
10. An `ENTRYPOINT` in exec form that runs the binary at the path you
    copied it to.

Exactly two `FROM` lines: one per stage.

```bash
docker build -t inventoryd:dev --build-arg VERSION=v1.4.0 .
docker run --rm inventoryd:dev -version
# => v1.4.0
```

**Think it through:**
- Which lines change on almost every commit, and which almost never? That
  decides the order of the copies.
- The final stage has no shell. Which forms of `ENTRYPOINT` can still
  work?

### 2. Keep the build context small

**Goal:** in `Dockerfile.dockerignore`, leave out everything the image must
not contain or doesn't need, and keep the source.

The test checks these paths from the real repository (most aren't in this
folder):

| Path | Must be | Why |
|---|---|---|
| `.git`, `.git/config` | left out | the Git history: large, and the build doesn't need it |
| `.env`, `.env.local` | left out | local settings and secrets |
| `bin/inventoryd` | left out | binaries from local builds (all of `bin/`) |
| `README.md`, `docs/runbook.md` | left out | Markdown docs, in any folder |
| `main.go`, `go.mod`, `go.sum` | kept | the build needs them |
| `internal/stock/stock.go`, `internal/stock/testdata/stock.json` | kept | the build and the tests need them |

Comments are welcome; they start with `#` in the first column.

**Think it through:**
- Which pattern covers both `.env` and `.env.local`, but not `main.go`?
- `*.md` only looks at the top of the context. What reaches into `docs/`?

### 3. Test every change

**Goal:** in `ci.yml`, write the workflow `ci` that vets and tests
`inventoryd` on pushes to `main` and on every pull request.

- `name: ci`.
- Triggers: `push` limited to the branch `main`, and `pull_request`.
- One job with the id `test`, on `ubuntu-latest`, with these steps in
  order:
  1. `actions/checkout`, pinned,
  2. `actions/setup-go`, pinned, with `go-version-file: go.mod`,
  3. a step that runs `go vet ./...`,
  4. a step that runs `go test -race ./...`.

**Think it through:**
- What's on the runner before checkout?
- `branches` takes a list, even for one branch.

### 4. Release on tags

**Goal:** in `release.yml`, write the workflow `release` that, when a
version tag is pushed, tests the code and then pushes the image
`ghcr.io/acme/inventoryd:<tag>`.

- `name: release`.
- Triggers: only `push`, of tags matching `'v*'`, with no `branches`.
- A job `test`, exactly like the one in task 3.
- A job `image` on `ubuntu-latest` that `needs` the `test` job, with
  exactly these `permissions`: `contents: read` and `packages: write`.
  Its steps, in this order (all pinned):
  1. `actions/checkout`,
  2. `docker/login-action` with `registry: ghcr.io`,
     `username: ${{ github.actor }}` and
     `password: ${{ secrets.GITHUB_TOKEN }}`,
  3. `docker/build-push-action` with `context: .`, `push: true`, the tag
     `ghcr.io/acme/inventoryd:${{ github.ref_name }}` in `tags`, and
     `VERSION=${{ github.ref_name }}` in `build-args`.

  A `docker/setup-buildx-action` step before the build is optional.

```bash
git tag v1.4.0 && git push origin v1.4.0
# release runs: test, then image pushes ghcr.io/acme/inventoryd:v1.4.0
```

**Think it through:**
- `image` runs on its own empty machine. What does it need before it can
  build?
- Which `ARG` in your Dockerfile does `build-args` fill in, and what does
  that change in the binary?

## Learn more

- [Build context and .dockerignore (Docker docs)](https://docs.docker.com/build/concepts/context/)
- [Publishing Docker images (GitHub docs)](https://docs.github.com/en/actions/tutorials/publish-packages/publish-docker-images)
- [docker/build-push-action](https://github.com/docker/build-push-action)
