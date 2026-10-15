# Capstone: go links, built and shipped

*Practice: cloud inventory report* closed *Cloud automation with the AWS SDK*,
the last new topic. The capstones add nothing big and new: they combine the
whole course into finished projects.

You'll build *go links*, the link shortener from the *System design
fundamentals* guide, and take it all the way: a tested Go service with a
SQLite store, a container image, Kubernetes manifests, a CI pipeline that
publishes the image, and a README that sells it. At the end it runs on a
real cluster on your laptop, and it's a project you can show an employer.

**New here:** `INSERT ... ON CONFLICT DO NOTHING`, `r.Pattern`, PersistentVolumeClaims. Everything else is practice from earlier modules, including `url.Parse` (*Container config: settings from the environment*), kustomize overlays (*Kustomize: one base, many environments*) and kind clusters (*Kubernetes fundamentals*).

## Learn

### What you're building

Type `go/runbook-db` in a browser and land on the database runbook. Here,
`go/` is the service's address; the name after it is looked up and
redirected.

| Request | Answer |
|---|---|
| `GET /{name}` | `302 Found` with `Location:` the link's URL, or 404 |
| `POST /api/links` | 201 and the new link, 400 if it's invalid, 409 if the name is taken |
| `GET /api/links` | 200, a JSON array of every link, sorted by name |
| `DELETE /api/links/{name}` | 204, or 404 |
| `GET /healthz` | liveness: 200 `{"status":"ok"}` |
| `GET /readyz` | readiness: 200 `{"status":"ready"}`, or 503 `{"status":"not ready"}` |
| `GET /metrics` | request and redirect counters, Prometheus text format |

A link in JSON is `{"name":"runbook-db","url":"https://wiki.example.com/runbooks/db"}`.
Errors are always `{"error":"..."}`.

### How it fits together

```text
             ┌──────────────────── golinks pod ─────────────────────┐
browser ──►  │ observe ──► ServeMux ──► handler ──► links.Store      │
 (Service)   │ (log+count)                          (SQLite)  ──► /data/golinks.db
             └──────────────────────────────────────────────────────┘     (PVC)
```

`internal/links` knows nothing about HTTP or SQL: it defines a `Link`, the
rules for a valid one, and the `Store` interface. `internal/store` is the
SQLite `Store`. `internal/server` turns HTTP into `Store` calls. The server
tests use a fake `Store`, so they never touch a database.

### File map

| File | What's in it | Task |
|---|---|---|
| `internal/links/links.go` | `Link`, the errors, `Validate`, the `Store` interface | 1 |
| `internal/store/migrations/001_create_links.sql` | the `links` table | 2 |
| `internal/store/sqlite.go` | `Open` (runs migrations), the SQLite `Store` | 2 |
| `internal/server/server.go` | `Server`, `New`, the routes | read only |
| `internal/server/handlers.go` | JSON helpers and the link handlers | 3 |
| `internal/server/health.go` | `/healthz` and `/readyz` | 4 |
| `internal/server/metrics.go` | the counters and `/metrics` | read only |
| `internal/server/middleware.go` | `observe`: logs and counts each request | 5 |
| `internal/config/config.go` | `Load`: settings from `GOLINKS_*` variables | 6 |
| `internal/server/run.go` | `Run`: serve, then shut down gracefully | 7 |
| `cmd/golinks/main.go` | wires it all up | read only |
| `Dockerfile` | the image | 8 |
| `Dockerfile.dockerignore` | what the image build leaves out | given |
| `deploy/base/` | Deployment, Service, ConfigMap, PVC, kustomization | 9 |
| `deploy/overlays/dev/` | the dev overlay | 10 |
| `ci.yml` | the CI workflow (`.github/workflows/ci.yml` once published) | 11 |
| `PORTFOLIO.md` | your project's README (`README.md` once published) | 12 |
| `Makefile` | `make run`, `make test`, `make image`, `make deploy`, ... | given |

