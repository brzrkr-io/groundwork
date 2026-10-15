# Concurrency

**In one sentence:** concurrency is structuring a program as several tasks
that make progress independently, so waiting on one doesn't stall the rest.

**Why it matters:** a health checker that probes 500 hosts one by one takes
minutes; probing them concurrently takes about as long as the slowest one.
Almost every ops tool you'll write (checkers, log shippers, controllers,
cloud scripts) spends its life waiting on the network, and concurrency is how
it waits for many things at once.

**You'll learn:**

- What concurrency is, and how it differs from parallelism.
- Why programs that *wait* gain the most from it.
- How goroutines relate to processes, threads and CPU cores.
- Why shared data is the core danger, and how a race loses updates.
- The tools: mutexes, channels, WaitGroups, `select` and context.
- The hazards (deadlock, leaks, too much at once) and the common patterns.

You can read this guide before you reach the *Concurrency* module. The first
sections use plain words only. The Go code uses what that module teaches:
`go`, channels, `sync.WaitGroup`, `sync.Mutex`, `select` and `context`.

## One cook or several

**Picture a kitchen with one cook.** They put pasta on to boil, then stand
and watch the pot for ten minutes. Only then do they start chopping salad.
Dinner is slow, and the cook spent most of the time doing nothing.

A better cook starts the pasta, chops salad *while* it boils, and checks the
oven in between. Still one person, but no task blocks the others. Add a
second cook and two things can truly happen at the same moment.

Now the ops version. You must check 500 hosts, and each check waits about one
second for the network to answer.

```text
One at a time:

host 1  [=wait=]
host 2          [=wait=]
host 3                  [=wait=]
  ...                           ...  500 checks x 1 s = about 8 minutes

All at once:

host 1  [=wait=]
host 2  [=wait=]
host 3  [=wait=]
  ...
host 500[=wait=]                     about 1 s (as slow as the slowest)
```

The work didn't get faster. Each check still takes a second. What changed is
that the waits **overlap** instead of queueing up behind each other.

## Concurrency is not parallelism

These two words get mixed up, so pin them down. Rob Pike, one of Go's
creators, put it this way: *concurrency is about dealing with lots of things
at once; parallelism is about doing lots of things at once.*

| | Concurrency | Parallelism |
|---|---|---|
| What it is | a way to **structure** a program | a way to **run** it |
| Kitchen | one cook juggling three dishes | two cooks working side by side |
| Needs many cores? | no | yes |

A *CPU core* is the part of the processor that runs instructions. A laptop
has maybe 8; a big server has 64 or more. One core runs one thing at any
instant, but it switches between tasks so fast that, to you, they all seem to
be moving.

```text
One core, concurrent (tasks take turns):
core 1: A A B B A C C B A C ...

Two cores, concurrent AND parallel:
core 1: A A A C C A ...
core 2: B B C B B C ...
```

A concurrent program *may* run in parallel if there are spare cores. Go
decides that for you. Your job is the structure: split the work into
independent tasks and say how they talk to each other.

## Waiting vs computing

Not every program gains from concurrency. Ask: what is it mostly doing?

- **I/O-bound:** mostly *waiting* for input/output: the network, a disk, a
  database, another program. A health check is 1 ms of work and 999 ms of
  waiting for a reply.
- **CPU-bound:** mostly *computing*: compressing a file, hashing, resizing
  images. The core is busy the whole time.

| Task | Bound by | What concurrency gives you |
|---|---|---|
| Probe 500 HTTP endpoints | waiting (network) | huge win: overlap the waits |
| Upload 1,000 logs to S3 | waiting (network) | huge win |
| Gzip 100 rotated log files | computing | one file per core at most |

Most DevOps code is I/O-bound, so concurrency pays off hugely: one core can
keep thousands of network requests in flight. For CPU-bound work, 8 cores
give at most about 8 times the speed, and 10,000 tasks just add overhead.

## Processes, threads and goroutines

Three layers run your code. You met the first two in *The operating system
for programmers* guide; here's how goroutines sit on top.

- A *process* is a running program with its own private memory. Your Go
  binary, `nginx` and `kubelet` are separate processes. They can't see each
  other's variables.
- A *thread* is a path of execution inside a process, scheduled onto cores by
  the operating system. Threads in one process **share** its memory. Each
  thread gets a fixed stack, often megabytes, and switching threads goes
  through the kernel.
