# Distributed systems basics

**In one sentence:** a distributed system is several programs on different machines that must cooperate over a network that can be slow, lossy or down, so every call can fail in ways a single program never sees.

**Why it matters:** your deploy tool, your CI pipeline and every service you run on call already are distributed systems; the outages you will be paged for are mostly about timeouts, retries, stale caches and one slow dependency dragging everything else down.

**You'll learn:**

- Why "no answer" never tells you whether the other side did the work.
- How to use timeouts, retries with backoff and jitter, and idempotency together.
- What at-most-once, at-least-once and "exactly-once" really mean.
- Consistency versus availability in plain words, and why clocks can't be trusted.
- How rate limits, circuit breakers and caches protect a system, and what they cost.
- How logs, metrics and traces let you see inside a running system.

## One program becomes many

Picture a small deploy tool. To ship a release it makes three calls:

```text
  deploy tool
      |
      +--1--> container registry   "does image web:1.8.2 exist?"
      +--2--> Kubernetes API       "set web's image to web:1.8.2"
      +--3--> Slack API            "post: web 1.8.2 is rolling out"
```

In one Go program, a function call always arrives and always returns. Here, each arrow crosses a network to a machine you don't control, and any of them can be slow, down, or half-working.

A *distributed system* is any set of programs on separate machines that must cooperate over a network to get one job done. A Kubernetes cluster is one; so is your laptop plus GitHub plus a CI runner. The hard part is not the code in each box. It's the arrows.

## The network is not a function call

In the 1990s engineers at Sun wrote down the false assumptions everyone makes about networks, now called the *fallacies of distributed computing*. Each has burned an ops team somewhere:

| Fallacy | What really happens |
|---|---|
| The network is reliable | Packets drop; a switch reboots mid-deploy |
| Latency is zero | Each call costs 1 ms in a datacentre, 100+ ms across regions |
| Bandwidth is infinite | Pulling a 2 GB image to 300 nodes saturates the link |
| The network is secure | Anyone on the path may read or alter traffic |
| Topology doesn't change | Pods get new IPs; DNS answers change |
| There is one administrator | Your call crosses teams, clouds and vendors |
| Transport cost is zero | Cross-region traffic is billed per gigabyte |
| The network is homogeneous | Different MTUs, proxies and TLS versions on the way |

The *Networking from the ground up* guide explains the layers underneath. This guide is about what to do given that they fail.

## Partial failure: slow or dead?

In one program, things work or the whole program crashes. In a distributed system, one piece can fail while the rest keep running. That is *partial failure*, the defining problem of the field.

Say your deploy tool sends "set web's image" and hears nothing back. What happened?

```text
  you ----request---->X                 (1) request lost: nothing changed
  you ----request----> API  (crashes)   (2) API died: maybe changed
  you ----request----> API ---reply-->X (3) reply lost: it DID change
  you ----request----> API ...thinking  (4) just slow: reply coming soon
```

From your side, all four look identical: silence. You **cannot** tell a slow server from a dead one, or "never got it" from "did it, but the answer got lost". Every technique below is a way to live with that.

## Timeouts: decide how long to wait

Without a limit, case (4) can hang your tool forever. A *timeout* is the longest you will wait before treating the call as failed.

- **Always set one.** Go's default `http.Client` has no timeout at all. You'll set them in *Context and cancellation* and in the *Production services* module.
- **Pick it from data.** If the API normally answers in 200 ms and its 99th-percentile time is 900 ms, a 2 s timeout is sane. Thirty seconds is just a slower hang.
- **A timeout is not an answer.** After it fires you still don't know whether the work happened. That's why the next two sections go together.

## Retries, backoff and jitter

Many failures are brief: a pod restarting, a dropped connection, a `503 Service Unavailable`. Trying again often works. That's a *retry*.

Retrying badly makes outages worse. Imagine the registry hiccups and 1,000 CI jobs all retry instantly, again and again. The struggling registry now gets several times its normal traffic. That's a *retry storm*, and it can keep a service down long after the original problem is gone.

Two fixes:

- *Exponential backoff*: wait longer after each failure (100 ms, 200 ms, 400 ms, ...), up to a cap. Struggling servers get breathing room.
- *Jitter*: add randomness to each wait, so 1,000 clients don't all come back in the same millisecond.

```go
package main

import (
	"fmt"
	"math/rand/v2"
	"time"
)

// ceiling doubles base once per attempt, but never goes above limit.
func ceiling(attempt int, base, limit time.Duration) time.Duration {
	c := base
	for i := 0; i < attempt && c < limit; i++ {
		c *= 2
	}
	if c > limit {
		c = limit
	}
	return c
}

func main() {
	base, limit := 100*time.Millisecond, 2*time.Second
	for attempt := 0; attempt < 6; attempt++ {
		c := ceiling(attempt, base, limit)
		wait := rand.N(c) // full jitter: anywhere from 0 up to c
		fmt.Printf("retry %d: ceiling %v, wait %v\n",
			attempt+1, c, wait.Round(time.Millisecond))
	}
}
```

