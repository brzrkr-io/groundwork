# Practice: containerize and ship a service

`inventoryd` is the service from *Project: inventory service*, cut down to
its two health endpoints. It has no image and no pipeline yet. You'll make
it stop cleanly when its container is stopped, then give it everything it
needs before its first deploy: a Dockerfile, an ignore file, a CI workflow
and a release workflow.

**New here:** nothing; practice.

## Learn

Each recap is a reminder, not the full lesson. If a piece feels new, go back
to the exercise it names.

### Recap: draining on SIGTERM

From *Graceful shutdown: drain in-flight requests* and
*Health checks: readiness gates*. `docker stop` and Kubernetes stop a
container the same way: SIGTERM to its first process, a wait (10 seconds
for `docker stop`, 30 in Kubernetes), then SIGKILL, which shows up as exit
code 137. A service that drains in that window loses no requests:

| Step | Why |
|---|---|
| `signal.NotifyContext(ctx, syscall.SIGTERM, os.Interrupt)` | turns SIGTERM into a cancelled `ctx`; `main.go` already does this |
| `Serve` in a goroutine, its error on a buffered channel | so the code can wait for `ctx` at the same time |
| `select` on that channel and `ctx.Done()` | whichever happens first wins |
| store `false` in the readiness gate | `/readyz` answers 503, so no new traffic is sent here |
| `Shutdown` with `context.WithTimeout(context.Background(), grace)` | stops accepting, waits for requests in flight, but not forever |
| `Close` if `Shutdown` fails | cuts off whatever is still stuck after `grace` |

The signal only reaches your code if the binary *is* the first process: an
exec-form `ENTRYPOINT ["/prog"]`, as in the *Containers and images* guide.

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

From *Image security: non-root and no secrets*:

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

- Pin every action to a full commit SHA, with its release in a comment
  (`owner/repo@<40 hex characters> # vX.Y.Z`). The tests check every `uses`
  line. To find a SHA yourself:
  `git ls-remote --tags https://github.com/OWNER/REPO 'vX.Y.Z*'`.
- `actions/setup-go` with `go-version-file: go.mod` uses the same Go
  version as your `go.mod`.
- `go test -race` runs the tests with the race detector from
  *The race detector*. CI is where it's cheap to run every time.

The pins for this exercise:

| Action | Release | Commit |
|---|---|---|
| `actions/checkout` | v7.0.1 | `3d3c42e5aac5ba805825da76410c181273ba90b1` |
| `actions/setup-go` | v7.0.0 | `b7ad1dad31e06c5925ef5d2fc7ad053ef454303e` |
| `docker/login-action` | v4.6.0 | `dbcb813823bdd20940b903addbd779551569679f` |
| `docker/build-push-action` | v7.4.0 | `c3c9e263c25d99ce0380d002d59b67737d91b0dc` |
| `docker/setup-buildx-action` (optional) | v4.4.1 | `f87e5991a6d7451dcb8d9637bfbc97413f497069` |

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
├── main.go                    inventoryd itself: task 1 is its run method
├── go.mod, go.sum             the module (go.sum is there because the tests use sigs.k8s.io/yaml)
├── Dockerfile                 task 2: the image
├── Dockerfile.dockerignore    task 3: what the build leaves out
├── ci.yml                     task 4: in a real repository, .github/workflows/ci.yml
├── release.yml                task 5: in a real repository, .github/workflows/release.yml
├── exercise_test.go           one test per task
├── dockerfile_test.go         helpers that read Dockerfiles
└── workflow_test.go           helpers that read workflows
```

`main.go` listens on the port in `$PORT` (default `8080`). `/healthz`
answers `ok <version>`; `/readyz` answers `ready` (200) while the service
is ready, and `not ready` (503) otherwise. `inventoryd -version` prints the
version and exits. `version` is `"dev"` unless the build sets it with
`-ldflags`.

You don't need Docker: the tests run `run` and read your other files. If
you have Docker, the last section, "Run it for real", builds the image and
stops the real container.

### 1. Drain on SIGTERM

**Goal:** in `main.go`, rewrite the `run` method so `inventoryd` reports
ready while it serves and, once `ctx` is cancelled, drains instead of
cutting requests off.

`run` belongs to `service`, which holds `s.srv` (the `*http.Server`),
`s.ready` (the `atomic.Bool` that `/readyz` reads) and `s.grace` (how long
to drain). Today's draft never marks the service ready, and on SIGTERM it
calls `Close`, which drops every request in flight. Make it:

- start `s.srv.Serve(ln)` in a goroutine that sends its error on a
  channel with room for one value, then store `true` in `s.ready`;
- wait for whichever comes first: Serve's error (return it as it is), or
  `ctx.Done()`;
- when `ctx` is done: store `false` in `s.ready`, log a line containing
  `draining` (`log.Print("draining")`), then call `s.srv.Shutdown` with a
  context that times out after `s.grace`;
- return `nil` if `Shutdown` succeeds. If it fails, call `s.srv.Close()`
  and return its error wrapped as `forced shutdown: <error>` with `%w`.

What the tests check:

```text
GET /readyz while run serves         => 200
ctx cancelled, nothing in flight     => run returns nil, s.ready is false, "draining" logged
ctx cancelled, one request in flight => s.ready turns false at once; the request
                                        still gets 200; then run returns nil
