## Task 1: Pin the base images

Two `FROM` lines name an image. Each needs a fixed tag and its digest from
the README's table; the stage name `AS build` stays where it is.

### Closer

The pieces go in this order: image, `:`, tag, `@`, the whole `sha256:...`
digest. The final stage's tag changes from `latest` to `nonroot`:
`FROM gcr.io/distroless/static-debian13:nonroot@sha256:...`.

### Answer

```dockerfile
FROM golang:1.26@sha256:727841129a8d3bbd6b99dc9f63427b62bb5a98f6e6e856a5b1399bf19152c03e AS build

FROM gcr.io/distroless/static-debian13:nonroot@sha256:f0b02dd57038ad9b032fcb6ad271ebeaf4cc1804bd067250270908b22f4b0114
```

## Task 2: Label the image

The labels go in the final stage. Their values come from build arguments,
and the final stage can't see the build stage's `ARG VERSION`, so it needs
its own `ARG` lines first.

### Closer

Declare both arguments, then one `LABEL` with three `key=value` pairs,
continued with `\`:

```dockerfile
ARG VERSION
ARG REVISION
LABEL org.opencontainers.image.source="..." \
      org.opencontainers.image.revision="${...}" \
      ...
```

### Answer

```dockerfile
FROM gcr.io/distroless/static-debian13:nonroot
COPY --from=build /out/worker /worker
ARG VERSION=dev
ARG REVISION
LABEL org.opencontainers.image.source="https://github.com/acme/worker" \
      org.opencontainers.image.revision="${REVISION}" \
      org.opencontainers.image.version="${VERSION}"
ENTRYPOINT ["/worker"]
```

## Task 3: Keep secrets out of the build context

Go through `testdata/context.txt` and group the files to leave out by
folder. A pattern naming a folder covers everything inside it, and
patterns start at the root of the context.

### Closer

A handful of lines cover it: one for `.git`, one or two for the `.env` files,
and one each for the folders `deploy/tls`, `secrets` and `bin`. Avoid
`*.key`: it only matches at the root.

### Answer

```text
.git
.env
.env.*
deploy/tls/
secrets/
bin/
```

`.env.*` also leaves out `.env.example`; add `!.env.example` after it if
you want to keep that one.
