# Rollouts: surge, unavailable and rollback

Change a Deployment's image and Kubernetes replaces its Pods a few at a
time, so the app stays up. Two numbers control how fast and how safely
that happens, and every change is kept as a *revision* you can roll back
to. Here you do the arithmetic Kubernetes does, and find the revision to
go back to.

**New here:** rolling updates, `maxSurge` and `maxUnavailable` and their rounding, rounding up with integer division, revisions and `kubernetes.io/change-cause`, `kubectl rollout status`/`history`/`undo`.

## Learn

### A rolling update

Change anything in a Deployment's Pod `template` (usually the image) and
the Deployment controller starts a *rollout*:

```text
ReplicaSet docs-v1: 4 ───> 3 ───> 2 ───> 1 ───> 0
ReplicaSet docs-v2: 0 ───> 1 ───> 2 ───> 3 ───> 4
```

It creates a new ReplicaSet for the new template, scales it up, and scales
the old one down, step by step. It only takes the next step when new Pods
are **ready**, which is why readiness probes (*Probes: readiness, liveness
and startup*) make rollouts safe: a new version that never gets ready
stops the rollout instead of replacing every working Pod.

### maxSurge and maxUnavailable

```yaml
spec:
  replicas: 10
  strategy:
    type: RollingUpdate
    rollingUpdate:
      maxSurge: 25%         # how many extra Pods may exist during the rollout
      maxUnavailable: 25%   # how many of the 10 may be missing
```

Each is a whole number or a percentage of `replicas`; both default to
`25%`. Percentages don't divide evenly, so Kubernetes rounds them, in the
safe direction for each:

| Setting | Rounds | 10 replicas, 25% | Meaning |
|---|---|---|---|
| `maxSurge` | **up** | 2.5 → 3 | at most 10 + 3 = **13** Pods at once |
| `maxUnavailable` | **down** | 2.5 → 2 | at least 10 − 2 = **8** available |

If both come to 0, no step would ever be allowed, so Kubernetes sets
`maxUnavailable` to 1. `maxSurge: 1, maxUnavailable: 0` is the careful
choice: one extra Pod at a time, never below full strength. `type:
Recreate` instead stops all old Pods before starting new ones (downtime,
but never two versions at once).

### Rounding with whole numbers

Go's integer division rounds down: `10 * 25 / 100` is `2`. To round up,
add "almost one" before dividing:

```go
total, percent := 7, 50
down := total * percent / 100        // 350 / 100 = 3
up := (total*percent + 99) / 100     // 449 / 100 = 4
```

Multiply before dividing: `25 / 100 * 10` is `0`.

`strings.CutSuffix(s, "%")` returns `s` without a trailing `%` and whether
it had one, as in *Custom types: Kubernetes memory quantities*:

```go
strings.CutSuffix("50%", "%") // => "50", true
strings.CutSuffix("3", "%")   // => "3", false
```

### Watching and undoing

```bash
kubectl set image deployment/docs docs=registry.example.com/wiki/docs:2.4.0
kubectl annotate deployment/docs kubernetes.io/change-cause="deploy docs:2.4.0"
kubectl rollout status deployment/docs --timeout=5m  # exits non-zero if it fails
kubectl rollout history deployment/docs
kubectl rollout undo deployment/docs                  # back to the previous revision
kubectl rollout undo deployment/docs --to-revision=3  # back to a chosen one
```

`kubernetes.io/change-cause` is the text the history shows next to each
revision; set it when you deploy, or every line says `<none>`. Here is real
output after rolling back from revision 3 to revision 2:

```text
deployment.apps/web
REVISION  CHANGE-CAUSE
1         <none>
3         image podinfo:6.9.2
4         image podinfo:6.9.1
```

**Undo makes a new revision.** Revision 2 became revision 4; the number 2
is gone. So "the revision before the newest" isn't always the version you
want: look it up by what it deployed. A rollout that makes no progress for
10 minutes (`progressDeadlineSeconds: 600`) is marked failed, and
`kubectl rollout status` says `error: deployment "docs" exceeded its
progress deadline`. Kubernetes does **not** roll back on its own: that's
your call, or your pipeline's.