One run printed this. The ceilings are always the same; the waits change every run:

```text
retry 1: ceiling 100ms, wait 96ms
retry 2: ceiling 200ms, wait 12ms
retry 3: ceiling 400ms, wait 251ms
retry 4: ceiling 800ms, wait 13ms
retry 5: ceiling 1.6s, wait 850ms
retry 6: ceiling 2s, wait 678ms
```

Step by step: `ceiling` starts at 100 ms and doubles once per attempt, stopping at 2 s. `rand.N(c)` picks a random duration from 0 up to (not including) `c`. Picking anywhere in the whole range is called "full jitter".

Rules that keep retries safe:

- **Only retry what might succeed next time.** A `503` or a timeout, yes. A `400 Bad Request` or `403 Forbidden` will fail the same way forever.
- **Cap the attempts** and give up with a clear error.
- **Retry at one layer only.** If the CLI, the library and the proxy each retry 3 times, one failing call becomes 3 × 3 × 3 = 27 requests.
- **Only retry operations that are safe to repeat.** Which brings us to idempotency.

You'll build this in *Retries: backoff and jitter* and *Retries: retryable HTTP errors*.

## Idempotency: safe to do twice

Your tool calls "create order for a db.large instance", the request times out, and you retry. Was that case (1), where nothing happened, or case (3), where the order exists and the reply got lost? If it was (3), you now pay for two databases.

An operation is *idempotent* when doing it twice has the same effect as doing it once.

| Operation | Idempotent? | Why |
|---|---|---|
| "Set replicas to 3" | Yes | Second time, it's already 3 |
| "Add 1 replica" | No | Each call adds another |
| "Delete pod web-7f9c" | Yes | Second time, it's already gone |
| "Create an order" | No | Each call makes a new order |

HTTP builds this into its methods. `PUT /services/web` means "make this resource look like this body", so repeating it is harmless. `POST /orders` means "create something new", so repeating it creates duplicates. `GET`, `PUT` and `DELETE` are defined as idempotent; `POST` is not.

To make a non-idempotent call safe to retry, use an *idempotency key*: the client invents a unique ID for the operation and sends it with every attempt. The server remembers keys it has seen and replays the first answer instead of doing the work again.

```go
package main

import "fmt"

// Orders is a toy order API that remembers idempotency keys.
type Orders struct {
	created []string          // every order we really made
	byKey   map[string]string // idempotency key -> order ID
}

// Create makes an order, unless this key was already used. Then it
// returns the first answer again and does nothing new.
func (o *Orders) Create(key, item string) string {
	if id, seen := o.byKey[key]; seen {
		return id
	}
	id := fmt.Sprintf("order-%d", len(o.created)+1)
	o.created = append(o.created, item)
	o.byKey[key] = id
	return id
}

func main() {
	o := &Orders{byKey: map[string]string{}}
	fmt.Println(o.Create("key-7f3a", "db.large")) // first try
	fmt.Println(o.Create("key-7f3a", "db.large")) // retry after a timeout
	fmt.Println(o.Create("key-91c2", "db.small")) // a different order
	fmt.Println("orders really created:", len(o.created))
}
```

It prints:

```text
order-1
order-1
order-2
orders really created: 2
```

The retry with `key-7f3a` found the key in the map and got `order-1` back without creating anything. Stripe and many AWS APIs work this way (AWS often calls the key a "client token"). A real server also stores keys durably, expires them after a while, and rejects a reused key with a different body; you'll handle that in *Idempotency keys: scope, expire and fingerprint*.

## Delivery guarantees

Put timeouts, retries and idempotency together and you get three possible promises about how often a message or request takes effect:

| Guarantee | How | Risk |
|---|---|---|
| At-most-once | Send once, never retry | Some work silently lost |
| At-least-once | Retry until acknowledged | Some work done twice |
| Exactly-once | At-least-once + dedup | Needs keys or idempotent work |

True *exactly-once delivery* over an unreliable network is impossible: case (3) above means you'll sometimes resend. What systems really offer is *effectively-once processing*: deliver at least once, and have the receiver ignore duplicates.

The ops habit: **assume at-least-once, make handlers idempotent.** A webhook, an SQS message, a Kubernetes event: any of them may arrive twice.

## Consistency versus availability

Imagine a config service with two copies, one in each datacentre. The link between the datacentres breaks. Someone writes `max_conns=500` to copy A. A client then asks copy B.

Copy B has two choices:

- **Refuse to answer** until it can check with A. Every answer is current, but B is unavailable. This favours *consistency*.
- **Answer with what it has** (`max_conns=200`). B stays available, but the answer is stale. This favours *availability*.

