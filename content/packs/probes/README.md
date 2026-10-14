# Probes: readiness, liveness and startup

In *Health checks: liveness and readiness* you wrote `/healthz` and
`/readyz` handlers in Go. *Probes* are the other half: they tell
Kubernetes to call those endpoints, and what to do when they fail. Get
them wrong and Kubernetes restarts healthy apps, or kills slow ones before
they start.

**New here:** `readinessProbe`, `livenessProbe`, `startupProbe`, `httpGet` probes, probe timing (`periodSeconds`, `failureThreshold`, `initialDelaySeconds`), reading probe failures in `describe`.

## Learn

### Three probes, three consequences

The *kubelet* (the Kubernetes agent on each node) runs every probe on a
timer, from outside the container.

| Probe | Question | When it fails |
|---|---|---|
| `readinessProbe` | Can you take traffic right now? | The Pod is taken out of its Services' endpoints. Nothing restarts. |
| `livenessProbe` | Are you stuck beyond repair? | The container is killed and restarted. |
| `startupProbe` | Have you finished starting? | Until it succeeds, the other two wait. If it never does, the container is restarted. |

READY `0/1` with STATUS `Running` in `kubectl get pods` is a failing
readiness probe. A RESTARTS count that keeps climbing on a `Running` Pod is
often a failing liveness probe.

### Writing a probe

```yaml
      containers:
        - name: docs
          image: nginx:1.27
          ports:
            - name: web
              containerPort: 8080
          readinessProbe:
            httpGet:              # healthy = any status from 200 to 399
              path: /ready
              port: web           # a container port's name or number
            periodSeconds: 5      # check every 5 seconds
          livenessProbe:
            httpGet:
              path: /alive
              port: web
            failureThreshold: 3   # restart after 3 failures in a row
```

`httpGet` is the usual kind for web apps. There are others: `tcpSocket`
(can I connect?) and `exec` (does this command exit 0?).

### Timing

| Field | Default | Meaning |
|---|---|---|
| `initialDelaySeconds` | 0 | Wait this long after the container starts |
| `periodSeconds` | 10 | Time between checks |
| `timeoutSeconds` | 1 | A check that takes longer counts as failed |
| `failureThreshold` | 3 | Failures in a row before the probe counts as failed |

So with the defaults, a liveness probe restarts a container after about
`periodSeconds × failureThreshold` = 30 seconds of failures. An app that
needs a minute to start is killed before it ever listens, again and again.

A *startup probe* fixes that without slowing down liveness for the rest of
the Pod's life: it allows `periodSeconds × failureThreshold` seconds to
start (`10 × 30` = 300s, say), and only then hands over to the liveness
probe. Giving liveness a long `initialDelaySeconds` instead works, but
delays every restart, even fast ones.

### Reading probe failures

`kubectl describe pod` shows the probes as Kubernetes understood them, and
an event for every kind of failure:

```text
    Liveness:   http-get http://:web/alive delay=0s timeout=1s period=10s successThreshold=1 failureThreshold=3
...
  Warning  Unhealthy  10s (x9 over 90s)  kubelet  Liveness probe failed: HTTP probe failed with statuscode: 404
  Normal   Killing    10s (x3 over 70s)  kubelet  Container docs failed liveness probe, will be restarted
```

| Message | Usual cause |
|---|---|
| `statuscode: 404` | Wrong `path` |
| `statuscode: 503` | The app answers, and says it isn't ready or healthy |
| `connect: connection refused` | Nothing listens on that port: wrong `port`, or the app hasn't started yet |
| `context deadline exceeded` | The app answered slower than `timeoutSeconds` |

> **Why is it written like this?**
> - **Kubernetes requires** exactly one way to check (`httpGet`,
>   `tcpSocket` or `exec`) per probe.
> - **Convention:** liveness checks only "is this process working", never
>   the database. If the database is down, restarting every app doesn't fix
>   it; failing readiness takes the Pods out of rotation instead.
> - **Your choice:** the timings. Tune them from how long your app really
>   takes, not from guesses.

## Your task

podinfo serves `/healthz` (liveness) and `/readyz` (readiness) on its port
named `http`, 9898. Three Deployments in `shop` need their probes fixed.

### 1. Add a readiness probe

**Goal:** in `1_web.yaml`, add a `readinessProbe` that sends an HTTP GET
to `/readyz` on the port named `http`.

```text
$ kubectl -n shop describe pod -l app=web | grep Readiness
    Readiness:  http-get http://:http/readyz delay=0s timeout=1s period=10s successThreshold=1 failureThreshold=3
```

**Think it through:**
- The probe belongs to the container, at the same level as `ports`.
- Which two fields does `httpGet` need?

### 2. Stop the restarts

**Goal:** fix the `livenessProbe` in `2_cart.yaml` so healthy cart Pods
stop being restarted.

The cart Pods restart about every 30 seconds. `testdata/cart_describe.txt`
has the evidence.

**Think it through:**
- Which status code does the probe get? What does that code mean?
- Compare the probe's path with the endpoints podinfo serves.

### 3. Give search time to start

**Goal:** add a `startupProbe` to `3_search.yaml` that checks `/healthz`
on port `http` and allows search **at least 120 seconds** to start.

search loads an index for a minute before it listens. It never becomes
ready, and `testdata/search_describe.txt` shows why. Keep the liveness and
readiness probes as they are.

```text
initialDelaySeconds + periodSeconds × failureThreshold >= 120
```

Fields you leave out take their defaults (0, 10 and 3).

**Think it through:**
- After how many seconds does the current liveness probe kill search?
  How long does search need?
- Which `periodSeconds` and `failureThreshold` give at least 120 seconds?

## Run it for real

Optional. Set up kind as in *Pods: your first manifest*. From the exercise
folder:

```bash
nix shell nixpkgs#kind
kind create cluster --name gw-probes
kubectl create namespace shop
kubectl apply -f 1_web.yaml -f 2_cart.yaml -f 3_search.yaml
kubectl -n shop get pods -w    # search: Running 0/1 for a minute, then 1/1; Ctrl+C
kubectl -n shop describe pod -l app=search | grep -A3 Startup
```

Now break one on purpose: change cart's liveness path back to `/health`,
apply, and run `kubectl -n shop get pods -l app=cart -w` until RESTARTS
climbs. Then clean up:

```bash
kind delete cluster --name gw-probes
```

## Learn more

- [Kubernetes: Liveness, readiness and startup probes](https://kubernetes.io/docs/concepts/configuration/liveness-readiness-startup-probes/)
- [Kubernetes: Configure liveness, readiness and startup probes](https://kubernetes.io/docs/tasks/configure-pod-container/configure-liveness-readiness-startup-probes/)
