# Multi-stage builds

The images from *Dockerfiles* ship the whole Go toolchain: about 900 MB,
with hundreds of packages a scanner can find holes in. A *multi-stage
build* compiles in one image and ships only the binary in another.

**New here:** `FROM ... AS name`, `COPY --from=`, `CGO_ENABLED=0`, static
binaries, distroless, `scratch`, `ENTRYPOINT` vs `CMD`.

## Learn

### Two stages in one Dockerfile

Every `FROM` starts a new *stage*: a fresh image with its own layers. Name a
stage with `AS`, and a later stage can copy files out of it with
`COPY --from=name`. Only the **last** stage becomes the image you tag and
push; earlier ones are thrown away.

```dockerfile
FROM golang:1.26 AS compile
WORKDIR /src
COPY . .
RUN CGO_ENABLED=0 go build -o /bin/inventory ./cmd/inventory

FROM gcr.io/distroless/static-debian13
COPY --from=compile /bin/inventory /inventory
ENTRYPOINT ["/inventory"]
```

```text
  stage "compile" (golang:1.26, ~900 MB)      final image (~10 MB)
  ┌──────────────────────────────┐            ┌──────────────────────┐
  │ Go toolchain, source code,   │  COPY      │ distroless static    │
  │ module cache                 │ ─────────▶ │ /inventory           │
  │ /bin/inventory               │  --from    │                      │
  └──────────────────────────────┘            └──────────────────────┘
```

The paths in `COPY --from=compile /bin/inventory /inventory` are both
absolute: the first is inside the `compile` stage, the second inside the new
one.

### Static binaries: `CGO_ENABLED=0`

By default, on Linux, Go can link a few packages (such as `net` and
`os/user`) against the system's C library. The binary then needs
`/lib/.../libc.so` and a *dynamic loader* at run time, which tiny images
don't have. `CGO_ENABLED=0` turns that off: the binary is *static*, one file
that needs nothing else.

```dockerfile
RUN CGO_ENABLED=0 go build -o /bin/inventory ./cmd/inventory
```

You can also set it for the whole stage with `ENV CGO_ENABLED=0` above the
build.

### Runtime images: distroless and scratch

| Base | What's inside | Size |
|---|---|---|
| `golang:1.26` | Debian, a shell, apt, compilers, Go | ~900 MB |
| `gcr.io/distroless/static-debian13` | CA certificates, time zones, `/etc/passwd`; **no shell, no package manager** | ~2 MB |
| `scratch` | nothing at all: an empty filesystem | 0 B |

With `scratch` you bring every file yourself. A Go program that calls an
HTTPS URL needs the *CA certificates* to check the server's certificate.
Without them every call fails with
`x509: certificate signed by unknown authority`. The `golang` image has
them at `/etc/ssl/certs/ca-certificates.crt`, and Go looks there, so copy
that file to the same path:

```dockerfile
COPY --from=compile /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
```

### `ENTRYPOINT` vs `CMD`

Both say what runs when the container starts. `ENTRYPOINT` is the program;
`CMD` is its default arguments, which `docker run image args...` replaces.

```dockerfile
ENTRYPOINT ["/inventory"]
CMD ["--port", "9090"]
```

`docker run inventory` runs `/inventory --port 9090`;
`docker run inventory --help` runs `/inventory --help`. For an image that
holds one program, `ENTRYPOINT` is the usual choice.

> **Why is it written like this?**
> - **Docker requires:** exec form (`["/inventory"]`) on distroless and
>   scratch. The shell form `ENTRYPOINT /inventory` means
>   `/bin/sh -c /inventory`, and there is no `/bin/sh`.
> - **Docker requires:** no `RUN` in a scratch or distroless stage. A
>   shell-form `RUN` needs `/bin/sh`, and there are no other programs to run.
> - **Convention:** most Dockerfiles call the build stage `build` (the
>   example says `compile` only to keep your task fresh), and put the binary
>   at a simple absolute path such as `/out/app`.
> - **Author's choice:** `static-debian13` without a tag here. The
>   `:nonroot` tag runs as an unprivileged user; you'll use it later.

## Your task

The team is moving every service to small images. Write the stages each
`TODO` asks for; the comments at the top say what each image is for.

### 1. Add a runtime stage

**Goal:** in `1_statusd.Dockerfile`, ship only the `statusd` binary from
the given build stage.

- Exactly two stages: keep the build stage, and add one after it.
- The new stage starts `FROM gcr.io/distroless/static-debian13`
  (`:nonroot` is fine too).
- It copies `/out/statusd` with `COPY --from=build`, to any absolute path.
- It runs that path with `ENTRYPOINT` in exec form.
- Nothing else: no `RUN`, no `COPY` from the build context.

**Think it through:**
- Which stage has the binary, and what is it called?
- Why can't the final stage run `go` or anything else?

### 2. Build a static binary

**Goal:** in `2_backupctl.Dockerfile`, write the build stage that the
given final stage copies from.

- `FROM golang:1.26 AS build`, above the given stage.
- `COPY go.mod go.sum ./`, `RUN go mod download`, then `COPY . .`, in that
  order, as in *Dockerfiles: fix the slow builds*.
- A `RUN` that builds `./cmd/backupctl` with `-o /out/backupctl`.
- cgo off: `CGO_ENABLED=0` on the build line, or `ENV CGO_ENABLED=0` in
  the stage.

**Think it through:**
- Where does the final stage expect to find the binary?
- What would the binary look for at run time if cgo were on?

### 3. Ship on scratch with certificates

**Goal:** in `3_notifier.Dockerfile`, write both stages, with a final image
that is empty except for `notifier` and the CA certificates.

- A build stage `FROM golang:1.26 AS build` that builds `./cmd/notifier`
  with cgo off and `-o /out/notifier`.
- A final stage `FROM scratch`, with nothing but `COPY --from=build`
  instructions and an `ENTRYPOINT`.
- It copies `/out/notifier` to any absolute path, and
  `/etc/ssl/certs/ca-certificates.crt` to the same path.
- `ENTRYPOINT` in exec form runs the copied binary.

**Think it through:**
- `notifier` posts over HTTPS. What does it need to trust the server?
- Is there a shell in `scratch` to run a shell-form `ENTRYPOINT`?

## Learn more

- [Multi-stage builds](https://docs.docker.com/build/building/multi-stage/)
- [Distroless images](https://github.com/GoogleContainerTools/distroless)
- [ENTRYPOINT in the Dockerfile reference](https://docs.docker.com/reference/dockerfile/#entrypoint)
- [cgo](https://pkg.go.dev/cmd/cgo)