That's the *CAP theorem* without the maths: when the network is split (a *partition*), a replicated system must give up either always-current answers or always-answering. Partitions happen, so the real question is which you give up during one.

Most systems you use pick availability and accept *eventual consistency*: copies may disagree for a while, but if writes stop, they all converge to the same value.

- **DNS:** you change a record, but resolvers keep the old answer until its TTL runs out.
- **Read replicas:** a database replica can lag the primary by seconds.
- **Kubernetes:** controllers act on a cached copy of the cluster that trails the API server slightly. The *Declarative systems and control loops* guide shows why that's fine.

Ops consequence: "I wrote it, read it back and got the old value" is often not a bug. Know which systems are eventually consistent before writing a check that reads right after a write.

## Clocks lie

Each machine has its own clock, and clocks drift apart. That difference is *clock skew*. NTP keeps it small, but small can mean milliseconds on a good day and minutes when NTP is broken. Clocks can even jump backwards when corrected.

So don't decide which event happened first by comparing wall-clock timestamps from different machines:

```text
  node-a log: 10:00:00.120  "deleted volume vol-9"
  node-b log: 10:00:00.080  "attached volume vol-9"
```

Did node-b attach before the delete? You can't tell: node-b's clock might be 100 ms behind.

- To **order** changes, use a counter that one authority hands out: a database sequence, a version number, Kubernetes' `resourceVersion`.
- To **measure** durations within one program, Go already does the right thing: `time.Now` carries a monotonic clock reading, so `time.Since(start)` isn't fooled by clock corrections. See *Time and durations: cert expiry, uptime, maintenance windows*.
- When reading logs from many machines during an incident, treat timestamps as approximate.

## Protecting systems from overload

A service has a fixed capacity. When demand exceeds it, everything slows down for everyone. Three tools keep that under control.

### Rate limiting

A *rate limit* caps how many requests a client may make per second. A common design is the *token bucket*: the bucket refills at a steady rate, each request spends a token, and an empty bucket means "wait" or `429 Too Many Requests`. AWS and the Kubernetes API both throttle you this way, so your automation must handle `429` by backing off. You'll write one in *Rate limiting: token buckets*.

### Backpressure

*Backpressure* means a busy part of a system pushes back on whoever is sending it work, instead of piling the work up. A bounded queue that blocks or rejects when full is backpressure. An unbounded queue just hides overload until memory runs out. A worker pool fed by a buffered channel, from the *Concurrency* module, is a built-in example.

### Circuit breakers

Say a payment API starts timing out after 10 s. Every request into your service now waits 10 s for it, ties up a goroutine and a connection, and soon your service falls over too. That's a *cascading failure*.

A *circuit breaker* sits in front of the dependency and counts failures:

```text
          too many failures
  CLOSED ------------------> OPEN  (fail fast, don't call)
    ^                          |
    |  trial call succeeds     | after a cool-down
    +----------- HALF-OPEN <---+
                 (let one call through to test)
                 trial fails -> back to OPEN
```

When open, calls fail instantly with a clear error instead of waiting. You lose that feature for a while but keep the rest of the service alive. You met a tiny version in *Functions as values: retries and a circuit breaker* and build a real one in *Circuit breakers: protect an HTTP client*.

## Caching and staleness

A *cache* keeps a copy of a slow or expensive answer so you don't fetch it every time: the list of instances from the AWS API, a token, a DNS answer.

Every cache trades freshness for speed. The copy is *stale* the moment the source changes. Questions to answer for every cache:

- **How stale is acceptable?** That sets the TTL (time to live). Thirty seconds for an instance list is often fine; zero tolerance means don't cache.
- **What happens at expiry?** If 500 requests all miss at once and all hit the backend, that's a *cache stampede*. The fix is to let one request load the value while the others wait for it, as in *TTL caching: collapse concurrent loads*.
- **What if the source is down?** Serving a slightly stale value is sometimes better than an error. Decide on purpose.

## Observability: seeing inside

You can't attach a debugger to 40 production pods. You learn what happened from what the system recorded. *Observability* is how well you can answer new questions about a running system from its outputs.

| Signal | Answers | Example |
|---|---|---|
| Logs | What happened, in detail? | `level=ERROR msg="push failed" image=web:1.8.2` |
| Metrics | How much, how often, how fast? | `http_requests_total{code="503"} 1240` |
| Traces | Where did this one request spend its time? | deploy 4.1 s: registry 0.2 s, k8s 3.8 s |

- **Logs** are events. Structured logs can be searched and filtered: *Structured logging with slog*.
- **Metrics** are cheap numbers over time, ideal for alerts: *Prometheus-style metrics: counters and /metrics*.
- **Traces** follow one request across services, stitched together by an ID each hop passes on. You add request IDs in *Middleware: logging, request IDs and panics*.

