# Multi-stage builds: fix the broken images

Four services were moved to multi-stage builds in a hurry. One image is
still huge, and three don't build or don't start. Each error message tells
you what's wrong, once you know how to read it.

**New here:** reading build and start-up errors from multi-stage images.

## Learn

### Recap

A multi-stage Dockerfile builds in one stage and ships another, as in
*Multi-stage builds*:

```dockerfile
FROM golang:1.26 AS compile
WORKDIR /src
COPY . .
RUN CGO_ENABLED=0 go build -o /bin/inventory ./cmd/inventory

FROM gcr.io/distroless/static-debian13
COPY --from=compile /bin/inventory /inventory
ENTRYPOINT ["/inventory"]
```

Four things must line up: the final base is small, the binary is static,
`--from=` names a real stage, and `ENTRYPOINT` runs the path the binary was
copied to.

### Reading the errors

| You see | When | It means |
|---|---|---|
| `exec /inventory: no such file or directory` | the container starts | the file *is* there, but it was built with cgo on: the kernel looks for the C *dynamic loader* (`/lib64/ld-linux-x86-64.so.2`) named inside the binary, and that's what's missing. Or the path in `ENTRYPOINT` really doesn't exist. |
| `exec: "/bin/sh": stat /bin/sh: no such file or directory` | the container starts | a shell-form `ENTRYPOINT` or `CMD` (no brackets) means `/bin/sh -c ...`, and distroless and scratch have no shell |
| `failed to compute cache key: ... "/bin/inventry": not found` | `docker build` | a `COPY` source doesn't exist: a typo in the path, or the build wrote the file somewhere else |
| `pull access denied, repository does not exist` for `docker.io/library/compiler` (older Docker: `invalid from flag value compiler`) | `docker build` | `--from=compiler` names no stage, so Docker tries to pull an *image* called `compiler` |

> **Why is it written like this?**
> - **Docker requires:** `--from=` takes a stage name (or number, counting
>   from 0), and if no stage has that name, an image name. That's why a
>   typo turns into a confusing "pull access denied".
> - **Docker requires:** the `exec` in "no such file or directory" is the
>   kernel starting your program. It reports the binary's path even when
>   the missing file is the loader the binary asks for.

### Why the big image matters

A final stage `FROM golang:1.26` works, but ships Debian, compilers, a shell
and Go: around 900 MB, and every package in it is one a security scanner
checks for known vulnerabilities (*CVEs*). The binary needs none of it.
Smaller image, fewer CVEs, faster pulls on every node.

## Your task

The scanner and the on-call log flagged four images. Each `TODO` quotes what
they saw. Fix each Dockerfile; keep its two stages and the build line. You
don't need Docker: the tests read your Dockerfiles. This folder holds only
the Dockerfiles, not the services' code, so the `docker build` lines in
their comments show how CI builds them and won't work here.

### 1. Shrink the 900 MB image

**Goal:** in `1_api.Dockerfile`, ship the api on a minimal runtime image.

- The final stage starts `FROM gcr.io/distroless/static-debian13`
  (`:nonroot` is fine too).
- It still copies `/out/api` from `build` and runs it with `ENTRYPOINT` in
  exec form.

**Think it through:**
- What does the api need at run time? What does `golang:1.26` add on top?

### 2. Fix: no such file or directory

**Goal:** in `2_worker.Dockerfile`, make `/worker` start on distroless
static.

- Build with cgo off: `CGO_ENABLED=0` on the build line, or
  `ENV CGO_ENABLED=0` in the build stage.
- Keep `-o /out/worker ./cmd/worker`.

**Think it through:**
- The file is there. What else does a binary built with cgo need?
- Why does the worker's use of package `net` matter here?

### 3. Fix the stage name and the entrypoint

**Goal:** in `3_reporter.Dockerfile`, make the image build and start.

- Every `COPY --from=` in the final stage names the build stage (rename
  either side, as long as they match).
- `ENTRYPOINT` is in exec form and runs the copied binary.

**Think it through:**
- What is the first stage called, and what does the `COPY` ask for?
- Which part of `ENTRYPOINT /reporter` needs a shell?

### 4. Run the file you copied

**Goal:** in `4_cron.Dockerfile`, make `ENTRYPOINT` run the binary the
`COPY` put in the image.

- The first item of the exec-form `ENTRYPOINT` is exactly the path the
  `COPY --from=build /out/cron ...` line copied to. Change either line.

**Think it through:**
- Where does the `COPY` put the file? Where does `ENTRYPOINT` look?

## Learn more

- [Multi-stage builds](https://docs.docker.com/build/building/multi-stage/)
- [ENTRYPOINT in the Dockerfile reference](https://docs.docker.com/reference/dockerfile/#entrypoint)
- [Distroless images](https://github.com/GoogleContainerTools/distroless)
