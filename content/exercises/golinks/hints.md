## Task 1: Validate new links

Four rules, checked in the README's order, each returning straight away.
The first three look at `l.Name`; the last one parses `l.URL` and looks at
its parts.

### Closer

A `switch` with no value after it runs the first `case` that's true. For the
name, `regexp.MustCompile` a pattern with `^` and `$` around it, and
`slices.Contains(Reserved, l.Name)`. For the URL, `u, err := url.Parse(l.URL)`,
then check `err`, `u.Scheme` and `u.Host`. Build each error with
`fmt.Errorf("%w: name is required", ErrInvalid)`.

### Answer

```go
var validName = regexp.MustCompile(`^[a-z0-9-]{1,64}$`)

func Validate(l Link) error {
	switch {
	case l.Name == "":
		return fmt.Errorf("%w: name is required", ErrInvalid)
	case !validName.MatchString(l.Name):
		return fmt.Errorf("%w: name must be at most 64 characters of a-z, 0-9 and -", ErrInvalid)
	case slices.Contains(Reserved, l.Name):
		return fmt.Errorf("%w: name is reserved", ErrInvalid)
	}
	u, err := url.Parse(l.URL)
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" {
		return fmt.Errorf("%w: url must be an absolute http or https URL", ErrInvalid)
	}
	return nil
}
```

## Task 2: Store links in SQLite

Start with the SQL file: one `CREATE TABLE` with two columns. Then `Create`
is one `INSERT` that does nothing on a taken name, and `Get` is one
`SELECT`. `Delete`, right below, shows how to read how many rows a
statement changed.

### Closer

The table: `name TEXT NOT NULL PRIMARY KEY` and `url TEXT NOT NULL`. In
`Create`, end the `INSERT` with `ON CONFLICT (name) DO NOTHING`, then
`res.RowsAffected()`: `0` means the name was taken. In `Get`, use
`s.db.QueryRowContext(ctx, ..., name).Scan(&l.URL)` and compare the error
with `errors.Is(err, sql.ErrNoRows)`.

### Answer

```sql
CREATE TABLE links (
    name TEXT NOT NULL PRIMARY KEY,
    url  TEXT NOT NULL
);
```

```go
func (s *SQLite) Create(ctx context.Context, l links.Link) error {
	res, err := s.db.ExecContext(ctx,
		`INSERT INTO links (name, url) VALUES (?, ?) ON CONFLICT (name) DO NOTHING`, l.Name, l.URL)
	if err != nil {
		return fmt.Errorf("create link %q: %w", l.Name, err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("create link %q: %w", l.Name, err)
	}
	if n == 0 {
		return links.ErrExists
	}
	return nil
}

func (s *SQLite) Get(ctx context.Context, name string) (links.Link, error) {
	l := links.Link{Name: name}
	err := s.db.QueryRowContext(ctx, `SELECT url FROM links WHERE name = ?`, name).Scan(&l.URL)
	if errors.Is(err, sql.ErrNoRows) {
		return links.Link{}, links.ErrNotFound
	}
	if err != nil {
		return links.Link{}, fmt.Errorf("get link %q: %w", name, err)
	}
	return l, nil
}
```

## Task 3: Serve the links API

`handleList` is the pattern: call the store with `r.Context()`, hand any
error to `s.writeStoreError`, and return. It already knows which error
means 400, 404, 409 or 500, so your handlers only decide what success
looks like.

### Closer

`r.PathValue("name")` reads the `{name}` part of the route. For the
redirect: `s.metrics.Redirect()`, then `http.Redirect(w, r, l.URL,
http.StatusFound)`. In `handleCreate`, validation errors also go through
`s.writeStoreError`, because they wrap `links.ErrInvalid`. For a 204, call
`w.WriteHeader(http.StatusNoContent)` and write nothing.

### Answer