Each task has one test: `TestValidate` (1), `TestStore` (2), `TestLinkAPI`
(3), `TestHealth` (4), `TestObserve` (5), `TestLoad` (6), `TestRun` (7), and
in the top folder `TestDockerfile` (8), `TestManifests` (9),
`TestDevOverlay` (10), `TestCIWorkflow` (11) and `TestPortfolio` (12). Every
test runs on its own, except where it says "finish task N first".

### New: inserting only if the key is free

In *CREATE TABLE: keys and constraints* a duplicate primary key made the
`INSERT` fail with an error. SQLite can instead skip the row quietly, and
`RowsAffected` tells you which happened:

```go
res, err := db.ExecContext(ctx,
	`INSERT INTO teams (slug, pager) VALUES (?, ?) ON CONFLICT (slug) DO NOTHING`,
	"payments", "pd-42")
if err != nil {
	return err
}
n, err := res.RowsAffected() // 1: inserted; 0: the slug was taken
```

No driver-specific error codes to decode, and no race between "check" and
"insert". One SQLite quirk: a `TEXT PRIMARY KEY` column accepts `NULL`
unless you also write `NOT NULL`.

`Open` (given) runs every `migrations/*.sql` file once, in name order, and
records it in a `schema_migrations` table, as in *Embedding files: database
migrations*. `":memory:"` gives the tests a fresh database each time.

### New: taking a URL apart

`url.Parse` from `net/url` splits a URL into fields. It's forgiving: text
without a scheme still parses, as a relative URL.

```go
u, err := url.Parse("https://grafana.example.com:3000/d/api?from=now-1h")
// u.Scheme == "https", u.Host == "grafana.example.com:3000", u.Path == "/d/api"
u, err = url.Parse("grafana.example.com/d/api")
// err == nil, but u.Scheme == "" and u.Host == "": it's only a path
```

So "it parsed" isn't "it's a link a browser can follow": check the fields.

### New: which route matched

`ServeMux` records the pattern it chose in `r.Pattern`, on the request it
was given. Middleware that wraps the mux can read it **after**
`next.ServeHTTP` returns:

```go
next.ServeHTTP(w, r)
fmt.Println(r.Pattern) // "GET /api/hosts/{id}", or "" if nothing matched (404, 405)
```

Metrics label by the pattern, never the path: `/runbook-db`, `/oncall`, ...
would each make a new time series, and Prometheus slows to a crawl when a
label has unbounded values.

### New: a Deployment you write by hand

In *Kubernetes objects: building a Deployment* you built these in Go. In a
repository they're YAML. Here is a different service, `statusd`, with the
production pieces this project needs:

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: statusd
spec:
  replicas: 1
  strategy:
    type: Recreate            # stop the old pod before starting the new one
  selector:
    matchLabels: {app: statusd}
  template:
    metadata:
      labels: {app: statusd}  # what the selector and the Service match
    spec:
      securityContext:
        runAsNonRoot: true    # refuse to start if the image runs as root
      containers:
        - name: statusd
          image: ghcr.io/acme/statusd:v2.3.0
          ports:
            - name: web
              containerPort: 9000
          envFrom:
            - configMapRef: {name: statusd-settings}  # every key becomes an env var
          readinessProbe:
            httpGet: {path: /ready, port: web}         # failing: no traffic
          livenessProbe:
            httpGet: {path: /live, port: web}          # failing: restart
          resources:
            requests: {cpu: 100m, memory: 32Mi}        # what the scheduler reserves
            limits: {memory: 64Mi}                     # above this: OOMKilled
          securityContext:
            allowPrivilegeEscalation: false
            readOnlyRootFilesystem: true
            capabilities: {drop: [ALL]}
          volumeMounts:
            - name: state
              mountPath: /var/lib/statusd
      volumes:
        - name: state
          persistentVolumeClaim: {claimName: statusd-state}
