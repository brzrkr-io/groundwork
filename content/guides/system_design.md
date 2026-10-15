# System design fundamentals

**In one sentence:** system design is deciding, before you build, which pieces a service is made of, how they talk to each other, how much load each must carry and what happens when one breaks, and then writing down why you made those choices.

**Why it matters:** most outages aren't caused by a wrong `if` statement. They come from design: one database everything depends on, a queue nobody sized, a "fast" call that crosses an ocean. Platform engineers review designs, size infrastructure and set reliability targets every week. This guide gives you the questions to ask and the arithmetic to answer them.

**You'll learn:**

- What a system design covers, and why you write it down
- How to turn a vague request into functional and non-functional requirements
- How to estimate load, storage and bandwidth on the back of an envelope
- Rough latency numbers that tell you what's cheap and what's slow
- How SLIs, SLOs and error budgets turn "reliable" into a number
- How to find single points of failure, and how to record a trade-off

## What system design is

Your team is asked: "Can we have a page that shows what version of every service is deployed?" You could start typing a handler. Or you could first ask:

- Who uses it, and how often? Five engineers, or every dashboard in the company?
- Where does the data come from? Who sends it, and what if they send it twice?
- What happens if the page is down during an incident, exactly when people need it?

Answering those before coding is *system design*. The output is a short *design doc*: a page or two that a reviewer can argue with before the code exists, when changing your mind is cheap.

A design doc usually has the same parts as this guide:

| Section | Answers |
|---|---|
| Requirements | What must it do, and how well? |
| Estimates | How much load, data and traffic? |
| Design | Which pieces, and how does a request flow through them? |
| Reliability | What can fail, and what happens then? |
| Trade-offs | What did we choose, what did we give up, and why? |

You'll write one in *Practice: deploy-status service*.

## Start with requirements

Requirements come in two kinds, and beginners usually forget the second.

### Functional requirements

*Functional requirements* say **what** the system does: its features, in plain verbs.

- CI can record "service `api`, version `1.4.2`, state `succeeded`".
- Anyone can ask for the latest deploy of one service.
- Unknown states are rejected.

### Non-functional requirements

*Non-functional requirements* say **how well** it does it. They decide the design far more than the features do.

| Quality | Question to ask | Example answer |
|---|---|---|
| Scale | How many users, requests, records? | 1,500 reads per second at peak |
| Latency | How fast must an answer be? | 95% of reads under 100 ms |
| Availability | How much downtime is acceptable? | 99.9% per month |
| Freshness | How stale may an answer be? | Up to 2 seconds |
| Durability | Can we ever lose data? | Losing a status is fine; CI resends |
| Cost | What's the budget? | One small VM and a managed database |

The same feature list with different non-functional answers gives a completely different system. "Show the deployed version" for five engineers is a script. For 10,000 dashboards it needs caches and replicas.

**Ask, don't assume.** When a requirement is missing, write your assumption down ("assume 300 services") so a reviewer can correct it.

## Back-of-the-envelope estimates

A *back-of-the-envelope estimate* is quick arithmetic with round numbers, good to within a factor of two or so. Its job is to tell you which **order of magnitude** you're in: 10 requests per second fits on a laptop; 100,000 does not.

### Numbers worth memorising

| Fact | Value | Round it to |
|---|---|---|
| Seconds in a day | 86,400 | 100,000 (10⁵) |
| Seconds in a month | about 2.6 million | 2.5 million |
| 1 KB, MB, GB, TB | 10³, 10⁶, 10⁹, 10¹² bytes | |
| 1 byte | 8 bits | |

Network speeds are quoted in **bits** per second, and storage in **bytes**. Mixing them up is an 8× mistake.

### Requests per second (QPS)

*QPS* (queries per second) is the load a service must handle. Start from a daily count:

```text
average QPS = requests per day / 86,400
```

Traffic isn't flat. Everyone deploys after stand-up and nobody at 3 a.m. So multiply by a *peak factor*, often 2 to 10:

```text
10 million requests a day / 86,400   ≈ 116 QPS average
116 × peak factor 5                  ≈ 580 QPS at peak
```

