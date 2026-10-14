# Load balancing: fix the least-connections balancer

Your team swapped round robin for a *least-connections* balancer, because
the slow report endpoint kept piling up on one backend. The new balancer has
three bugs: it picks at random when backends tie, its in-flight counts go
negative, and `go test -race` reports a data race. Fix all three.

**New here:** least-connections balancing. (Random map order is from *Maps*,
`sync.Once` and race reports from *The race detector* exercises.)

## Learn

### Recap

As in *Load balancing: round robin with health checks*, a load balancer
picks one backend per request. Round robin takes turns: first, second,
third, first again. A mutex guards the balancer's state because `net/http`
calls it from many goroutines at once.

### Why least connections

Round robin hands out equal **numbers** of requests, not equal **work**.
Suppose `/reports/monthly` takes 30 seconds and `/healthz` takes 1
millisecond, and they arrive in turn on two backends:

| Request | Round robin sends it to | Least connections sends it to |
|---|---|---|
| report | backend 1 | backend 1 (both idle, the earliest wins) |
| health check | backend 2 | backend 2 (backend 1 is busy) |
| report | backend 1, now running 2 reports | backend 2 (its health check already finished) |
| health check | backend 2 | backend 1 (both have 1, the earliest wins) |
| report | backend 1, now running 3 reports | backend 1 (each has 1 report, the earliest wins) |

With round robin, backend 1 ends up with every slow report while backend 2
sits idle. *Least connections* counts the requests **in flight** on each
backend, and sends each new one to the backend with the fewest. nginx calls
it `least_conn`; an AWS ALB calls it "least outstanding requests".

### Acquire and release

The balancer only knows a request has finished if you tell it. `Acquire`
picks a backend and adds 1 to its count. It also returns a `release`
function, which takes the 1 off again; the caller runs it when the request
is done:

```go
addr, release, err := lc.Acquire()
if err != nil {
	return err
}
defer release()
// forward the request to addr
```

`release` is a *closure*, as in *Functions as values*: it remembers which
backend it belongs to. It runs later, often on a different goroutine.

### Map order is random

`range` over a map visits the keys in a **different order each time**. Go
does this on purpose, so nobody relies on an order that was never promised:

```go
sizes := map[string]int{"small": 1, "medium": 2, "large": 3}
for name := range sizes {
	fmt.Println(name) // medium, large, small... or any other order
}
```

When order matters, loop over a slice that holds the keys in the order you
want, and look each one up in the map.

### Releasing exactly once

Code paths that end a request can overlap: a `defer release()` plus an
early `release()` on an error path, or a retry wrapper that releases again.
Each extra call takes 1 more off the count. The count goes negative, and the
backend looks *less* busy than idle, so it gets flooded.

`sync.Once`, as in *The race detector: hidden races*, runs a function
exactly once, however many times you call `Do`:

```go
var once sync.Once
closeLog := func() {
	once.Do(func() { fmt.Println("log closed") })
}
closeLog() // prints "log closed"
closeLog() // prints nothing
```

> **Why is it written like this?** Each closure must get its **own**
> `sync.Once`, so declare it right before the closure, inside the function
> that creates it. A `sync.Once` in the struct would be shared by every
> request. That placement is required for correctness; the names are your
> choice.

### Reading a race report

Run the tests with `go test -race`. When two goroutines touch the same
memory without a lock between them, the race detector prints a report like
this one, from the starter code (shortened; your line numbers may differ):

```text
WARNING: DATA RACE
Write at 0x00c000144100 by goroutine 70:
  groundwork.local/….(*LeastConn).Acquire.func1()
      …/exercise.go:53 +0xa8

Previous write at 0x00c000144100 by goroutine 49:
  groundwork.local/….(*LeastConn).Acquire()
      …/exercise.go:47 +0x433
```

Read it like this:

- There are two stacks: the access that tripped the detector, and the
  **previous** access it clashes with. At least one of them is a write.
- The top line of each stack in **your** code names the function and the
  file and line. `Acquire.func1` is the first anonymous function written
  inside `Acquire`: here, the `release` closure. A `.1` on the end
  (`Acquire.func1.1`) means a function written inside *that* one.
- Then ask: which of the two lines doesn't hold the lock?

Holding a lock while you create a closure doesn't help: the closure runs
**later**, long after `Acquire` has unlocked. Code inside it has to lock for
itself.

Sometimes the run stops instead with `fatal error: concurrent map writes`
(or `concurrent map iteration and map write`). Go's maps check for this
bug themselves. It's the same bug, with the same fix.

## Your task

The balancer sits in front of the reporting service's three backends. Fix
`Acquire` and its `release` closure in `exercise.go`.

### 1. Break ties in list order

**Goal:** make `Acquire` pick the backend with the fewest requests in
flight, and on a tie, the one that comes **first** in the list given to
`NewLeastConn`.

Right now it ranges over the `lc.active` map, so when backends tie, the
pick changes from run to run. That breaks the promise in `Acquire`'s doc
comment, and behaviour that changes every run can't be reproduced in a test
or replayed after an incident. `lc.addrs` holds the addresses in list order.
With no backends, `Acquire` already returns `""`, `nil` and `ErrNoBackends`;
keep that.

```go
lc := NewLeastConn("10.0.2.1:8080", "10.0.2.2:8080", "10.0.2.3:8080")
lc.Acquire() // => "10.0.2.1:8080" (the buggy version: any of the three)
lc.Acquire() // => "10.0.2.2:8080"
lc.Acquire() // => "10.0.2.3:8080"
lc.Acquire() // => "10.0.2.1:8080" (all have 1: the tie goes to the first)
```

**Think it through:**
- Which field gives you the backends in list order?
- If you start with the first backend as the best so far, should a later
  backend replace it when it has the *same* count, or only *fewer*?

### 2. Release only once

**Goal:** make each `release` take its request off the count once, however
many times it is called.

```go
_, release, _ := lc.Acquire() // picks "10.0.2.1:8080"
release()
release()
lc.Active("10.0.2.1:8080") // => 0 (the buggy version says -1)
```

Each request's `release` is separate: releasing one request never undoes
another one's.

**Think it through:**
- Where does the `sync.Once` have to live so that every request gets its own?
- What goes inside `once.Do`?

### 3. Fix the data race in release

**Goal:** make `release` safe to call from any goroutine, so `go test -race`
reports nothing.

The tests acquire and release from many goroutines at once, and release
some requests from a different goroutine than the one that acquired them.
Afterwards, every backend must have 0 requests in flight.

Once task 2 is done, the racing line sits inside `once.Do`, so the report
names it `Acquire.func1.1`, with `sync.Once`'s own code under it:

```text
WARNING: DATA RACE
Write at 0x00c00039c0e8 by goroutine 51:
  groundwork.local/….(*LeastConn).Acquire.func1.1()
      …/exercise.go:54 +0xa8
  sync.(*Once).doSlow()
      …/sync/once.go:78 +0xa1
```

**Think it through:**
- Which line in the report is your code, and which lock does it skip?
- `Acquire` locks `lc.mu` and defers the unlock. Is that lock still held
  when `release` runs?

## Learn more

- [nginx: Choosing a load-balancing method](https://docs.nginx.com/nginx/admin-guide/load-balancer/http-load-balancer/#method)
- [AWS: Routing algorithms for Application Load Balancer target groups](https://docs.aws.amazon.com/elasticloadbalancing/latest/application/edit-target-group-attributes.html#modify-routing-algorithm)
- [Data Race Detector](https://go.dev/doc/articles/race_detector)
- [sync.Once](https://pkg.go.dev/sync#Once)
