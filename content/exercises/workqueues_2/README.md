# Workqueues: retries with backoff

A controller that can't reach an API retries, but not forever. Here you fix a
worker that gets retries wrong in three different ways.

## Learn

### Recap

Event handlers `Add` `"namespace/name"` keys to a workqueue. A worker loop
calls `q.Get()`, does the work, and must call `q.Done(key)` afterwards. Until
`Done`, a key that is added again only gets a "changed again" mark and never
goes back into the queue.

### Retrying with the rate limiter

The queue from `workqueue.NewTypedRateLimitingQueue` has three more methods:

| Method | What it does |
|---|---|
| `q.AddRateLimited(key)` | Count one more failure for `key`, then add it back after a delay. |
| `q.NumRequeues(key)` | How many failures are counted for `key` so far. |
| `q.Forget(key)` | Reset the failure count for `key` to 0. It does **not** remove the key from the queue. |

The delay comes from the *rate limiter* you pass to the constructor. The usual
one backs off exponentially per key, as in *Retries: backoff and jitter*: 1st retry
after the base delay, then twice that, then four times, up to a maximum. Each
key has its own count, so one broken object doesn't slow down the others.

```go
limiter := workqueue.NewTypedItemExponentialFailureRateLimiter[string](
	5*time.Millisecond, // base delay
	time.Minute,        // maximum delay
)
q := workqueue.NewTypedRateLimitingQueue(limiter)
```

The tests pass in a limiter with a 1ms base and a 10ms maximum, so retries
finish in milliseconds.

The rules a worker follows after each attempt:

- **Success:** call `Forget(key)`. Otherwise the old failures still count, and
  the next time the object breaks it is dropped too early.
- **Failure, retries left:** `AddRateLimited(key)`.
- **Failure, no retries left:** `Forget(key)` and drop the key. Retrying
  forever hammers a broken service and fills the queue with keys that will
  never work. A later change to the object adds the key again anyway.

`Done` is separate from all of these: it ends this turn of the key, whatever
happened.

## Your task

The `monitoring` team syncs alert rules from ConfigMaps to an alerting
service that is sometimes down. Their worker in `exercise.go` has three bugs.

### 1. Let keys come back

After a key's first turn, adding it again never puts it back in the queue.
Fix `ProcessNextItem` so every key it takes from `q.Get()` is marked done,
whichever way the function returns.

```go
q.Add("monitoring/alert-rules")
ProcessNextItem(q, sync) // calls sync("monitoring", "alert-rules"), returns true
q.Add("monitoring/alert-rules")
q.Len() // => 1 (the buggy version gives 0)
```

### 2. Forget on success

Fix `HandleResult` so that when `err` is `nil` it resets the key's failure
count and returns `false`:

```go
// "payments/alert-rules" has failed 3 times before
HandleResult(q, "payments/alert-rules", nil) // => false
q.NumRequeues("payments/alert-rules")        // => 0
```

### 3. Give up after MaxRetries

`HandleResult` retries every failure forever. Change it so that when `err` is
not `nil`:

- if `q.NumRequeues(key) < MaxRetries`, call `q.AddRateLimited(key)` and
  return `false`;
- otherwise call `q.Forget(key)` and return `true`. `ProcessNextItem` then
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

## Learn more

- [pkg.go.dev: workqueue.TypedRateLimitingInterface](https://pkg.go.dev/k8s.io/client-go/util/workqueue#TypedRateLimitingInterface)
- [pkg.go.dev: NewTypedItemExponentialFailureRateLimiter](https://pkg.go.dev/k8s.io/client-go/util/workqueue#NewTypedItemExponentialFailureRateLimiter)
- [sample-controller: controller.go](https://github.com/kubernetes/sample-controller/blob/master/controller.go)