Design for the peak, not the average.

### Storage

```text
storage = items per day × bytes per item × days kept
```

A CI system stores 50,000 build logs a day, each about 200 KB, for 90 days:

```text
50,000 × 200 KB = 10 GB a day
10 GB × 90 days = 900 GB
```

That fits on one disk. Add replicas (×2 or ×3) and some headroom, and you know what to ask the storage team for.

### Bandwidth

```text
bandwidth = QPS × bytes per response × 8   (bits per second)
```

At 580 QPS, with 20 KB responses:

```text
580 × 20,000 bytes = 11.6 MB/s
11.6 MB/s × 8      ≈ 93 Mbit/s
```

A 1 Gbit/s network link has room. 50,000 QPS of the same responses would not.

### Servers

```text
servers = ceil(peak QPS / QPS one server can handle) + spares
```

If a load test shows one instance handles 200 QPS, then 580 QPS needs `ceil(2.9)` = 3 instances, plus one spare so a crash or a rolling deploy doesn't overload the rest. That spare is called *N+1*.

The same sums in Go, as you'll write them in *Capacity estimates: QPS, storage and bandwidth*:

```go
package main

import (
	"fmt"
	"math"
)

func main() {
	perDay := 10_000_000
	avg := float64(perDay) / 86_400
	peak := avg * 5
	servers := int(math.Ceil(peak/200)) + 1
	fmt.Printf("avg %.0f QPS, peak %.0f QPS, %d servers\n", avg, peak, servers)
}
```

It prints `avg 116 QPS, peak 579 QPS, 4 servers`. Step by step: divide by the seconds in a day, multiply by the peak factor, divide by one server's capacity, round **up** (2.9 servers means 3), then add the spare.

## Latency numbers every engineer should know

Estimates also cover **time**. These are rough, order-of-magnitude numbers for modern hardware. The exact values change; the ratios between them barely do.

| Operation | Roughly |
|---|---|
| Read a variable from CPU cache | 1 ns |
| Lock and unlock an uncontended mutex | 20 ns |
| Read from main memory (RAM) | 100 ns |
| Read 4 KB at random from an SSD | 20–100 µs |
| Round trip to another machine in the same datacentre | 0.5 ms |
| A simple indexed database query over the network | 1–5 ms |
| Spinning-disk seek | 5–10 ms |
| Round trip between continents | 100–150 ms |

Read it as a ladder: each step down is slower, and the bottom row is about 100 million times slower than the top.

- **Memory is cheap, the network is expensive.** A cache in your process answers in nanoseconds; asking the database answers in milliseconds, tens of thousands of times slower.
- **Round trips add up.** A handler that makes 20 sequential database calls at 2 ms each spends 40 ms just waiting. One query that fetches all 20 rows costs 2 ms.
- **Distance is physics.** Light in fibre needs about 70 ms to cross the Atlantic and back. No code change fixes that; putting a copy of the data nearer the user does.

## SLIs, SLOs and error budgets

"The service should be reliable" can't be checked. Site reliability engineering turns it into numbers.

- A *service level indicator* (SLI) measures what users feel, as a ratio of good events to all events: `successful requests / all requests`, or `requests faster than 300 ms / all requests`.
- A *service level objective* (SLO) is the target for an SLI over a window: "99.9% of requests succeed, measured over 30 days".
- A *service level agreement* (SLA) is a contract with customers, with penalties when it's missed. SLAs are set looser than SLOs, so you notice trouble before it costs money.

### The error budget

If the SLO is 99.9%, then 0.1% of requests may fail. That allowance is the *error budget*:

```text
budget = (1 − SLO) × events in the window

99.9% SLO, 30 days, 10 million requests:
0.001 × 10,000,000 = 10,000 failed requests allowed
0.001 × 30 days × 1,440 min = 43.2 minutes of full outage allowed
```

| SLO | Downtime allowed per 30 days |
|---|---|
| 99% | 7.2 hours |
| 99.9% | 43.2 minutes |
| 99.99% | 4.3 minutes |
| 99.999% | 26 seconds |