- A *goroutine* is Go's lightweight task. It starts with a stack of a few
  kilobytes that grows as needed, so a program can run hundreds of thousands
  of them. The Go runtime (code built into every Go binary) runs your
  goroutines on a small number of OS threads.

```text
  your goroutines     g  g  g  g  g  g  g  g  g  g  g  g  (thousands)
                       \  |  /     \  |  /     \  |  /
  Go runtime          [ scheduler: picks a ready goroutine ]
                            |                 |
  OS threads             thread 1          thread 2   (a handful)
                            |                 |
  CPU cores              core 1            core 2     (GOMAXPROCS = 2)
```

When a goroutine waits (for the network, a channel, a timer), the scheduler
parks it and runs another one on that thread. That is how one thread keeps
many checks in flight.

*GOMAXPROCS* is how many threads may run Go code at the same instant. It
defaults to the number of CPUs your program may use. Since Go 1.25 that
includes a container's CPU limit, so a pod limited to 2 CPUs gets 2, not the
node's 64.

You start a goroutine by putting `go` in front of a function call. The caller
does not wait; both carry on.

## Your first goroutines

Here is the 500-host story with 5 hosts and a fake check that only sleeps,
the way a real check waits on the network.

```go
package main

import (
	"fmt"
	"sync"
	"time"
)

// check pretends to probe a host. It only waits, like a network call.
func check(host string) string {
	time.Sleep(200 * time.Millisecond)
	return host + " ok"
}

func main() {
	hosts := []string{"web-1", "web-2", "web-3", "db-1", "cache-1"}

	start := time.Now()
	for _, h := range hosts {
		fmt.Println(check(h))
	}
	took := time.Since(start).Round(100 * time.Millisecond)
	fmt.Println("one at a time:", took)

	start = time.Now()
	var wg sync.WaitGroup
	for _, h := range hosts {
		wg.Go(func() {
			fmt.Println(check(h))
		})
	}
	wg.Wait()
	took = time.Since(start).Round(100 * time.Millisecond)
	fmt.Println("all at once:", took)
}
```

One run printed:

```text
web-1 ok
web-2 ok
web-3 ok
db-1 ok
cache-1 ok
one at a time: 1s
cache-1 ok
web-2 ok
web-3 ok
web-1 ok
db-1 ok
all at once: 200ms
```

Step by step:

1. The first loop calls `check` five times in a row: 5 × 200 ms = 1 s.
2. The second loop hands each call to `wg.Go`, which starts it in a new
   goroutine and counts it. The loop finishes almost instantly.
3. `wg.Wait()` blocks `main` until all five goroutines have returned.
4. All five sleeps overlap, so the whole batch takes about 200 ms.

**Look at the order of the second batch.** It changes from run to run. Once
tasks run concurrently, *you* no longer decide who finishes first; the
scheduler and the network do. More on this in the *Patterns* section.

## The core problem: shared state

Goroutines are cheap to start. The hard part is when two of them touch the
**same** data. (The *State and tracing code by hand* guide covers what state
is; here it gets a second actor.)

### A lost update, step by step

Say a server counts requests with `requests++`. That looks like one step,
but the CPU does three:

1. **read** the current value from memory,
2. **add** 1 to it,
3. **write** the new value back.

Two goroutines handle requests at the same moment. Here is one possible
interleaving:

| Step | Goroutine A | Goroutine B | `requests` in memory |
|---|---|---|---|
| 1 | reads 41 | | 41 |
| 2 | | reads 41 | 41 |
| 3 | adds 1 → 42 | | 41 |
| 4 | | adds 1 → 42 | 41 |
| 5 | writes 42 | | 42 |
| 6 | | writes 42 | 42 |

Two requests arrived, but the counter went up by one. B's update was
**lost**, with no error and no crash. This *read-modify-write* pattern (read
a value, change it, write it back) is behind most concurrency bugs.

It's also why these bugs are nasty in production: the bad interleaving is
rare. Tests pass a thousand times, and then your dashboard undercounts during
a traffic spike.

### Data race vs race condition

- A *data race* is two goroutines touching the same memory at the same time,
  at least one of them writing, with nothing coordinating them. In Go, that
  is always a bug: the result is unpredictable.
- A *race condition* is the wider idea: the program's result depends on the
  timing of events. You can have one without any data race. Two deploy jobs
  each check "is the slot free?", both see yes, and both deploy. Each step
  was safe; the *sequence* wasn't.

