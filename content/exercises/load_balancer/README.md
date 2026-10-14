# Load balancing: round robin with health checks

A *load balancer* spreads requests across several copies of a service and
stops sending traffic to the copies that are down. You'll build the heart of
one: take turns, skip the broken ones, and fail fast when nothing is left.

**New here:** load balancers, round robin, wrapping an index with `%`,
changing slice elements through their index, doing slow work outside a lock.

## Learn

### What a load balancer does

Your API runs as several copies, called *backends* (or replicas, or
targets), each with its own address such as `10.0.1.5:8080`. Clients only
know one address: the load balancer's. For every request, the balancer picks
a backend and passes the request on.

You'll meet load balancers everywhere in ops work:

| Where | What it's called |
|---|---|
| nginx | an `upstream` block listing the backend servers |
| Kubernetes | a Service, spreading traffic across its Pods |
| AWS | an Application Load Balancer (ALB) and its target group |

### Round robin

The simplest way to pick is *round robin*: take turns. First backend, second,
third, then back to the first. Keep the position of whose turn it is, and
after each pick move it on by one. The remainder operator `%` wraps it back
to 0 after the last position:

```go
speakers := []string{"ana", "ben", "cy"}
turn := 0
for range 5 {
	fmt.Println(speakers[turn])
	turn = (turn + 1) % len(speakers)
}
// ana ben cy ana ben
```

> **Why is it written like this?** `(turn + 1) % 3` is `1, 2, 0, 1, 2, ...`:
> the remainder after dividing by 3 is always 0, 1 or 2, so it's always a
> valid index. Go has no built-in "wrap around", so `%` is the usual way.
> Dividing by zero panics, so `% len(...)` needs at least one element.

### Skip the backends that are down

A backend crashes. If the balancer keeps taking turns, every third request
fails. So the balancer tracks each backend's health and **skips** the ones
marked down. A *health check* marks them down when a probe (such as
`GET /healthz`) fails, and up again when it passes.

Skipping means: start at whose turn it is, and walk forward until you find a
healthy backend. Each backend needs to be looked at **at most once** per
call; after a full lap, you know none of them is healthy.

### Changing slice elements

`range` hands you a **copy** of each element. Changing the copy doesn't
change the slice. To change an element, go through its index:

```go
for _, s := range shelves {
	s.count = 0 // changes a copy: shelves is untouched
}
for i := range shelves {
	shelves[i].count = 0 // changes the element in the slice
}
```

> **Why is it written like this?** Go requires this: the loop variable is a
> new variable holding a copy. `for i := range shelves` (index only) is the
> usual way to modify elements in place.

### When every backend is down: fail fast

If nothing is healthy, waiting won't help, and every waiting request ties up
a goroutine and a client. So `Next` returns an error straight away. The HTTP
handler in front of it turns that error into `503 Service Unavailable`, which
is exactly what an ALB answers when a target group has no healthy targets:

```go
addr, err := pool.Next()
if errors.Is(err, balancer.ErrNoHealthy) {
	http.Error(w, "no healthy backends", http.StatusServiceUnavailable)
	return
}
```

`ErrNoHealthy` is a *sentinel error*, as in *Errors*: one exported value that
callers compare against with `errors.Is`. Its message is
`balancer: no healthy backends`.

### Why the mutex

`net/http` runs every request on its own goroutine. A hundred requests at
once means a hundred goroutines calling `Next` on the **same** `RoundRobin`.
Reading `rr.next`, then writing it back, is two steps: without a lock, two
goroutines can read the same value, send both requests to the same backend,
and the race detector reports a data race.

So every method locks `rr.mu` first and unlocks it when done, as in
*Mutexes*. `Healthy` in your starter shows the pattern: `Lock`, then
`defer Unlock`.

### Slow work outside the lock

A health check talks to the network, and one slow backend can take seconds
to answer. If `CheckAll` held `rr.mu` all that time, every `Next` would wait
too, and your whole site would stall during each health check.

The fix is a three-step pattern: lock and **copy** what you need, unlock and
do the slow work, then lock again and **store** the results:

```go
s.mu.Lock()
urls := slices.Clone(s.urls) // slices.Clone returns a new copy of a slice
s.mu.Unlock()

sizes := make([]int, len(urls))
for i, u := range urls {
	sizes[i] = download(u) // slow, and no lock held
}

s.mu.Lock()
defer s.mu.Unlock()
s.sizes = sizes
```

> **Why is it written like this?** `defer` runs when the function returns,
> so it can't unlock in the middle; the first unlock has to be an explicit
> `Unlock()` call. Copying is required: after `Unlock`, another goroutine may
> change `s.urls` while you read it. Using a slice of results, one slot per
> URL, is the author's choice.