Each extra nine costs roughly ten times more effort: more replicas, more automation, faster on-call.

### Why a budget, not perfection

100% is the wrong target. Users can't tell 99.99% from 100% through their own flaky Wi-Fi, and chasing it means never deploying. The budget is permission to take risks:

- Budget left: ship features, run experiments, do risky migrations.
- Budget spent: freeze risky deploys and spend the time on reliability.

That rule, agreed in advance, is an *error budget policy*. It turns "dev wants speed, ops wants stability" into a number both sides accept.

### Burn rate

The *burn rate* is how fast you're spending the budget. A burn rate of 1 spends exactly the whole budget by the end of the window. A burn rate of 14.4 sustained for an hour spends 2% of a 30-day budget, which is worth paging someone for. You'll compute both in *SLOs: error budgets and burn rates*.

### Dependencies multiply

If your service calls three others in a row, and each is 99.9% available, the chain is at best:

```text
0.999 × 0.999 × 0.999 ≈ 0.997   (99.7%)
```

You can't promise more than your dependencies give you, unless you design around them with caches, fallbacks or redundancy.

## Single points of failure and redundancy

A *single point of failure* (SPOF) is any one part whose failure takes the whole system down. Draw the request path and put your finger on each box: "if this dies, what happens?"

Common SPOFs in real systems:

- one database instance with no replica;
- every server in one availability zone;
- one NAT gateway, one load balancer, one DNS provider;
- a TLS certificate that nobody is alerted about before it expires;
- the one person who knows how the deploy script works.

*Redundancy* means having more than one of something, so another copy takes over:

| Pattern | How it works | Cost |
|---|---|---|
| N+1 | One more instance than the load needs | Cheap, common for stateless services |
| Active-passive | A standby waits and takes over on failure (*failover*) | Standby sits idle; failover takes time |
| Active-active | Several copies all serve traffic | Must keep their data in sync |
| Multi-zone | Copies in separate datacentres | Cross-zone traffic costs money and latency |

Redundancy only helps if failures are **independent**. Two replicas each 99% available give `1 − 0.01 × 0.01 = 99.99%`, but only if they don't share a power supply, a zone, a config file or a bad deploy. Rolling out a broken config to every replica at once defeats all of them. Keeping failures small is called limiting the *blast radius*.

Copies of data across machines bring their own problems: lag, conflicts, split brain. Those are covered in the *Distributed systems basics* guide.

## Trade-offs, written down

Every design decision gives something up. Caching makes reads fast but answers stale. A queue absorbs bursts but delays the work. Recurring trade-offs:

| You gain | You pay with |
|---|---|
| Speed (caching, replicas) | Freshness, consistency |
| Availability (answer from a copy) | Sometimes-stale answers |
| Throughput (queues, batching) | Latency for each item |
| Redundancy | Money, complexity |
| Simplicity (one database) | A ceiling on scale |

**Write each decision down.** Six months later, someone will ask "why didn't we just use Kafka?", and "it seemed right" isn't an answer. A short *architecture decision record* (ADR) is enough:

```text
Decision:  Remove cached statuses on write instead of updating them.
Context:   Two workers may write the same service at once.
Gave up:   One extra database read after each deploy.
Because:   An update can leave an older value in the cache forever;
           a removal can't.
```

Good trade-off notes name the alternative, the cost and the reason. "We chose Postgres" is not a trade-off; "Postgres over DynamoDB: we lose automatic sharding, which we don't need below 10,000 writes per second, and keep SQL, which the team knows" is.

## Worked example: an internal link shortener

Your company wants *go links*: type `go/runbook-db` in a browser and land on the right wiki page. Let's design it end to end.

### 1. Requirements

Functional:

- Create a short name for a URL (`runbook-db` → `https://wiki.internal/...`).
- Visiting `go/<name>` redirects to the URL.
- Owners can change where a name points.

Non-functional:

- Redirects must be fast (under 50 ms) and very available: on-call engineers use them during incidents.
- An edited link may take up to a minute to take effect everywhere.
- Links must never be lost.