The tools below fix data races. Race conditions need you to think about
which steps must happen together, as one unit.

### Catching it: the race detector

Here is the broken counter as a full program. It's wrong on purpose:

```go
package main

import (
	"fmt"
	"sync"
)

func main() {
	requests := 0
	var wg sync.WaitGroup
	for range 1000 {
		wg.Go(func() {
			requests++ // read, add 1, write back: NOT safe
		})
	}
	wg.Wait()
	fmt.Println("requests:", requests)
}
```

With plain `go run .` it may print 1000, or 987, or 739. Run it with Go's
*race detector* instead, `go run -race .`, and you get a report like this
(trimmed; your paths and numbers will differ):

```text
==================
WARNING: DATA RACE
Read at 0x00c0000180f8 by goroutine 12:
  main.main.func1()
      /tmp/race/main.go:13 +0x2e

Previous write at 0x00c0000180f8 by goroutine 9:
  main.main.func1()
      /tmp/race/main.go:13 +0x44
...
==================
requests: 739
Found 2 data race(s)
exit status 66
```

Read it like a stack trace: **what** happened (a read racing a write), and
**where** (both at `main.go:13`, the `requests++` line).

The detector only sees races that actually happen while it watches. So run
your tests with `go test -race`, ideally in CI, with tests that really
exercise the concurrent paths. You'll practise reading these reports in
*The race detector*.

## Tool: the mutex

A *mutex* ("mutual exclusion") is a lock. Only one goroutine can hold it at
a time; the others wait at `Lock()` until it's released.

```go
package main

import (
	"fmt"
	"sync"
)

func main() {
	var (
		mu       sync.Mutex
		requests int // guarded by mu
	)
	var wg sync.WaitGroup
	for range 1000 {
		wg.Go(func() {
			mu.Lock()
			requests++
			mu.Unlock()
		})
	}
	wg.Wait()
	fmt.Println("requests:", requests)
}
```

This prints `requests: 1000` every time, and `go run -race .` is silent.

The code between `Lock` and `Unlock` is the *critical section*: the steps
that must happen as one unit. Rules of thumb:

- **Keep it small.** Lock, touch the data, unlock. Never hold a lock while
  you make an HTTP call or read a file: every other goroutine queues behind
  your slow network.
- **Every access goes through the lock**, reads included. One unlocked read
  is still a data race.
- **Write down what it guards.** A comment like `// guarded by mu` saves the
  next reader.

For a single number there is a lighter tool. `sync/atomic` gives types such
as `atomic.Int64` whose `Add` does the read-modify-write as one indivisible
step. Use it for counters and flags; use a mutex when several values must
change together.

## Tool: channels

A *channel* is a pipe between goroutines: one sends a value in, another
receives it out. Go's motto, from the Effective Go document:

> Do not communicate by sharing memory; instead, share memory by
> communicating.

Instead of many goroutines poking one shared variable, each value has a
single owner at a time. Sending it on a channel hands it over.

```text
  checker goroutines            main goroutine
  ------------------            --------------
  web-1 --"web-1 ok"--\
  web-2 --"web-2 ok"---+--> [ results channel ] --> prints, counts
  db-1  --"db-1 down"-/
```

Only `main` touches the totals, so they need no lock. Key facts:

| Action | Code | Behaviour |
|---|---|---|
| make | `make(chan string)` | *unbuffered*: a send waits until a receiver takes it |
| make buffered | `make(chan string, 10)` | *buffered*: a send waits only when 10 are already queued |
| send | `ch <- v` | waits until there's a receiver (unbuffered) or room (buffered) |
| receive | `v := <-ch` | waits until a value arrives |
| close | `close(ch)` | "no more values"; `for v := range ch` ends |

That waiting is the point. A channel both moves the data **and**
synchronises the two goroutines. Only the sender should close a channel, and
only when nothing else will be sent on it: a send on a closed channel panics.

## Tool: WaitGroup

A *WaitGroup* answers "has everyone finished?" You've already seen it:
`wg.Go(f)` starts `f` and counts it, and `wg.Wait()` blocks until the count
is back to zero. Without it, `main` returns early, and when `main` returns,
the program exits, killing every goroutine mid-check.

## Tools: select and context

**`select`** waits on several channel operations and runs whichever is ready
first. The classic use is a timeout: "a result, or give up after 2 seconds."