## Your task

Write the helpers for a deploy bot in `exercise.go`.

### 1. Resolve a surge or unavailable value

**Goal:** make `Resolve(value, replicas, roundUp)` return the number of
Pods `value` stands for.

- `"N"`: the number N.
- `"N%"`: N percent of `replicas`, rounded **up** if `roundUp` is true and
  **down** otherwise.
- Anything else, or a negative N: an error.

```go
Resolve("2", 10, true)    // => 2, nil
Resolve("25%", 10, true)  // => 3, nil   (2.5 rounded up)
Resolve("25%", 10, false) // => 2, nil   (2.5 rounded down)
Resolve("two", 10, true)  // => 0, error
```

**Think it through:**
- After `strings.CutSuffix`, both forms leave a string to convert with
  `strconv.Atoi`. What does the second return value tell you?
- How do you round up with integer division?

### 2. Work out the rollout bounds

**Goal:** make `Bounds(replicas, maxSurge, maxUnavailable)` return the
most Pods at once and the fewest available Pods during a rollout.

- Resolve `maxSurge` rounding **up** and `maxUnavailable` rounding
  **down**. Return any error.
- If both come to 0, use 1 for `maxUnavailable`.
- `maxPods` = replicas + surge; `minAvailable` = replicas − unavailable,
  but never below 0.

```go
Bounds(10, "25%", "25%") // => 13, 8, nil
Bounds(1, "25%", "25%")  // => 2, 1, nil
Bounds(3, "0", "0")      // => 3, 2, nil
```

**Think it through:**
- With one replica and the defaults, why must surge round up?
- Which built-in function keeps a number from going below 0?

### 3. Find the revision to roll back to

**Goal:** make `RevisionFor(history, image)` return the newest revision
whose CHANGE-CAUSE contains `image` as a whole word.

`testdata/history.txt` is real-looking `kubectl rollout history` output:

```go
RevisionFor(<history.txt>, "checkout:1.4.2") // => 7, nil
RevisionFor(<history.txt>, "checkout:1.5.0") // => 6, nil
RevisionFor(<history.txt>, "checkout:1.4")   // => 0, error (no word is exactly that)
```

- Skip lines that don't start with a revision number, such as the two
  header lines and blank lines.
- A match is one of the words after the number equal to `image`.
- If the same image appears twice (after a rollback), return the newer.
- No match: return an error.

**Think it through:**
- `strings.Fields` gives the words. How can `strconv.Atoi` tell a header
  from a revision row?
- Which `slices` function says whether a word is in a slice?

## Run it for real

Optional. Set up kind as in *Pods: your first manifest*, then watch a
rollout and a rollback:

```bash
nix shell nixpkgs#kind
kind create cluster --name gw-rollouts
kubectl create deployment web --image=ghcr.io/stefanprodan/podinfo:6.9.0 --replicas=4
kubectl rollout status deployment/web
kubectl set image deployment/web podinfo=ghcr.io/stefanprodan/podinfo:6.9.1
kubectl annotate deployment/web kubernetes.io/change-cause="image podinfo:6.9.1"
kubectl get rs -w                         # old ReplicaSet down, new one up; Ctrl+C
kubectl set image deployment/web podinfo=ghcr.io/stefanprodan/podinfo:6.9.2
kubectl annotate deployment/web kubernetes.io/change-cause="image podinfo:6.9.2" --overwrite
kubectl rollout status deployment/web
kubectl rollout history deployment/web
kubectl rollout undo deployment/web --to-revision=2
kubectl rollout history deployment/web    # revision 2 is now 4
kind delete cluster --name gw-rollouts
```

## Learn more

- [Kubernetes: Deployments, rolling update](https://kubernetes.io/docs/concepts/workloads/controllers/deployment/#rolling-update-deployment)
- [Kubernetes: Performing a rolling update](https://kubernetes.io/docs/tutorials/kubernetes-basics/update/update-intro/)
- [kubectl rollout reference](https://kubernetes.io/docs/reference/kubectl/generated/kubectl_rollout/)