```go
func (s *Server) handleRedirect(w http.ResponseWriter, r *http.Request) {
	l, err := s.store.Get(r.Context(), r.PathValue("name"))
	if err != nil {
		s.writeStoreError(w, r, err)
		return
	}
	s.metrics.Redirect()
	http.Redirect(w, r, l.URL, http.StatusFound)
}

// In handleCreate, after decodeJSON:
	if err := links.Validate(l); err != nil {
		s.writeStoreError(w, r, err)
		return
	}
	if err := s.store.Create(r.Context(), l); err != nil {
		s.writeStoreError(w, r, err)
		return
	}
	writeJSON(w, http.StatusCreated, l)

func (s *Server) handleDelete(w http.ResponseWriter, r *http.Request) {
	if err := s.store.Delete(r.Context(), r.PathValue("name")); err != nil {
		s.writeStoreError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
```

## Task 4: Report health and readiness

`/healthz` has nothing to check: it answers ok. `/readyz` has two reasons
to say no, checked in order: the server isn't serving (`s.ready`), or the
store doesn't answer its `Ping`.

### Closer

Both answers are `writeJSON(w, status, map[string]string{"status": ...})`.
For the ping, `ctx, cancel := context.WithTimeout(r.Context(), pingTimeout)`
with `defer cancel()`, then `s.store.Ping(ctx)`. Add `"context"` to the
imports.

### Answer

```go
func (s *Server) handleHealthz(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) handleReadyz(w http.ResponseWriter, r *http.Request) {
	if !s.ready.Load() {
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{"status": "not ready"})
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), pingTimeout)
	defer cancel()
	if err := s.store.Ping(ctx); err != nil {
		s.logger.WarnContext(r.Context(), "not ready: store ping failed", "error", err)
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{"status": "not ready"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ready"})
}
```

## Task 5: Log and count every request

Everything happens after `next.ServeHTTP` returns: by then the recorder
holds the status, and the mux has filled in `r.Pattern`. Work out the route
and the level first, then make one metrics call and one log call.

### Closer

Return `http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
... })`. Inside: `start := time.Now()`, `rec := &statusRecorder{ResponseWriter:
w, status: http.StatusOK}`, `next.ServeHTTP(rec, r)`. Log with
`s.logger.LogAttrs(r.Context(), level, "request", slog.String("method",
r.Method), ...)`, with `slog.Int` for the status and `slog.Int64` for
`time.Since(start).Milliseconds()`.

### Answer

```go
func (s *Server) observe(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(rec, r)

		route := r.Pattern
		if route == "" {
			route = "unmatched"
		}
		s.metrics.Request(route, rec.status)

		level := slog.LevelInfo
		if rec.status >= 500 {
			level = slog.LevelError
		}
		s.logger.LogAttrs(r.Context(), level, "request",
			slog.String("method", r.Method),
			slog.String("path", r.URL.Path),
			slog.String("route", route),
			slog.Int("status", rec.status),
			slog.Int64("duration_ms", time.Since(start).Milliseconds()),
		)
	})
}
```

## Task 6: Read settings from the environment

Build a `Config` holding the four defaults first. Then, for each variable,
read it once with `getenv`; if it isn't empty, parse it and either store it
or remember an error. Return the config and all the errors joined.

### Closer

`if v := getenv("GOLINKS_DB"); v != "" { cfg.DBPath = v }` is the shape
for each. A `switch v` maps the four level words to `slog.LevelDebug` and
friends. `time.ParseDuration(v)` reads `"25s"`. Collect errors in a
`[]error` and return `errors.Join(errs...)`, which is `nil` when the slice
is empty.

### Answer

```go
cfg := Config{Addr: ":8080", DBPath: "golinks.db", LogLevel: slog.LevelInfo, ShutdownGrace: 10 * time.Second}
var errs []error
if v := getenv("GOLINKS_ADDR"); v != "" {
	cfg.Addr = v
}
if v := getenv("GOLINKS_DB"); v != "" {
	cfg.DBPath = v
}
if v := getenv("GOLINKS_LOG_LEVEL"); v != "" {
	switch v {
	case "debug":
		cfg.LogLevel = slog.LevelDebug
	case "info":
		cfg.LogLevel = slog.LevelInfo
	case "warn":
		cfg.LogLevel = slog.LevelWarn
	case "error":
		cfg.LogLevel = slog.LevelError
	default:
		errs = append(errs, fmt.Errorf("GOLINKS_LOG_LEVEL: %q is not debug, info, warn or error", v))
	}
}
if v := getenv("GOLINKS_SHUTDOWN_GRACE"); v != "" {
	d, err := time.ParseDuration(v)
	switch {
	case err != nil:
		errs = append(errs, fmt.Errorf("GOLINKS_SHUTDOWN_GRACE: %w", err))
	case d <= 0:
		errs = append(errs, fmt.Errorf("GOLINKS_SHUTDOWN_GRACE: %s is not positive", v))
	default:
		cfg.ShutdownGrace = d
	}
}
return cfg, errors.Join(errs...)
```