A *context* (`context.Context`) carries a cancellation signal and an
optional deadline through your call chain. When it's cancelled (say, the
program is shutting down) or a request's 5-second budget runs out,
everything working on it sees `ctx.Done()` close and can stop.

```go
package main

import (
	"context"
	"fmt"
	"time"
)

// check pretends to probe a host that takes delay to answer. It gives
// up early if ctx is cancelled or its deadline passes.
func check(ctx context.Context, delay time.Duration) error {
	select {
	case <-time.After(delay):
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func main() {
	timeout := 300 * time.Millisecond
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	fmt.Println("web-1:", check(ctx, 100*time.Millisecond))
	fmt.Println("db-1: ", check(ctx, 5*time.Second))
}
```

It prints:

```text
web-1: <nil>
db-1:  context deadline exceeded
```

Step by step:

1. `WithTimeout` makes a context that expires 300 ms from now. `defer
   cancel()` releases it when `main` ends.
2. web-1 answers after 100 ms, before the deadline: `time.After` wins and
   `check` returns `nil`.
3. db-1 would take 5 s. At the 300 ms mark (200 ms into this check),
   `ctx.Done()` closes, that case wins, and `check` returns the context's
   error. The program doesn't sit there for 5 seconds.

Real code passes `ctx` down to `http.NewRequestWithContext` and friends, so
cancelling one context stops a whole tree of work. You'll build this in
*Context and cancellation*.

## Hazards

### Deadlock

A *deadlock* is when goroutines wait for each other forever. Smallest
version (broken on purpose):

```go
package main

import "fmt"

func main() {
	results := make(chan string)
	results <- "web-1 ok" // waits for a receiver... forever
	fmt.Println(<-results)
}
```

`go run .` prints (your path will differ):

```text
fatal error: all goroutines are asleep - deadlock!

goroutine 1 [chan send]:
main.main()
	/tmp/deadlock/main.go:7 +0x36
exit status 2
```

The unbuffered send waits for a receiver, but the only receiver is the next
line of the same goroutine. Go noticed because **every** goroutine was
stuck. The classic two-party version: A holds lock 1 and waits for lock 2,
while B holds lock 2 and waits for lock 1.

That check is best-effort. In a real server other goroutines are still
running (the HTTP listener, say), so a deadlock doesn't crash; requests just
hang. Even this tiny program just hangs under `go run -race .` instead of
printing the error.

### Goroutine leaks

A *goroutine leak* is a goroutine blocked forever on a channel nobody will
ever use again. It never exits, and its memory is never freed. A checker
that leaks one goroutine per timed-out request looks fine for a day, then
runs out of memory and is killed (Kubernetes shows `OOMKilled`).

The fix is to give every goroutine a guaranteed way out: the channel it
reads gets closed, or it also selects on `ctx.Done()`.

### Starvation

*Starvation* is when a goroutine is ready to run but rarely gets its turn,
for example because others keep grabbing the same lock. Holding locks briefly
and not hogging shared resources prevents most of it.

### Too much concurrency

Goroutines are cheap; what they talk to isn't. Start 500 checks at once and
you may hit the open-file limit, flood a small service, or get throttled by a
cloud API. So limit how many run at once. A buffered channel makes a neat
*semaphore*: a box of N tokens, where you take one to start and give it back
when done.

```go
package main

import (
	"fmt"
	"sync"
	"time"
)

func check(host string) {
	time.Sleep(50 * time.Millisecond) // pretend network wait
}

func main() {
	var hosts []string
	for i := 1; i <= 500; i++ {
		hosts = append(hosts, fmt.Sprintf("host-%03d", i))
	}

	slots := make(chan struct{}, 20) // 20 free slots

	var (
		mu      sync.Mutex
		running int // checks in progress right now
		peak    int // the most we ever saw at once
	)

	start := time.Now()
	var wg sync.WaitGroup
	for _, h := range hosts {
		wg.Go(func() {
			// Take a slot. This waits while all 20 are taken.
			slots <- struct{}{}
			// Give the slot back when this check is done.
			defer func() { <-slots }()

			mu.Lock()
			running++
			peak = max(peak, running)
			mu.Unlock()

			check(h)

			mu.Lock()
			running--
			mu.Unlock()
		})
	}
	wg.Wait()
	fmt.Println("hosts checked:", len(hosts))
	fmt.Println("most at once: ", peak)
	took := time.Since(start).Round(100 * time.Millisecond)
	fmt.Println("took about:   ", took)
}
```

