## Task 1: Download modules before copying the code

`go mod download` only needs two files. Copy those first, download, and only
then copy everything else.

### Closer

Four lines after `WORKDIR`, in this order: a `COPY` of `go.mod` and
`go.sum`, the `RUN` that downloads, `COPY . .`, and the unchanged build.

### Answer

```dockerfile
FROM golang:1.26
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o /usr/local/bin/api ./cmd/api
CMD ["api"]
```

## Task 2: Install packages in one step

Join the two `RUN` lines into one, and clean up the package lists at the
end of that same `RUN`.

### Closer

`RUN apt-get update && apt-get install -y --no-install-recommends ... && rm -rf ...`.
Break it over lines with a `\` at the end of each line but the last.

### Answer

```dockerfile
RUN apt-get update \
    && apt-get install -y --no-install-recommends ca-certificates curl \
    && rm -rf /var/lib/apt/lists/*
```

## Task 3: Install tools before copying the code

The install doesn't read your code, so it can move up, above the `COPY` of
the whole context.

### Closer

Cut the `go install` line and paste it right after `FROM`. Leave
`go generate` and `go build` where they are.

### Answer

```dockerfile
FROM golang:1.26
RUN go install golang.org/x/tools/cmd/stringer@v0.52.0
WORKDIR /app
COPY go.mod go.sum /app/
RUN go mod download
COPY . /app/
RUN go generate ./...
RUN go build -o /usr/local/bin/reportgen ./cmd/reportgen
CMD ["reportgen"]
```