```

`{app: statusd}` is YAML's one-line form of a map; the block form you've
seen means the same. A *PersistentVolumeClaim* (PVC) asks the cluster for
a disk; the pod mounts it, and it outlives the pod:

```yaml
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: statusd-state
spec:
  accessModes: [ReadWriteOnce]   # one node mounts it read-write
  resources:
    requests:
      storage: 5Gi
```

`ReadWriteOnce` is why SQLite means **one replica** and `Recreate`: the
default rolling update starts the new pod while the old one still holds
the disk. More on this in
*Shipping it for real: clusters, registries and your portfolio*.

### New: kustomize bases and overlays

`kubectl apply -k <folder>` reads the folder's `kustomization.yaml`,
combines the files it lists, applies changes, and sends the result to the
cluster. A *base* holds what every environment shares; an *overlay* lists
the base and says what differs:

```yaml
# deploy/overlays/staging/kustomization.yaml, for statusd
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
resources:
  - ../../base                   # a folder with its own kustomization.yaml
images:
  - name: ghcr.io/acme/statusd   # find this image (any tag) ...
    newTag: v2.4.0-rc1           # ... and run this tag instead
patches:
  - path: more-memory.yaml       # merged into the object with the same kind and name
```

A patch is a partial object: `apiVersion`, `kind`, `metadata.name`, and
only the fields to change. `kubectl kustomize deploy/overlays/staging`
prints the result without applying it.

> **Why is it written like this?**
> - **Go requires** that `internal/` packages are imported only from inside
>   this module, so `internal/store` can't become someone else's API.
> - **Kubernetes requires** `apiVersion`, `kind` and `metadata.name` on
>   every object, and every ConfigMap value to be a string: `":8080"` needs
>   quotes, or YAML may read it differently.
> - **Convention:** `cmd/<binary>` plus `internal/`; settings from
>   environment variables with a common prefix; one manifest per file;
>   `deploy/base` and `deploy/overlays/<env>`.
> - **Author's choice:** the names (`golinks-config`, `golinks-data`), the
>   `GOLINKS_` prefix, `unmatched` as the route label for 404s and 405s,
>   and the port name `http`.

## Your task

Your team wants go links before the next incident. Build it task by task:
the Go service first (1–7), then everything that ships it (8–12). Run
`go test ./...` as you go; `make test` adds the race detector.

### 1. Validate new links

**Goal:** make `Validate` in `internal/links/links.go` return an error for the first rule a link breaks, or `nil`.

Check the rules **in this order**:

| Rule | Message |
|---|---|
| `Name` is not empty | `name is required` |
| `Name` is 1 to 64 characters, each `a`-`z`, `0`-`9` or `-` | `name must be at most 64 characters of a-z, 0-9 and -` |
| `Name` isn't in `Reserved` (`api`, `healthz`, `readyz`, `metrics`) | `name is reserved` |
| `URL` parses, its scheme is `http` or `https`, and it has a host | `url must be an absolute http or https URL` |

Every error wraps `ErrInvalid` with `%w`, so its full text is
`invalid link: ` and the message. You'll need the imports `fmt`, `net/url`,
`regexp` and `slices`.

```go
Validate(links.Link{Name: "runbook-db", URL: "https://wiki.example.com/db"})
// => nil
Validate(links.Link{Name: "Runbook", URL: "https://wiki.example.com/db"})
// => "invalid link: name must be at most 64 characters of a-z, 0-9 and -"
Validate(links.Link{Name: "docs", URL: "docs.example.com"})
// => "invalid link: url must be an absolute http or https URL"
```

**Think it through:**
- A regular expression without `^` and `$` matches anywhere inside the name. What would `Run book!` match?
- Why are `healthz` and `api` off limits? Look at the routes in `server.go`.

### 2. Store links in SQLite

**Goal:** create the `links` table in `internal/store/migrations/001_create_links.sql`, then write `Create` and `Get` in `internal/store/sqlite.go`.

The table `links` has two `TEXT` columns, both `NOT NULL`: `name`, the
primary key, and `url`.

- `Create(ctx, l)` inserts `l`. If a link with that name exists, it changes
  nothing and returns `links.ErrExists`. Wrap any other error with the name,
  such as `create link "docs": ...`.
- `Get(ctx, name)` returns the link. If there's no such row, it returns
  `links.ErrNotFound`.

`List` and `Delete` are written: use them as the pattern. Use `?`
placeholders, never string concatenation (*database/sql: fix the SQL
injection*).

```go
s, _ := store.Open(ctx, ":memory:")
s.Create(ctx, links.Link{Name: "docs", URL: "https://docs.example.com"}) // => nil
s.Create(ctx, links.Link{Name: "docs", URL: "https://evil.example.com"}) // => links.ErrExists
s.Get(ctx, "docs")  // => {docs https://docs.example.com}, nil
s.Get(ctx, "nope")  // => links.ErrNotFound
```

If you ran golinks before this task, delete `golinks.db`: it recorded the
empty migration as done.

**Think it through:**
- `Delete` learns "no such link" from `RowsAffected`. Which statement makes an `INSERT` report 0 rows instead of failing?
- Which error does `Scan` return when the query found no row?

### 3. Serve the links API

**Goal:** write `handleRedirect`, `handleCreate` and `handleDelete` in `internal/server/handlers.go`.

`handleList` shows the pattern. `s.writeStoreError(w, r, err)` already maps
errors to responses: `links.ErrNotFound` to 404, `links.ErrExists` to 409,
`links.ErrInvalid` to 400 with the error's text, anything else to 500
`internal error`.

| Handler | Situation | Status | Body |
|---|---|---|---|
| `GET /{name}` | link found | 302, header `Location:` its URL | (whatever `http.Redirect` writes) |
| | no such link | 404 | `{"error":"link not found"}` |
| `POST /api/links` | not JSON, or an unknown field (given) | 400 | `{"error":"invalid JSON body"}` |
| | fails `links.Validate` | 400 | `{"error":"invalid link: name is reserved"}` (its message) |
| | name taken | 409 | `{"error":"link already exists"}` |
| | created | 201 | the link as JSON |
| `DELETE /api/links/{name}` | deleted | 204 | none |
| | no such link | 404 | `{"error":"link not found"}` |

Before each redirect, call `s.metrics.Redirect()` once. Read the name from
the path with `r.PathValue("name")`, and redirect with
`http.Redirect(w, r, url, http.StatusFound)`.

```console
$ curl -i -X POST localhost:8080/api/links -d '{"name":"oncall","url":"https://pager.example.com/schedules/platform"}'
HTTP/1.1 201 Created
{"name":"oncall","url":"https://pager.example.com/schedules/platform"}
$ curl -i localhost:8080/oncall
HTTP/1.1 302 Found
Location: https://pager.example.com/schedules/platform
```

**Think it through:**
- Why `302 Found` and not `301 Moved Permanently`? The *System design fundamentals* guide's worked example answers it.
- Which handler can skip `writeJSON` entirely?

### 4. Report health and readiness

**Goal:** write `handleHealthz` and `handleReadyz` in `internal/server/health.go`.

- `GET /healthz` answers 200 `{"status":"ok"}`, always.
- `GET /readyz` answers 503 `{"status":"not ready"}` when `s.ready` is
  false. Otherwise it calls `s.store.Ping` with a context derived from
  `r.Context()` that times out after `pingTimeout` (one second): an error
  means 503 `{"status":"not ready"}`, success means 200
  `{"status":"ready"}`.

`Run` (task 7) sets `s.ready`. Add the `context` import.

**Think it through:**
- If the database is down, would restarting golinks help? That decides which probe checks it.
- `context.WithTimeout` returns a cancel function too. What do you do with it?

### 5. Log and count every request

**Goal:** make `observe` in `internal/server/middleware.go` serve each request, then count it in the metrics and write one log line.

1. Wrap `w` in a `statusRecorder` whose `status` starts at `http.StatusOK`,
   note the start time, and serve the request with `next.ServeHTTP`.
2. The *route* is `r.Pattern`, or `"unmatched"` when that's empty.
3. Call `s.metrics.Request(route, status)`.
4. Write one log line with `s.logger`, message `"request"`, at level
   `ERROR` if the status is 500 or more, otherwise `INFO`, with these
   fields:

| Field | Value |
|---|---|
| `method` | `r.Method` |
| `path` | `r.URL.Path` |
| `route` | the route |
| `status` | the recorded status, as a number |
| `duration_ms` | whole milliseconds since the start, as a number |

```json
{"time":"…","level":"INFO","msg":"request","method":"GET","path":"/oncall","route":"GET /{name}","status":302,"duration_ms":0}
```

and `/metrics` then shows:

```text
golinks_http_requests_total{route="GET /{name}",code="302"} 1
```

**Think it through:**
- When is `r.Pattern` filled in: before or after `next.ServeHTTP`?
- `slog.Int64` takes an `int64`. What does `Duration.Milliseconds()` return?

### 6. Read settings from the environment

**Goal:** write `Load` in `internal/config/config.go`: build a `Config` from the variables `getenv` returns.

| Variable | Field | Default | Valid values |
|---|---|---|---|
| `GOLINKS_ADDR` | `Addr` | `:8080` | any |
| `GOLINKS_DB` | `DBPath` | `golinks.db` | any |
| `GOLINKS_LOG_LEVEL` | `LogLevel` | `slog.LevelInfo` | `debug`, `info`, `warn`, `error` |
| `GOLINKS_SHUTDOWN_GRACE` | `ShutdownGrace` | `10 * time.Second` | a positive duration such as `20s` |

An empty value counts as unset. For an invalid value, return an error that
**names the variable**, like `GOLINKS_LOG_LEVEL: "loud" is not debug, info,
warn or error`. Report every invalid variable, not just the first:
collect the errors and return `errors.Join(errs...)`.

```go
Load(func(k string) string { return map[string]string{"GOLINKS_LOG_LEVEL": "debug"}[k] })
// => Config{Addr: ":8080", DBPath: "golinks.db", LogLevel: slog.LevelDebug, ShutdownGrace: 10s}, nil
```

**Think it through:**
- `time.ParseDuration("30")` fails: a duration needs a unit. What about `"0s"`?
- Why take `getenv` as a parameter instead of calling `os.Getenv`?

### 7. Shut down gracefully

**Goal:** finish `Run(ctx, ln, grace)` in `internal/server/run.go` so it reports ready, serves, and when `ctx` ends drains in-flight requests within `grace`.

1. Store `true` in `s.ready`, then run `srv.Serve(ln)` in a goroutine that
   sends its error on a channel with a buffer of 1.
2. `select` on that channel and `ctx.Done()`:
   - `Serve` returned first: store `false` in `s.ready`, return its error.
   - `ctx` ended: store `false` in `s.ready`, log `"shutting down"` at
     Info level with the attribute `"grace"` set to `grace.String()`, and
     call `srv.Shutdown` with a **new** context from
     `context.Background()` that times out after `grace`.
3. If `Shutdown` fails (the grace period ran out), call `srv.Close()` and
   return the error wrapped as `graceful shutdown: <err>` with `%w`.
   Otherwise return `nil`.

Kubernetes sends `SIGTERM`, which `main.go` turns into a cancelled `ctx`.
Readiness fails at once, so the Service stops sending new requests while
the old ones finish.

**Think it through:**
- `ctx` is already cancelled when `Shutdown` runs. What would `Shutdown(ctx)` do?
- Why must the channel be buffered?

### 8. Write the Dockerfile

**Goal:** in `Dockerfile`, build golinks in a build stage and run it from a distroless final stage as a non-root user.

The build stage:

1. `FROM golang:1.26 AS build` (any stage name works).
2. Set a `WORKDIR` such as `/src`. Copy `go.mod` and `go.sum`, `RUN go mod
   download`, then copy everything with `COPY . .`, in that order.
3. `ARG VERSION=dev`, before the build.
4. One `RUN` that builds `./cmd/golinks` with `CGO_ENABLED=0`, stamps the
   version with `-ldflags "-X main.version=${VERSION}"` (double quotes;
   extra flags like `-s -w` may go inside them), and writes the binary
   with `-o`, such as `-o /out/golinks`.

The final stage:

5. `FROM gcr.io/distroless/static-debian13:nonroot` (any tag works).
6. `COPY --from=build` the binary from its `-o` path to, say, `/golinks`.
7. `USER 65532:65532`, `EXPOSE 8080`, and an exec-form `ENTRYPOINT` that
   runs the binary where you copied it.

Exactly two `FROM` lines. `modernc.org/sqlite` is pure Go, so the static
binary needs nothing from the image.

```bash
docker build -t golinks:dev --build-arg VERSION=v0.1.0 .
docker run --rm -p 8080:8080 -e GOLINKS_DB=/tmp/golinks.db golinks:dev
```

**Think it through:**
- Which `COPY` changes on every commit? It goes last, after the module download.
- The final stage has no shell. Which `ENTRYPOINT` form still works?

### 9. Write the Kubernetes manifests

**Goal:** write `configmap.yaml`, `pvc.yaml` and `deployment.yaml` in `deploy/base`, so they fit the given `service.yaml` and `kustomization.yaml`.

`configmap.yaml`: a ConfigMap `golinks-config` whose `data` has
`GOLINKS_ADDR: ":8080"`, `GOLINKS_DB: /data/golinks.db`,
`GOLINKS_LOG_LEVEL: info` and `GOLINKS_SHUTDOWN_GRACE: 20s`.

`pvc.yaml`: a PersistentVolumeClaim `golinks-data`, access mode
`ReadWriteOnce`, requesting `storage: 1Gi`.

`deployment.yaml`: a Deployment `golinks` (`apps/v1`) with:

- [ ] `replicas: 1` and `strategy.type: Recreate`
- [ ] pod template label `app: golinks`, and `selector.matchLabels` that matches it
- [ ] one container `golinks`, image `ghcr.io/<you>/golinks:v0.1.0`, where
      `<you>` is your GitHub username in lowercase (no `:latest`; leave out
      `imagePullPolicy` or set `IfNotPresent`)
- [ ] one port: `containerPort: 8080` named `http`
- [ ] `envFrom` with `configMapRef` `golinks-config`
- [ ] `readinessProbe` `httpGet` path `/readyz`, and `livenessProbe` path
      `/healthz`, both on port `http`
- [ ] `resources.requests` with `cpu` and `memory`, and `limits.memory` at
      least the request
- [ ] `runAsNonRoot: true` (pod or container `securityContext`), and on the
      container `allowPrivilegeEscalation: false`,
      `readOnlyRootFilesystem: true`, `capabilities: drop: [ALL]`
- [ ] a volume from the claim `golinks-data`, mounted at `/data`

`GOLINKS_SHUTDOWN_GRACE` must stay below the pod's
`terminationGracePeriodSeconds` (30 by default): after that, Kubernetes
kills the process mid-drain. The test also checks that the Service selects
your pods and reaches port 8080.

**Think it through:**
- Three names tie four files together. Which file mentions which name?
- With `readOnlyRootFilesystem`, where can SQLite write?

### 10. Add a dev overlay

**Goal:** in `deploy/overlays/dev`, write `kustomization.yaml` and `configmap-patch.yaml` so the dev cluster runs your local image with debug logs.

`kustomization.yaml`:

- `apiVersion: kustomize.config.k8s.io/v1beta1` and `kind: Kustomization`
- `resources` with `../../base`
- `images` with one entry: `name` is your base image without its tag
  (`ghcr.io/<you>/golinks`), `newName: golinks`, `newTag: dev`
- `patches` with `- path: configmap-patch.yaml`

`configmap-patch.yaml`: a ConfigMap named `golinks-config` whose `data`
has only `GOLINKS_LOG_LEVEL: debug`.

The test reads your base Deployment's image, so finish task 9 first.

```bash
kubectl kustomize deploy/overlays/dev | grep -E 'image:|LOG_LEVEL'
#   GOLINKS_LOG_LEVEL: debug
#         image: golinks:dev
```

**Think it through:**
- `golinks:dev` only exists on your laptop and in kind. Why does `imagePullPolicy` matter here?
- Why list only one key in the patch?

### 11. Test and ship with CI

**Goal:** in `ci.yml`, write the workflow `ci` that tests every change and publishes the image when you push a version tag.

- `name: ci`
- `on`: `push` with `branches: [main]` and `tags: ['v*']`, and `pull_request`
- top-level `permissions` with only `contents: read`
- job `test` on `ubuntu-latest`: pinned `actions/checkout`, pinned
  `actions/setup-go` with `go-version-file: go.mod`, then steps running
  `go vet ./...` and `go test -race ./...`, in that order
- job `image` on `ubuntu-latest`:
  - `needs: test` and `if: startsWith(github.ref, 'refs/tags/v')`
  - `permissions`: exactly `contents: read` and `packages: write`
  - steps, all pinned: `actions/checkout`; `docker/login-action` with
    `registry: ghcr.io`, `username: ${{ github.actor }}`,
    `password: ${{ secrets.GITHUB_TOKEN }}`; `docker/build-push-action`
    with `context: .`, `push: true`, `tags:
    ghcr.io/<you>/golinks:${{ github.ref_name }}` and `build-args:
    VERSION=${{ github.ref_name }}`

Pin every action as in *CI workflows*: the full commit SHA, then
`# vX.Y.Z`. The test checks every `uses` line. The pins, the same as in
*Practice: containerize and ship a service*:

| Action | Release | Commit |
|---|---|---|
| `actions/checkout` | v7.0.1 | `3d3c42e5aac5ba805825da76410c181273ba90b1` |
| `actions/setup-go` | v7.0.0 | `b7ad1dad31e06c5925ef5d2fc7ad053ef454303e` |
| `docker/login-action` | v4.6.0 | `dbcb813823bdd20940b903addbd779551569679f` |
| `docker/build-push-action` | v7.4.0 | `c3c9e263c25d99ce0380d002d59b67737d91b0dc` |
| `docker/setup-buildx-action` (optional) | v4.4.1 | `f87e5991a6d7451dcb8d9637bfbc97413f497069` |

**Think it through:**
- A pull request also runs this workflow. What stops it from publishing an image?
- Why does the workflow default to `contents: read` when one job needs more?

### 12. Write your portfolio README

**Goal:** fill in `PORTFOLIO.md`, the README a stranger reads first when they open your repository.

Keep the template's headings. The test drops `<!-- comments -->`, then
checks:

| Part | Rule |
|---|---|
| everywhere | no `TODO`, no placeholders like `<you>` |
| the badge | the CI badge line, with your GitHub username |
| `## Problem` | at least 40 words |
| `## Design` | at least 60 words outside code blocks, and a ``` code block (a diagram) |
| `## Run it` | code blocks with `go run` or `make run`, and with `kubectl apply -k` or `make deploy` |
| `## Trade-offs` | at least 3 bullet points (`- `), each at least 8 words |
| `## What I'd do next` | at least 3 bullet points, each at least 4 words |