## Task 7: Shut down gracefully

This is *Graceful shutdown: HTTP server and worker* again: serve in a
goroutine, then wait for whichever comes first, a `Serve` error or the end
of `ctx`. The ready flag goes up before serving and down on either way out.

### Closer

`serveErr := make(chan error, 1)`, `go func() { serveErr <- srv.Serve(ln) }()`,
then `select { case err := <-serveErr: ...; case <-ctx.Done(): }`. For the
grace period: `shutdownCtx, cancel := context.WithTimeout(context.Background(),
grace)`, `defer cancel()`, `srv.Shutdown(shutdownCtx)`. Add `"fmt"` to the
imports for the wrapped error.

### Answer

```go
srv := s.newHTTPServer()
s.ready.Store(true)
serveErr := make(chan error, 1)
go func() { serveErr <- srv.Serve(ln) }()

select {
case err := <-serveErr:
	s.ready.Store(false)
	return err
case <-ctx.Done():
}

s.ready.Store(false)
s.logger.Info("shutting down", "grace", grace.String())
shutdownCtx, cancel := context.WithTimeout(context.Background(), grace)
defer cancel()
if err := srv.Shutdown(shutdownCtx); err != nil {
	srv.Close()
	return fmt.Errorf("graceful shutdown: %w", err)
}
return nil
```

## Task 8: Write the Dockerfile

It's the Dockerfile from *Practice: containerize and ship a service* with
three changes: the package to build is `./cmd/golinks`, the binary is
called `golinks`, and the final stage uses the distroless `:nonroot` tag.

### Closer

Build stage: `FROM golang:1.26 AS build`, `WORKDIR /src`, `COPY go.mod
go.sum ./`, `RUN go mod download`, `COPY . .`, `ARG VERSION=dev`, then
`RUN CGO_ENABLED=0 go build -ldflags "-X main.version=${VERSION}" -o
/out/golinks ./cmd/golinks`. Final stage: `COPY --from=build`, `USER`,
`EXPOSE`, and an exec-form `ENTRYPOINT`.

### Answer

```dockerfile
FROM golang:1.26 AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
ARG VERSION=dev
RUN CGO_ENABLED=0 go build -trimpath -ldflags "-s -w -X main.version=${VERSION}" -o /out/golinks ./cmd/golinks

FROM gcr.io/distroless/static-debian13:nonroot
COPY --from=build /out/golinks /golinks
USER 65532:65532
EXPOSE 8080
ENTRYPOINT ["/golinks"]
```

## Task 9: Write the Kubernetes manifests

Write the two small files first: the ConfigMap is four settings under
`data:`, and the claim asks for 1Gi of ReadWriteOnce storage. Then copy the
`statusd` Deployment from the README and change it piece by piece against
the task's checklist.

### Closer

Names tie the files together: `envFrom: - configMapRef: name:
golinks-config`, and a volume with `persistentVolumeClaim: claimName:
golinks-data` mounted at `/data`. Quote `":8080"` in the ConfigMap, since
every value there must be a string. The probes use `port: http`, the name
of the container port.

### Answer