### 2. Estimates

Assume 5,000 employees, each clicking 20 links a day, and 500 new links a day:

```text
reads:   5,000 × 20 = 100,000 a day / 86,400 ≈ 1.2 QPS average
peak:    × 10 (an incident: everyone opens the same runbook) ≈ 12 QPS
writes:  500 a day ≈ 0.006 QPS
storage: 500 a day × 365 × 500 bytes ≈ 91 MB a year
```

That's tiny. Reads outnumber writes 200 to 1, and the whole data set fits in memory.

### 3. API

```text
POST /links          {"name": "runbook-db", "url": "https://..."}  -> 201
PUT  /links/{name}   {"url": "https://..."}                         -> 200
GET  /{name}         -> 302 Found, Location: https://...
```

Why `302` (temporary redirect) and not `301` (permanent)? Browsers may cache a 301 with no expiry, so an edited link would never take effect for someone who already clicked it. That's a trade-off: 302 sends every click to the service, which the estimates say it can easily take.

### 4. Design

```text
browser -> load balancer -> link service (2 replicas) -> Postgres (primary + standby)
                                 |
                          in-memory cache, 60 s TTL
```

- Two stateless replicas behind a load balancer: no single instance is a SPOF.
- Postgres with a synchronous standby in another zone, plus nightly backups: a write isn't confirmed until two zones have it, so links are never lost. (An *asynchronous* standby can be a few seconds behind and lose the newest links when it takes over.)
- An in-process cache with a 60-second TTL: redirects are answered from memory, and the 60 s TTL is exactly the freshness requirement.

### 5. Failure modes

| What fails | What happens | Mitigation |
|---|---|---|
| One replica | The load balancer's health check removes it | N+1 replicas |
| Postgres primary | Writes fail until failover; cached redirects keep working | Standby and failover; serve from cache |
| The whole service | `go/` links stop during an incident | Publish a static export of runbook links |

### 6. The same design at a different scale

Now marketing wants to put short links in SMS alerts to 10 million customers, who'll click 100 million times a day:

```text
100,000,000 / 86,400 ≈ 1,160 QPS average, × 5 peak ≈ 5,800 QPS
```

Same features, three orders of magnitude more load. Now you'd add a shared cache, a CDN in front of redirects, read replicas, and a different way to generate names (a counter encoded in base 62 gives 62⁷ ≈ 3.5 trillion seven-character codes). The *Scaling: caches, queues and load balancers* guide covers those tools. The estimate is what told you that you needed them.

## Common mistakes

- **Designing before estimating.** Building sharding and Kafka for 2 requests a second, or a single SQLite file for 50,000.
- **Forgetting non-functional requirements.** The features are easy; the latency, availability and freshness targets decide the design.
- **Designing for the average.** Peaks are 2 to 10 times higher, and outages happen at the peak.
- **Bits versus bytes.** Network links are quoted in bits per second; multiply bytes by 8.
- **Aiming for 100%.** It can't be reached, and chasing it stops all change. Pick an SLO and spend the budget.
- **Redundancy that shares a fate.** Two replicas in one zone, behind one load balancer, with one config, fail together.
- **Undocumented trade-offs.** If the reason isn't written down, the next person will "fix" the decision and reintroduce the problem it solved.

## Check yourself

1. A service handles 26 million requests a day with a peak factor of 4. Roughly what peak QPS must it handle?
2. Give one functional and one non-functional requirement for a service that records deploy events from CI.
3. An API returns 50 KB responses at 2,000 requests per second. Roughly how much bandwidth is that in Mbit/s?
4. An SLO says 99.9% of requests succeed over 30 days, and the service gets 20 million requests in that window. How many may fail?
5. A handler makes 30 database calls one after another, each a 2 ms network round trip. Where does the time go, and what's the usual fix?
6. Two replicas of a service run in the same availability zone with the same config. Why isn't that as redundant as it looks?
7. Why does an internal link shortener answer with HTTP 302 rather than 301 when links can be edited?

## Answers