**Think it through:**
- Which decisions in tasks 2, 9 and 11 cost you something? Those are your trade-offs.
- What would you change to run golinks for 10,000 people?

## Run it for real (optional)

The tests read your files; this runs them. You need Docker, `kubectl` and
*kind* (Kubernetes in Docker). On NixOS, `nix shell nixpkgs#kind` opens a
shell with `kind` in it. Run everything from this folder.

```bash
nix shell nixpkgs#kind
kind create cluster --name golinks          # also switches kubectl to it
docker build -t golinks:dev --build-arg VERSION=v0.1.0 .
kind load docker-image golinks:dev --name golinks
kubectl apply -k deploy/overlays/dev
kubectl rollout status deployment/golinks --timeout=120s
kubectl port-forward service/golinks 8080:80   # leave it running
```

In a second terminal:

```bash
curl -i -X POST localhost:8080/api/links -d '{"name":"runbook-db","url":"https://wiki.example.com/runbooks/db"}'
curl -i localhost:8080/runbook-db             # 302, Location: https://wiki.example.com/runbooks/db
curl -s localhost:8080/metrics
kubectl logs deployment/golinks
kubectl delete pod -l app=golinks             # the Deployment starts a new pod
```

Restart the port-forward once the new pod is ready, and `curl -i
localhost:8080/runbook-db` still redirects: the link lives on the volume,
not in the pod. (`make kind-up kind-load deploy` runs the same steps, with
the cluster name `golinks`.)

