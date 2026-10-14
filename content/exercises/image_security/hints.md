## Task 1: Run as a non-root user

One instruction sets who the container runs as. It belongs in the final
stage, the one that starts at the second `FROM`, above `ENTRYPOINT`.

### Closer

`USER` takes `UID:GID` as numbers. Distroless's unprivileged user and group
are both 65532: `USER <uid>:<gid>`.

### Answer

```dockerfile
FROM gcr.io/distroless/static-debian13
COPY --from=build /out/api /api
EXPOSE 8080
USER 65532:65532
ENTRYPOINT ["/api"]
```

## Task 2: Keep the password out of the image

Find every line that mentions the password: the `ARG`, the `ENV` that copies
it, and the comment above them. The program reads `DATABASE_PASSWORD` from
its environment, so whoever runs the container supplies it.

### Closer

Delete the `ARG DATABASE_PASSWORD=...` line, the `ENV DATABASE_PASSWORD=...`
line and the comment that names the staging password. Leave the `ENV` with
`DATABASE_HOST` and `DATABASE_USER` alone.

### Answer

```dockerfile
FROM gcr.io/distroless/static-debian13:nonroot
COPY --from=build /out/migrate /migrate
ENV DATABASE_HOST=db.internal DATABASE_USER=migrate

# DATABASE_PASSWORD comes from the environment at run time: docker run -e,
# or a Kubernetes Secret.
ENTRYPOINT ["/migrate"]
```

## Task 3: Use a build secret for private modules

The token's only job was to log git in to GitHub. A netrc file mounted as a
secret does that job for one `RUN` step, so the `ARG` and the `git config`
part can go.

### Closer

Delete `ARG GITHUB_TOKEN` and the `git config ... &&` part. Put the mount
flag between `RUN` and the command:
`RUN --mount=type=secret,id=...,target=... go mod download`.

### Answer

```dockerfile
FROM golang:1.26 AS build
ENV GOPRIVATE=github.com/acme/*
WORKDIR /src
COPY go.mod go.sum ./

RUN --mount=type=secret,id=netrc,target=/root/.netrc go mod download
```

The secret exists only while that step runs, so no layer, cache entry or
history row holds it.