It prints (the time may vary slightly):

```text
hosts checked: 500
most at once:  20
took about:    1.3s
```

500 goroutines start, but sending into `slots` blocks once it holds 20. So
at most 20 checks run at once: 25 rounds of 50 ms ≈ 1.3 s. The `running`
and `peak` counters are shared, so a mutex guards them.

## Patterns

A few shapes cover most real programs.

### Worker pool

A *worker pool* is a fixed number of goroutines pulling jobs from one
channel. It's the other way to cap concurrency, and it doesn't start a
goroutine per job.

```text
                     +--> worker 1 --+
 [ jobs channel ] ---+--> worker 2 --+--> [ results channel ] --> main
                     +--> worker 3 --+
```

```go
package main

import (
	"fmt"
	"sync"
	"time"
)

func check(host string) string {
	time.Sleep(100 * time.Millisecond) // pretend network wait
	return host + " ok"
}

func main() {
	var hosts []string
	for i := 1; i <= 12; i++ {
		hosts = append(hosts, fmt.Sprintf("host-%02d", i))
	}

	jobs := make(chan string)
	results := make(chan string)
	start := time.Now()

	// 1. Start a fixed number of workers.
	var wg sync.WaitGroup
	for w := 1; w <= 3; w++ {
		wg.Go(func() {
			for h := range jobs {
				results <- fmt.Sprintf("worker %d: %s", w, check(h))
			}
		})
	}

	// 2. Feed the jobs, then close the channel: "no more work".
	go func() {
		for _, h := range hosts {
			jobs <- h
		}
		close(jobs)
	}()

	// 3. When every worker has returned, close results.
	go func() {
		wg.Wait()
		close(results)
	}()

	// 4. main collects until results is closed.
	count := 0
	for r := range results {
		fmt.Println(r)
		count++
	}
	took := time.Since(start).Round(100 * time.Millisecond)
	fmt.Printf("checked %d hosts in %v\n", count, took)
}
```

One run printed (worker numbers and line order change every run):

```text
worker 3: host-03 ok
worker 1: host-02 ok
worker 2: host-01 ok
worker 1: host-05 ok
worker 2: host-06 ok
worker 3: host-04 ok
worker 1: host-07 ok
worker 2: host-08 ok
worker 3: host-09 ok
worker 1: host-10 ok
worker 2: host-11 ok
worker 3: host-12 ok
checked 12 hosts in 400ms
```

Notice how every goroutine knows how it stops. Workers stop when `jobs` is
closed. The closer goroutine stops after `wg.Wait()`. `main` stops when
`results` is closed. 12 jobs ÷ 3 workers = 4 rounds × 100 ms = 400 ms. You'll
build this in *Worker pools*.

### Fan-out, fan-in

Two names you'll see everywhere: *fan-out* means several goroutines read
from one channel to share the work, and *fan-in* means merging several
channels into one. The worker pool above does both.

```text
fan-out:                         fan-in:
           +--> worker A           logs from web-1 --+
 jobs -----+--> worker B           logs from web-2 --+--> one stream
           +--> worker C           logs from web-3 --+
```

### Pipeline

A *pipeline* is a chain of stages joined by channels. Each stage receives,
does one job, and sends on, like `grep ERROR app.log | sed ...` in a shell.

```go
package main

import (
	"fmt"
	"strings"
)

// read sends each log line, then closes its output.
func read(lines []string) <-chan string {
	out := make(chan string)
	go func() {
		for _, l := range lines {
			out <- l
		}
		close(out)
	}()
	return out
}

// onlyErrors passes on the lines that contain "ERROR".
func onlyErrors(in <-chan string) <-chan string {
	out := make(chan string)
	go func() {
		for l := range in {
			if strings.Contains(l, "ERROR") {
				out <- l
			}
		}
		close(out)
	}()
	return out
}

// tag adds the host name in front of each line.
func tag(host string, in <-chan string) <-chan string {
	out := make(chan string)
	go func() {
		for l := range in {
			out <- host + " | " + l
		}
		close(out)
	}()
	return out
}

func main() {
	logs := []string{
		"INFO  started",
		"ERROR disk full",
		"INFO  request ok",
		"ERROR timeout talking to db",
	}
	for l := range tag("web-1", onlyErrors(read(logs))) {
		fmt.Println(l)
	}
}
```

