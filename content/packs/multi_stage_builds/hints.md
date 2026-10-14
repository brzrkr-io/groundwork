## Task 1: Add a runtime stage

A second `FROM` starts the new stage. It needs just two more lines: one to
copy the binary out of the stage called `build`, one to run it.

### Closer

`COPY --from=build <path in build> <path in new image>`, then
`ENTRYPOINT ["<path in new image>"]`.

### Answer

```dockerfile
FROM gcr.io/distroless/static-debian13
COPY --from=build /out/statusd /statusd
ENTRYPOINT ["/statusd"]
```

## Task 2: Build a static binary

The final stage copies `/out/backupctl` from a stage named `build`, so
write that stage above it, the cache-friendly way.

### Closer

`FROM golang:1.26 AS build`, `WORKDIR`, copy the module files, download,
copy the rest, then `RUN CGO_ENABLED=0 go build -o ... ./cmd/backupctl`.

### Answer

```dockerfile
FROM golang:1.26 AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /out/backupctl ./cmd/backupctl
```

## Task 3: Ship on scratch with certificates

The build stage is like the one in task 2. The final stage starts from
nothing, so copy two files into it: the binary and the certificate bundle.

### Closer

`FROM scratch`, then two `COPY --from=build` lines. The certificate file
keeps its path: `/etc/ssl/certs/ca-certificates.crt` on both sides.

### Answer

```dockerfile
FROM golang:1.26 AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /out/notifier ./cmd/notifier

FROM scratch
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=build /out/notifier /notifier
ENTRYPOINT ["/notifier"]
```