1. 26,000,000 / 86,400 ≈ 300 QPS on average; × 4 ≈ 1,200 QPS at peak. Design for the peak.
2. Functional: "CI can record a service's version and state." Non-functional: "accepting an event takes under 50 ms" or "a repeated event must not change the result". Functional says what; non-functional says how well.
3. 2,000 × 50,000 bytes = 100 MB/s; × 8 = 800 Mbit/s. That's most of a 1 Gbit/s link, so it's time for compression or a CDN.
4. The error budget is (1 − 0.999) × 20,000,000 = 20,000 failed requests.
5. 30 × 2 ms = 60 ms spent waiting on the network, not computing. Fetch everything in one query (or a few in parallel) to cut the round trips.
6. Their failures aren't independent: a zone outage or a bad config rollout takes both down at once. Spread replicas across zones and roll changes out gradually.
7. Browsers may cache a 301 with no expiry, so people who clicked before an edit would keep going to the old URL. A 302 sends every click to the service, which can then return the current target.

## Key terms

| Term | Meaning |
|---|---|
| System design | Choosing a system's parts, connections and capacity before building it |
| Design doc | A short written plan a reviewer can argue with before code exists |
| Functional requirement | What a system must do: its features |
| Non-functional requirement | How well it must do it: scale, latency, availability, cost |
| Back-of-the-envelope estimate | Quick arithmetic with round numbers to find the order of magnitude |
| QPS | Queries (requests) per second |
| Peak factor | How much busier the busiest moment is than the average |
| N+1 | Running one more instance than the load needs |
| SLI | Service level indicator: a measured ratio of good events to all events |
| SLO | Service level objective: the target for an SLI over a time window |
| SLA | Service level agreement: a contract with penalties, looser than the SLO |
| Error budget | The failures an SLO allows: (1 − SLO) × events |
| Burn rate | How fast the error budget is being spent; 1 means exactly on schedule |
| Single point of failure | One part whose failure takes the whole system down |
| Redundancy | Extra copies so another takes over when one fails |
| Failover | Switching work from a failed component to its standby |
| Blast radius | How much of the system one failure can affect |
| Architecture decision record | A short note of a decision, its alternatives, its cost and its reason |

## Watch

- [Latency Numbers Programmer Should Know: Crash Course System Design #1](https://www.youtube.com/watch?v=FqR5vESuKe0) · ByteByteGo · 6 min: the latency ladder from CPU cache to cross-continent round trips, animated.
- [SLIs, SLOs, SLAs, oh my! (class SRE implements DevOps)](https://www.youtube.com/watch?v=tEylFyxbDLE) · Google Cloud Tech · 8 min: Google's SREs explain the three terms and how they relate.
- [The Art of SLOs (Service Level Objectives)](https://www.youtube.com/watch?v=E3ReKuJ8ewA) · Google Cloud Tech · 4 min: a short introduction to Google's SLO workshop: measure what users feel, then spend the error budget on purpose.
- [System Design Interview: A Step-By-Step Guide](https://www.youtube.com/watch?v=i7twT3x5yv8) · ByteByteGo · 10 min: the order this guide follows (requirements, estimates, design, deep dive), as interviewers expect it.

## Go deeper

- [Service Level Objectives](https://sre.google/sre-book/service-level-objectives/) (Google SRE book): the chapter that defined SLIs, SLOs and SLAs, with examples of choosing them.
- [Embracing Risk](https://sre.google/sre-book/embracing-risk/) (Google SRE book): why 100% is the wrong target, and how error budgets work as policy.
- [Numbers Every Programmer Should Know By Year](https://colin-scott.github.io/personal_website/research/interactive_latency.html) (Colin Scott): an interactive version of the latency table, showing how each number changed over the years.
- [Design Docs at Google](https://www.industrialempathy.com/posts/design-docs-at-google/) (Malte Ubl): what goes into a design doc and when one is worth writing.
- [Documenting Architecture Decisions](https://www.cognitect.com/blog/2011/11/15/documenting-architecture-decisions) (Michael Nygard): the original short post that introduced ADRs.