It prints:

```text
web-1 | ERROR disk full
web-1 | ERROR timeout talking to db
```

```text
read --> onlyErrors --> tag --> main prints
```

All three stages run at once: while `tag` handles line 2, `read` can
already be sending line 4. Each stage closes its output when its input runs
dry, and that "done" signal flows down the chain. With one goroutine per
stage, order is kept. A log shipper (read, parse, filter, batch, send) is
exactly this shape. You'll build one in *Pipelines*.

### Results arrive in any order

When several goroutines feed one channel, results arrive in **finishing**
order, not starting order. If order matters, you have two simple options:

- Send the index along with the result, and sort or place it afterwards.
- Make a slice of the right length first, and let goroutine `i` write only
  `results[i]`. Different elements are different memory, so there's no race,
  as long as nobody reads the slice before `wg.Wait()` returns.

## Where you'll meet this in ops

| Ops job | Concurrency shape |
|---|---|
| Health-check a fleet | fan-out with a limit, collect results |
| Ship logs to a backend | pipeline: read, parse, batch, send |
| Kubernetes controller | a workqueue feeding a few workers |
| Script over AWS accounts | worker pool sized under the API rate limit |
| Stop a job on Ctrl-C or `SIGTERM` | a goroutine turns the signal into a cancelled context |
| Run `kubectl` with a deadline | `exec.CommandContext` kills the child when the context ends |
| HTTP server | `net/http` runs your handlers in goroutines, many at once |

Two notes:

- **Kubernetes controllers** run a small, fixed number of workers that pull
  object keys from a *workqueue*. The queue makes sure two workers never
  handle the same object at once. See the
  *Declarative systems and control loops* guide.
- **Cloud APIs throttle you.** AWS returns throttling errors when you call
  too fast. Twenty workers often finish sooner than five hundred goroutines
  that spend their time being throttled and retrying.
  *Rate limiting: token buckets* in the *Production services* module goes
  further.

Every `net/http` server already runs handlers concurrently, each request's
handler in a goroutine. So any map or counter shared between your handlers
needs a mutex, even if you never typed `go` yourself.

## A checklist before you write concurrent code

Ask these four questions every time:

1. **Who owns this data?** Each piece of shared state is either guarded by
   one mutex, or owned by one goroutine and passed around on channels.
2. **How does each goroutine stop?** A closed channel, a cancelled context,
   or simply returning. If you can't answer, you have a leak.
3. **What limits how many run?** A worker pool or semaphore, sized for what's
   downstream: the API, the database, the file limit.
4. **Who waits for whom?** Something (`wg.Wait`, a final channel read) must
   keep `main` alive until the work is done, and must collect the errors.

Then run the tests with `-race`.

## Common mistakes

- **Returning before goroutines finish.** `main` returns, the program exits,
  and half your checks never ran. Wait with a WaitGroup or collect every
  result.
- **Touching a map or counter from several goroutines without a lock.**
  Maps are not safe for concurrent writes; Go may even crash with `fatal
  error: concurrent map writes`.
- **Holding a lock during slow work.** Locking around an HTTP call turns
  your concurrent program back into a sequential one.
- **Forgetting to close a channel** that someone ranges over. The `for range`
  waits forever: a deadlock or a leak.
- **Unlimited goroutines against a real service.** 10,000 simultaneous
  requests is a denial-of-service attack on your own database.
- **Expecting order.** Concurrent results come back in any order. Your test
  that passes "because the output looks right" will flake.

## Check yourself

1. A script resizes 1,000 images on a 4-core machine. Another fetches 1,000
   URLs. Which gains more from running 100 goroutines, and why?
2. Two goroutines each run `stock--` once on a shared `stock := 5`. Which
   final values are possible without a lock?
3. What is the difference between a data race and a race condition? Give an
   ops example of a race condition.
4. A worker goroutine does `for job := range jobs { ... }`, and the program
   hangs at the end. What is the most likely cause?
5. Your tool checks 2,000 hosts and opens a connection to each. It starts
   failing with "too many open files". What do you change?
6. A teammate says: "This laptop has one core, so my program can't be
   concurrent." Are they right?

## Answers

1. The URL fetcher. It is I/O-bound, so 100 waits overlap. Resizing is
   CPU-bound: 4 cores do at most 4 images at once, and the rest only add
   switching overhead.