## Your task

Your team's API runs on four VMs behind a small Go load balancer. Finish it
so it takes turns, skips VMs that are down, and fails fast when none are
left.

### 1. Rotate through the backends

**Goal:** make `Next` return the backends in list order, one per call,
starting again at the first after the last.

Lock `rr.mu` for the whole call. Return the address of the backend at
position `rr.next` and a `nil` error, and move `rr.next` on by one, back to
`0` after the last backend. For this task, every backend is healthy and there
is at least one.

```go
rr := New("10.0.1.5:8080", "10.0.1.6:8080", "10.0.1.7:8080")
rr.Next() // => "10.0.1.5:8080", nil
rr.Next() // => "10.0.1.6:8080", nil
rr.Next() // => "10.0.1.7:8080", nil
rr.Next() // => "10.0.1.5:8080", nil
```

The tests also call `Next` 200 times from 8 goroutines at once on four
backends, and expect exactly 50 requests per backend.

**Think it through:**
- Which field holds the turn, and which operator wraps it back to 0?
- How do you make sure the lock is released on every `return`?

### 2. Skip backends that are down

**Goal:** make `MarkDown` and `MarkUp` change a backend's health, and make
`Next` skip backends that are down.

- `MarkDown(addr)` sets `healthy` to `false` on the backend with that
  address; `MarkUp(addr)` sets it to `true`. An address that isn't in the
  list changes nothing. Both lock `rr.mu`.
- `Next` starts at `rr.next` and walks forward, wrapping around, looking at
  **at most** `len(rr.backends)` backends. It returns the first healthy one,
  and leaves `rr.next` at the position just after it.

```go
rr := New("10.0.1.5:8080", "10.0.1.6:8080", "10.0.1.7:8080")
rr.MarkDown("10.0.1.6:8080")
rr.Next() // => "10.0.1.5:8080", nil
rr.Next() // => "10.0.1.7:8080", nil  (10.0.1.6 skipped)
rr.Next() // => "10.0.1.5:8080", nil
rr.MarkUp("10.0.1.6:8080")
rr.Next() // => "10.0.1.6:8080", nil
```

**Think it through:**
- `range` gives you copies. How do you change the `healthy` field of the
  element in the slice itself?
- In `Next`, which loop runs a fixed number of times? What happens inside
  each lap: look at a backend, move the turn on, and then what?

### 3. Fail fast when every backend is down

**Goal:** make `Next` return `""` and `ErrNoHealthy` when no backend is
healthy.

That includes a balancer with no backends at all (`New()`): `Next` must
return the error, not panic. After a failed call, the rotation carries on
where it was once backends come back:

```go
rr := New("10.0.1.5:8080", "10.0.1.6:8080", "10.0.1.7:8080")
rr.Next() // => "10.0.1.5:8080", nil
// all three marked down
rr.Next() // => "", ErrNoHealthy
// all three marked up
rr.Next() // => "10.0.1.6:8080", nil
New().Next() // => "", ErrNoHealthy
```

**Think it through:**
- When your loop from task 2 finishes without returning, what does that
  tell you?
- With zero backends, does your code ever index the slice or use `%`?

### 4. Probe every backend

**Goal:** make `CheckAll(probe)` call `probe` once per backend, set each
backend's health to its answer, and return how many backends are healthy.

- `probe(addr)` returns `true` for a healthy backend and `false` for a down
  one. Call it for one backend at a time, in list order.
- Don't hold `rr.mu` while `probe` runs. The tests call `Next` while a probe
  is still running, and expect an answer straight away.
- Use the three steps from the Learn section. Add `"slices"` to the imports
  if you use `slices.Clone`.

```go
rr := New("10.0.1.5:8080", "10.0.1.6:8080", "10.0.1.7:8080", "10.0.1.8:8080")
rr.CheckAll(probe) // probe fails only for 10.0.1.6 => 3
rr.Healthy()       // => ["10.0.1.5:8080" "10.0.1.7:8080" "10.0.1.8:8080"]
```

**Think it through:**
- What do you copy under the lock, and where do the probe results go
  while you're not holding it?
- After locking again, which index ties a result back to its backend?

## Learn more

- [nginx: ngx_http_upstream_module](https://nginx.org/en/docs/http/ngx_http_upstream_module.html)
- [Kubernetes: Service](https://kubernetes.io/docs/concepts/services-networking/service/)
- [AWS: Health checks for Application Load Balancer target groups](https://docs.aws.amazon.com/elasticloadbalancing/latest/application/target-group-health-checks.html)
- [Wikipedia: Round-robin scheduling](https://en.wikipedia.org/wiki/Round-robin_scheduling)
