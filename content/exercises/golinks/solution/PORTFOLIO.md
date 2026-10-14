# golinks

[![ci](https://github.com/octocat/golinks/actions/workflows/ci.yml/badge.svg)](https://github.com/octocat/golinks/actions/workflows/ci.yml)

Short links for a team: type `go/runbook-db` and land on the right page. A
small Go service with a SQLite store, packaged as a container and deployed
to Kubernetes with kustomize.

## Problem

During an incident, the on-call engineer needs the database runbook now,
not after five minutes of searching the wiki and three Slack threads. Long
URLs get lost in bookmarks and chat history, and they change when pages
move. A go link gives every important page one short, memorable name that a
team can update in one place, so `go/runbook-db` always works.

## Design

```text
browser ── GET /runbook-db ──► golinks ──► SQLite (/data/golinks.db, on a PVC)
                                  │
                       302 Found, Location: https://wiki…
```

golinks is one Go binary. `internal/links` holds the rules for a valid link
and the `Store` interface; `internal/store` implements it with SQLite
through the pure-Go `modernc.org/sqlite` driver, migrating the schema at
startup; `internal/server` holds the HTTP handlers. Every request writes one
JSON log line and increments a Prometheus counter by route and status, and
`/healthz` and `/readyz` drive the Kubernetes liveness and readiness probes.
On SIGTERM the server stops reporting ready, finishes in-flight requests
and exits. Settings come from `GOLINKS_*` environment variables, filled in
from a ConfigMap.

## Run it

```bash
make run    # or: go run ./cmd/golinks
curl -X POST localhost:8080/api/links -d '{"name":"go-docs","url":"https://go.dev/doc/"}'
curl -i localhost:8080/go-docs
```

```bash
kind create cluster --name golinks
make kind-load deploy     # docker build, kind load docker-image, kubectl apply -k deploy/overlays/dev
kubectl port-forward service/golinks 8080:80
```

## Trade-offs

- SQLite instead of Postgres: one file and no database server to run, but only one replica can write, so there is no high availability.
- One replica with the Recreate strategy, so every deploy has a few seconds of downtime while the old pod releases the volume.
- 302 instead of 301 redirects, so browsers never cache a link that someone later edits, at the cost of a request to golinks on every click.
- Metrics written by hand in the Prometheus text format instead of the client library, which keeps dependencies small but means no histograms yet.

## What I'd do next

- Add owners and an edit endpoint, so only the creator can change a link.
- Move to Postgres and run two replicas behind the Service.
- Add a request-duration histogram and a Grafana dashboard.
- Sign images with cosign and pin actions to commit SHAs.