2. 3 (both updates applied) or 4 (one update lost). `stock--` is really
   read, subtract 1, write back. If both goroutines read 5 before either
   writes, both write 4 and one update vanishes.
3. A data race is unsynchronised access to the same memory with at least one
   write; the race detector can find it. A race condition is any bug where
   the result depends on timing. Example: two deploy jobs both check "is
   anyone deploying?", both see no, and both deploy.
4. Nobody closed `jobs`. `range` over a channel only ends when it's closed,
   so the worker waits forever for another job.
5. Limit concurrency: a worker pool or a semaphore of, say, 50. Every
   connection still gets made, but only 50 file descriptors are open at once.
6. No. Concurrency is structure; one core can interleave many goroutines,
   and while one waits on the network another runs. What one core can't give
   you is parallelism.

## Key terms

| Term | Meaning |
|---|---|
| concurrency | structuring a program as independent tasks that can overlap |
| parallelism | actually running several things at the same instant |
| CPU core | a unit of the processor that runs one instruction stream |
| I/O-bound | mostly waiting on network, disk or other programs |
| CPU-bound | mostly computing; limited by core count |
| process | a running program with its own private memory |
| thread | an OS-scheduled path of execution sharing its process's memory |
| goroutine | Go's cheap task, run on OS threads by the Go runtime |
| GOMAXPROCS | how many threads may run Go code at the same instant |
| read-modify-write | read a value, change it, write it back; unsafe if shared |
| data race | unsynchronised access to shared memory with at least one write |
| race condition | a bug where the result depends on timing |
| race detector | `-race`: Go's tool that reports data races as they happen |
| mutex | a lock only one goroutine can hold at a time |
| critical section | the code between `Lock` and `Unlock` |
| channel | a typed pipe that passes values between goroutines |
| unbuffered channel | a channel with no queue: each send waits until a receiver takes the value |
| buffered channel | a channel with a queue of N: sends wait only when N values are already waiting |
| WaitGroup | a counter you can wait on until all goroutines finish |
| context | carries cancellation and deadlines through a call chain |
| deadlock | goroutines waiting on each other forever |
| goroutine leak | a goroutine blocked forever, never freed |
| starvation | a ready goroutine that rarely gets to run |
| semaphore | N tokens that limit how many tasks run at once |
| worker pool | a fixed set of goroutines pulling jobs from a channel |
| fan-out / fan-in | spreading work to many goroutines / merging many channels into one |
| pipeline | stages joined by channels, each doing one job |
| workqueue | a Kubernetes queue of object keys feeding controller workers |

## Watch

- [Concurrency Vs Parallelism!](https://www.youtube.com/watch?v=RlM9AfWf1WU) · ByteByteGo · 4 min: a quick animated picture of the two words this guide pins down, good to watch first.
- [Concurrency in Go](https://www.youtube.com/watch?v=LvgVSSpwND8) · Jake Wright · 19 min: the friendliest walk-through for a beginner, building goroutines, WaitGroup, channels, select and a worker pool one step at a time.
- [Concurrency is not Parallelism by Rob Pike](https://www.youtube.com/watch?v=oV9rvDllKEg) · gnbitcom · 31 min: the classic talk behind this guide's framing, with gophers moving books to show how structure, not cores, makes a design concurrent.
- [GopherCon 2017: Understanding Channels - Kavya Joshi](https://www.youtube.com/watch?v=KBZlN0izeiY) · Gopher Academy · 22 min: a deeper look at what a channel really is (a queue and a lock) and why a blocked goroutine waits; watch once the Tools sections feel solid.

## Go deeper

- [Rob Pike, "Concurrency is not parallelism"](https://go.dev/blog/waza-talk):
  the talk behind this guide's framing, with gophers moving books.
- [Go Concurrency Patterns: Pipelines and cancellation](https://go.dev/blog/pipelines):
  pipelines, fan-out, fan-in and stopping cleanly, step by step.
- [Data Race Detector](https://go.dev/doc/articles/race_detector): how to
  use `-race` and read its reports, with real-world race examples.
- [Share Memory By Communicating](https://go.dev/blog/codelab-share): a short
  explanation of Go's motto, with a URL-polling example.
- [Operating Systems: Three Easy Pieces](https://pages.cs.wisc.edu/~remzi/OSTEP/):
  a free book; its "Concurrency" chapters explain threads and locks
  underneath.