```yaml
# configmap.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: golinks-config
data:
  GOLINKS_ADDR: ":8080"
  GOLINKS_DB: /data/golinks.db
  GOLINKS_LOG_LEVEL: info
  GOLINKS_SHUTDOWN_GRACE: 20s
---
# pvc.yaml
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: golinks-data
spec:
  accessModes: [ReadWriteOnce]
  resources:
    requests:
      storage: 1Gi
---
# deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: golinks
  labels:
    app: golinks
spec:
  replicas: 1
  strategy:
    type: Recreate
  selector:
    matchLabels:
      app: golinks
  template:
    metadata:
      labels:
        app: golinks
    spec:
      securityContext:
        runAsNonRoot: true
      containers:
        - name: golinks
          image: ghcr.io/octocat/golinks:v0.1.0
          ports:
            - name: http
              containerPort: 8080
          envFrom:
            - configMapRef:
                name: golinks-config
          readinessProbe:
            httpGet:
              path: /readyz
              port: http
          livenessProbe:
            httpGet:
              path: /healthz
              port: http
          resources:
            requests:
              cpu: 50m
              memory: 64Mi
            limits:
              memory: 128Mi
          securityContext:
            allowPrivilegeEscalation: false
            readOnlyRootFilesystem: true
            capabilities:
              drop: [ALL]
          volumeMounts:
            - name: data
              mountPath: /data
      volumes:
        - name: data
          persistentVolumeClaim:
            claimName: golinks-data
```

## Task 10: Add a dev overlay

An overlay is a kustomization that lists the base as a resource, then says
what to change. Here that's two things: which image to run, and one
setting in the ConfigMap.

### Closer

`resources: - ../../base`. Under `images:`, an entry whose `name` is your
base image without the tag, with `newName: golinks` and `newTag: dev`.
Under `patches:`, `- path: configmap-patch.yaml`. The patch file is a
ConfigMap with only `apiVersion`, `kind`, `metadata.name` and the one key
under `data`.

### Answer

```yaml
# kustomization.yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
resources:
  - ../../base
images:
  - name: ghcr.io/octocat/golinks
    newName: golinks
    newTag: dev
patches:
  - path: configmap-patch.yaml
---
# configmap-patch.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: golinks-config
data:
  GOLINKS_LOG_LEVEL: debug
```

## Task 11: Test and ship with CI

It's the `ci` and `release` workflows from *Practice: containerize and ship
a service* in one file. The new parts are the extra `tags` under `push`,
the workflow-wide `permissions`, and an `if:` that runs the image job for
tags only.

### Closer

Under `on: push:` list both `branches: [main]` and `tags: ['v*']`. At the
top level, `permissions: contents: read`. The `image` job gets `needs:
test`, `if: startsWith(github.ref, 'refs/tags/v')`, its own `permissions:`
with `contents: read` and `packages: write`, then checkout, login and
build-push.

### Answer

```yaml
name: ci

on:
  push:
    branches: [main]
    tags: ['v*']
  pull_request:

permissions:
  contents: read

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v7
      - uses: actions/setup-go@v7
        with:
          go-version-file: go.mod
      - run: go vet ./...
      - run: go test -race ./...

  image:
    needs: test
    if: startsWith(github.ref, 'refs/tags/v')
    runs-on: ubuntu-latest
    permissions:
      contents: read
      packages: write
    steps:
      - uses: actions/checkout@v7
      - uses: docker/login-action@v4
        with:
          registry: ghcr.io
          username: ${{ github.actor }}
          password: ${{ secrets.GITHUB_TOKEN }}
      - uses: docker/build-push-action@v7
        with:
          context: .
          push: true
          tags: ghcr.io/octocat/golinks:${{ github.ref_name }}
          build-args: VERSION=${{ github.ref_name }}
```

## Task 12: Write your portfolio README

Write for a stranger who has two minutes: what problem, how it's built,
how to try it, and what you'd weigh differently with more time. Your notes
from tasks 2, 9 and 11 are most of the Trade-offs section already.

### Closer

Problem: a short story of the pain (an incident, a lost bookmark). Design:
a ```text diagram of the request path, then a paragraph naming each
package and what the service reports about itself. Run it: two ```bash
blocks. Trade-offs: "- X instead of Y: what you gain, what you give up."

### Answer

```markdown
## Trade-offs

- SQLite instead of Postgres: one file and no database server to run, but only one replica can write, so there is no high availability.
- One replica with the Recreate strategy, so every deploy has a few seconds of downtime while the old pod releases the volume.
- 302 instead of 301 redirects, so browsers never cache a link that someone later edits, at the cost of a request to golinks on every click.

## What I'd do next

- Add owners and an edit endpoint, so only the creator can change a link.
- Move to Postgres and run two replicas behind the Service.
- Add a request-duration histogram and a Grafana dashboard.
```

The other sections are yours to write: see `PORTFOLIO.md` in the solution
for a full example.