If something's stuck, `kubectl get pods`, then `kubectl describe pod -l
app=golinks` and read the Events at the bottom. `ErrImagePull` means kind
doesn't have the image: check the name in your overlay and rerun
`kind load docker-image`.

If `kind create cluster` fails with `could not find a log line that
matches "Reached target .*Multi-User System.*"`, your Docker daemon logs to
journald in a way `docker logs` can't read back. On NixOS, set
`virtualisation.docker.logDriver = "json-file";` in your configuration and
run `sudo nixos-rebuild switch`.

Clean up when you're done:

```bash
kind delete cluster --name golinks
docker rmi golinks:dev
```

## Publish it

1. Create an empty **public** repository called `golinks` on GitHub (no
   README, no licence: you bring your own).
2. Copy this folder into a new one, and turn it into the project:

   ```bash
   cp -r . ~/code/golinks && cd ~/code/golinks
   mv PORTFOLIO.md README.md                  # replaces these instructions
   mkdir -p .github/workflows && mv ci.yml .github/workflows/ci.yml
   rm ./*_test.go                             # the course's checks of these files
   go mod edit -module github.com/<you>/golinks
   grep -rl groundwork.local/golinks --include='*.go' . | xargs sed -i 's#groundwork.local/golinks#github.com/<you>/golinks#'
   go mod tidy && go test ./...
   printf 'bin/\n*.db\n' > .gitignore
   ```

   Your actions are pinned to commit SHAs, so nothing updates them by
   itself. Let Dependabot open a pull request when a new release comes out;
   it changes the SHA and the `# vX.Y.Z` comment together:

   ```bash
   cat > .github/dependabot.yml <<'EOF'
   version: 2
   updates:
     - package-ecosystem: github-actions
       directory: /
       schedule:
         interval: weekly
   EOF
   ```

3. Commit and push:

   ```bash
   git init -b main && git add . && git commit -m "golinks: a go links service"
   git remote add origin git@github.com:<you>/golinks.git
   git push -u origin main
   ```

4. Open the repository's **Actions** tab: `ci` runs `test` and skips
   `image`. Fix anything red before going on.
5. Release: `git tag v0.1.0 && git push origin v0.1.0`. This time `image`
   runs and pushes `ghcr.io/<you>/golinks:v0.1.0`. It appears under your
   profile's **Packages**; in its settings, set the visibility to public so
   any cluster can pull it.
6. Check that the badge at the top of your README is green, then pin the
   repository on your GitHub profile (**Customize your pins**).

## Learn more

- [SQLite: the ON CONFLICT clause of INSERT (upsert)](https://www.sqlite.org/lang_upsert.html)
- [net/http: Request.Pattern and ServeMux patterns](https://pkg.go.dev/net/http#ServeMux)
- [Kubernetes: Persistent Volumes](https://kubernetes.io/docs/concepts/storage/persistent-volumes/)
- [Declarative management with kustomize](https://kubernetes.io/docs/tasks/manage-kubernetes-objects/kustomization/)
- [kind: quick start](https://kind.sigs.k8s.io/docs/user/quick-start/)