a request stuck longer than s.grace  => run returns "forced shutdown: context deadline exceeded"
Serve fails (closed listener)        => run returns Serve's error
```

**Think it through:**
- Why turn readiness off *before* calling `Shutdown`, not after?
- `Shutdown` needs a context. Why not `ctx` itself? What state is it in?

### 2. Write the Dockerfile

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
    copied it to, so SIGTERM reaches the `run` you wrote in task 1.

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

### 3. Keep the build context small

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

### 4. Test every change

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

### 5. Release on tags

**Goal:** in `release.yml`, write the workflow `release` that, when a
version tag is pushed, tests the code and then pushes the image
`ghcr.io/acme/inventoryd:<tag>`.

- `name: release`.
- Triggers: only `push`, of tags matching `'v*'`, with no `branches`.
- A job `test`, exactly like the one in task 4.
- A job `image` on `ubuntu-latest` that `needs` the `test` job, with
  exactly these `permissions`: `contents: read` and `packages: write`.
  Its steps, in this order (all pinned, from the table in Learn):
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

## Run it for real (optional)

The tests read your files; this runs them. You need Docker. From this
folder:

```bash
docker build -t inventoryd:dev --build-arg VERSION=v1.4.0 .
docker run -d --name inventoryd -p 8080:8080 inventoryd:dev
curl localhost:8080/healthz      # ok v1.4.0
curl localhost:8080/readyz       # ready
docker stop inventoryd           # SIGTERM now, SIGKILL after 10 seconds
docker logs inventoryd
docker inspect --format '{{.State.ExitCode}}' inventoryd
docker rm inventoryd
```

`docker logs` shows the container's whole life, and the exit code is `0`:

```text
2026/10/13 09:12:01 inventoryd v1.4.0 listening on [::]:8080
2026/10/13 09:12:09 draining
2026/10/13 09:12:09 inventoryd stopped cleanly
```

`docker stop` returned at once, because `run` drained and returned. If it
takes 10 seconds instead, with no `draining` line and exit code 137, SIGTERM
never reached your code: check that `ENTRYPOINT` is in exec form, with
square brackets.

On NixOS, Docker sends container logs to journald by default. If
`docker logs inventoryd` prints nothing, the same lines are in
`journalctl CONTAINER_NAME=inventoryd`. To get `docker logs` back, set
`virtualisation.docker.logDriver = "json-file";` in your NixOS
configuration. *kind* (Kubernetes in Docker) needs that setting too; for
plain Docker it's optional.

## Learn more

- [Build context and .dockerignore (Docker docs)](https://docs.docker.com/build/concepts/context/)
- [Publishing Docker images (GitHub docs)](https://docs.github.com/en/actions/tutorials/publish-packages/publish-docker-images)
- [docker/build-push-action](https://github.com/docker/build-push-action)