In an incident you usually go metrics (what's wrong, since when), then traces (which hop), then logs (exactly what it said).

## Common mistakes

- **No timeout on network calls.** One stuck dependency then hangs your tool or exhausts your service.
- **Retrying instantly, forever, at every layer.** Back off, add jitter, cap attempts, retry in one place.
- **Retrying non-idempotent calls.** Without an idempotency key, retrying "create" can make duplicates.
- **Write-then-read checks against eventually consistent systems.** Expect old values for a while, or read from the authority.
- **Caching without deciding on staleness.** Every cache needs a TTL and a plan for expiry and outages.
- **Logs as free text only.** `deploy failed` with no service, version or request ID is useless at 3 a.m.

## Check yourself

1. Your tool sends `POST /volumes` to create a disk and gets a timeout. Name two different things that may have happened on the server.
2. Three hundred CI jobs fail to reach the registry at the same moment. Each retries after exactly 1 s. What goes wrong, and what fixes it?
3. A queue promises at-least-once delivery. What must your message handler do?
4. You update a DNS record, then immediately run `dig` from another office and see the old IP. Is DNS broken?
5. A dependency you call starts timing out on every request. Which resilience pattern stops it from taking your service down too, and how?

## Answers

1. The request may never have arrived (no disk), or the disk was created and the reply was lost, or the server is still working on it. You can't tell which, so retry with an idempotency key or check before creating again.
2. All 300 retry in the same second, hitting the registry with a synchronized spike exactly when it's weak: a retry storm. Exponential backoff spreads attempts over time; jitter spreads the clients apart.
3. Be idempotent: processing a message twice must have the same effect as once, for example by recording processed message IDs and skipping repeats.
4. Probably not. Resolvers cache answers until the record's TTL expires. DNS is eventually consistent; the other office sees the new IP once its cache expires.
5. A circuit breaker. After repeated failures it opens and fails calls instantly instead of waiting on timeouts, so goroutines and connections aren't tied up. After a cool-down it lets a trial call through.

## Key terms

| Term | Meaning |
|---|---|
| Distributed system | Programs on several machines cooperating over a network |
| Partial failure | One part fails while others keep running |
| Timeout | The longest you'll wait before treating a call as failed |
| Retry storm | Many clients retrying together and overloading a service |
| Exponential backoff | Waiting longer after each failed attempt |
| Jitter | Randomness added to waits so clients don't sync up |
| Idempotent | Doing it twice has the same effect as once |
| Idempotency key | A client-chosen ID that lets the server spot a repeat |
| At-least-once | Delivery that retries, so duplicates are possible |
| Effectively-once | At-least-once delivery plus duplicate removal |
| CAP theorem | During a partition, choose consistency or availability |
| Eventual consistency | Copies may differ for a while but converge |
| Clock skew | Different machines' clocks disagreeing |
| Rate limit | A cap on requests per unit of time |
| Backpressure | A busy component pushing back on new work |
| Circuit breaker | Fails fast after repeated failures to stop cascades |
| Cache stampede | Many requests reloading an expired cache entry at once |
| Observability | Seeing inside a system through logs, metrics and traces |

## Watch

- [CAP Theorem Simplified](https://www.youtube.com/watch?v=BHqjEjzAicA) · ByteByteGo · 6 min: a short animated version of the consistency-versus-availability choice when the network splits.
- [Distributed Systems 2.3: System models](https://www.youtube.com/watch?v=y8f7ZG_UnGI) · Martin Kleppmann · 21 min: a Cambridge lecture on lost messages, crashed nodes and unpredictable delays, the partial failures this guide starts from; the rest of the series goes further.
- [Timeouts, Retries and Idempotency In Distributed Systems](https://www.infoq.com/presentations/distributed-systems-resiliency/) · Sam Newman, QCon London 2025 · 51 min: the deeper talk for this guide, covering how to pick timeouts, why retries need jitter, and how request IDs make retries safe; it has a transcript to read on the phone too.

## Go deeper

- [Fallacies of distributed computing](https://en.wikipedia.org/wiki/Fallacies_of_distributed_computing) (Wikipedia): the eight assumptions, with their history.
- [Exponential Backoff And Jitter](https://aws.amazon.com/blogs/architecture/exponential-backoff-and-jitter/) (AWS Architecture Blog): graphs showing why jitter beats plain backoff.
- [Designing robust and predictable APIs with idempotency](https://stripe.com/blog/idempotency) (Stripe): idempotency keys explained by a team that relies on them.
- [Addressing Cascading Failures](https://sre.google/sre-book/addressing-cascading-failures/) (Google SRE book): how overload spreads and the tools that stop it.
- [CircuitBreaker](https://martinfowler.com/bliki/CircuitBreaker.html) (Martin Fowler): the pattern's states, with a short example.
