## Task 1: Drain on SIGTERM

This is `Run` and `Stop` from *Graceful shutdown: drain in-flight requests*
in one function, plus the draining gate from
*Health checks: readiness gates*. Serve in the background, then wait for
whichever comes first: Serve failing, or `ctx` being cancelled.

### Closer

Start `srv.Serve(ln)` in a goroutine that sends its error on a channel with
room for one value, then `ready.Store(true)`. A `select` waits on that
channel and on `ctx.Done()`. In the `ctx.Done()` case: `ready.Store(false)`,
`log.Print("draining")`, then `srv.Shutdown` with a context from
`context.WithTimeout(context.Background(), grace)`. If `Shutdown` fails,
`srv.Close()` and wrap its error.

### Answer

```go
	serveErr := make(chan error, 1)
	go func() {
		serveErr <- srv.Serve(ln)
	}()
	ready.Store(true)

	select {
	case err := <-serveErr:
		ready.Store(false)
		return err
	case <-ctx.Done():
	}

	ready.Store(false)
	log.Print("draining")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), grace)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		srv.Close()
		return fmt.Errorf("forced shutdown: %w", err)
	}
	return nil
```

Readiness fails before `Shutdown`, so the load balancer stops sending new
requests while the ones already here finish.

## Task 2: Write the Dockerfile

Two stages, one `FROM` each. Write the build stage top to bottom in the
order of the numbered rules, then the final stage, which only copies the
binary in and says how to run it.

### Closer

The build stage: `FROM ... AS build`, `WORKDIR`, `COPY go.mod go.sum ./`,
`RUN go mod download`, `COPY . .`, `ARG VERSION=dev`, then the exact
`RUN CGO_ENABLED=0 go build ...` line. The final stage needs
`COPY --from=build /out/inventoryd /inventoryd`, the numeric `USER`,
`EXPOSE`, and `ENTRYPOINT ["..."]` with the same path.

### Answer

```dockerfile
FROM golang:1.26 AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
ARG VERSION=dev
RUN CGO_ENABLED=0 go build -ldflags "-X main.version=${VERSION}" -o /out/inventoryd .

FROM gcr.io/distroless/static-debian13
COPY --from=build /out/inventoryd /inventoryd
USER 65532:65532
EXPOSE 8080
ENTRYPOINT ["/inventoryd"]
```

## Task 3: Keep the build context small

One line per group in the table: the Git folder, the env files, the `bin`
folder, and the Markdown files. A folder's name alone leaves out
everything inside it.

### Closer

`*` after a name catches every file that starts with it, so one pattern
covers `.env` and `.env.local`. For files in any folder, start the pattern
with `**/`.

### Answer

```text
.git
.env*
bin/
**/*.md
```

## Task 4: Test every change

This is the workflow from *CI workflows*, with `-race` added to the tests:
`name`, `on` with two events, and one job whose four steps set up the code
and Go, then vet and test.

### Closer

`push:` gets `branches: [main]` under it, and `pull_request:` stays bare.
The setup-go step takes `with:` and `go-version-file: go.mod`. The last two
steps are `- run:` lines.

### Answer

```yaml
name: ci

on:
  push:
    branches: [main]
  pull_request:

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@3d3c42e5aac5ba805825da76410c181273ba90b1 # v7.0.1
      - uses: actions/setup-go@b7ad1dad31e06c5925ef5d2fc7ad053ef454303e # v7.0.0
        with:
          go-version-file: go.mod
      - run: go vet ./...
      - run: go test -race ./...
```

## Task 5: Release on tags

Start from your `ci.yml`: change the name and the trigger, keep the `test`
job as it is, and add the `image` job beside it under `jobs:`.

### Closer

The trigger is `push:` with `tags: ['v*']` under it. The `image` job gets
`needs: test`, a `permissions:` block with the two settings, and three
steps: checkout, `docker/login-action` with a `with:` block, and
`docker/build-push-action` with `context`, `push`, `tags` and
`build-args` under its `with:`. Every `uses` gets the SHA and release from
the README's pin table.

### Answer

```yaml
name: release

on:
  push:
    tags: ['v*']

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@3d3c42e5aac5ba805825da76410c181273ba90b1 # v7.0.1
      - uses: actions/setup-go@b7ad1dad31e06c5925ef5d2fc7ad053ef454303e # v7.0.0
        with:
          go-version-file: go.mod
      - run: go vet ./...
      - run: go test -race ./...

  image:
    needs: test
    runs-on: ubuntu-latest
    permissions:
      contents: read
      packages: write
    steps:
      - uses: actions/checkout@3d3c42e5aac5ba805825da76410c181273ba90b1 # v7.0.1
      - uses: docker/login-action@dbcb813823bdd20940b903addbd779551569679f # v4.6.0
        with:
          registry: ghcr.io
          username: ${{ github.actor }}
          password: ${{ secrets.GITHUB_TOKEN }}
      - uses: docker/build-push-action@c3c9e263c25d99ce0380d002d59b67737d91b0dc # v7.4.0
        with:
          context: .
          push: true
          tags: ghcr.io/acme/inventoryd:${{ github.ref_name }}
          build-args: VERSION=${{ github.ref_name }}
```
