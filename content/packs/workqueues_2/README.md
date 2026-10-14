# Workqueues: retries with backoff

A controller that can't reach an API retries, but not forever. Here you fix a
worker that gets retries wrong in three different ways.

**New here:** `AddRateLimited`, `NumRequeues`, `Forget`, per-key exponential backoff.

## Learn

### Recap

Event handlers `Add` `"namespace/name"` keys to a workqueue. A worker calls
`q.Get()`, does the work, and must then call `q.Done(key)`.

Until `Done`, adding the key again only sets a "changed again" mark; the key
never goes back into the queue. `defer` is the usual way to make sure `Done`
runs on every return path.

### Retrying with the rate limiter

The queue from `workqueue.NewTypedRateLimitingQueue` has three more methods:

| Method | What it does |
|---|---|
| `q.AddRateLimited(key)` | Count one more failure for `key`, then add it back after a delay. |
| `q.NumRequeues(key)` | How many failures are counted for `key` so far. |
| `q.Forget(key)` | Reset the failure count for `key` to 0. It does **not** remove the key from the queue. |

The delay comes from the *rate limiter*. The usual one backs off
exponentially per key, as in *Retries: backoff and jitter*: base delay, then
twice that, then four times, up to a maximum.

```go
limiter := workqueue.NewTypedItemExponentialFailureRateLimiter[string](
	5*time.Millisecond, // base delay
	time.Minute,        // maximum delay
)
q := workqueue.NewTypedRateLimitingQueue(limiter)
```

Each key has its own count, so one broken object doesn't slow the others. The
tests use a 1ms base and 10ms maximum, so retries finish in milliseconds.

### The three outcomes

After each attempt, a worker does exactly one of these:

| Outcome | What to do | Why |
|---|---|---|
| Success | reset the failure count | otherwise old failures still count, and the next breakage is dropped too early |
| Failure, retries left | add it back with a delay | the service may recover |
| Failure, no retries left | reset the count and drop the key | retrying forever hammers a broken service; a later change to the object adds the key again anyway |

`Done` is separate from all of these: it ends this turn of the key, whatever
happened.

### Comparing with a limit

A limit like `MaxRetries` is usually a `const`. Compare it with `<` to ask
"is there room left?":

```go
const MaxBackups = 5

if len(backups) < MaxBackups {
	// still room for another backup
}
```

> **Why is it written like this?** Go requires the `if` condition to be a
> `bool`, with no parentheses needed. An exported `MaxRetries` constant
> (capital letter) is the author's choice, so tests and callers can read it.

## Your task

The `monitoring` team syncs alert rules from ConfigMaps to an alerting
service that is sometimes down. Their worker in `exercise.go` has three bugs.

### 1. Let keys come back

**Goal:** fix `ProcessNextItem` so every key it takes from `q.Get()` is marked done, whichever way the function returns.

Right now, after a key's first turn, adding it again never puts it back in
the queue.

```go
q.Add("monitoring/alert-rules")
ProcessNextItem(q, sync) // calls sync("monitoring", "alert-rules"), returns true
q.Add("monitoring/alert-rules")
q.Len() // => 1 (the buggy version gives 0)
```

**Think it through:**
- Which queue method ends a key's turn?
- Where can you put it so it runs even on an early `return`?

### 2. Forget on success

**Goal:** fix `HandleResult` so that when `err` is `nil` it resets the key's failure count and returns `false`.

Don't add the key back to the queue.

```go
// "payments/alert-rules" has failed 3 times before
HandleResult(q, "payments/alert-rules", nil) // => false
q.NumRequeues("payments/alert-rules")        // => 0
q.Len()                                      // => 0
```

**Think it through:**
- Look at the method table: which one resets the count?
- Find the branch in `HandleResult` that runs when `err` is `nil`.

### 3. Give up after MaxRetries

**Goal:** make `HandleResult` stop retrying a failing key once it has failed `MaxRetries` times before.

When `err` is not `nil`:

- If the key's earlier failures are **fewer than** `MaxRetries`: add it back
  with a delay and return `false`.
- Otherwise: reset its failure count and return `true`. `ProcessNextItem` then
  logs a warning.

With `MaxRetries = 3`, a key that always fails runs 4 times: the first try
plus 3 retries.

| Earlier failures (`NumRequeues`) | `HandleResult(q, key, err)` returns | `NumRequeues` afterwards | Key comes back? |
|---|---|---|---|
| 0 | `false` | 1 | yes, after a short delay |
| 2 | `false` | 3 | yes |
| 3 | `true` | 0 | no |
| 5 | `true` | 0 | no |

The last test runs the whole worker against a sync that always fails and
checks it is called exactly 4 times. It needs your fixes from tasks 1 and 3.

**Think it through:**
- Which method tells you the earlier failures, and which constant is the limit?
- Which two methods from the table do the two branches need?

## Learn more

- [pkg.go.dev: workqueue.TypedRateLimitingInterface](https://pkg.go.dev/k8s.io/client-go/util/workqueue#TypedRateLimitingInterface)
- [pkg.go.dev: NewTypedItemExponentialFailureRateLimiter](https://pkg.go.dev/k8s.io/client-go/util/workqueue#NewTypedItemExponentialFailureRateLimiter)
- [sample-controller: controller.go](https://github.com/kubernetes/sample-controller/blob/master/controller.go)
