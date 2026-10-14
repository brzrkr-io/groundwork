# Practice: deploy-status service

"What version of `api` is running right now?" gets asked in every incident
channel. Here you design and build the small service that answers it: CI
pipelines send it deploy events, and dashboards ask it for the latest
status. You write a short design doc first, then the code.

**New here:** nothing; practice with a bounded queue, an idempotent
consumer, cache-aside reads and a written design.

## Learn

### Recap

- A *design doc* states requirements, estimates, the design and its
  trade-offs before anyone writes code (see the *System design
  fundamentals* guide).
- A **bounded queue** is a buffered channel. A `select` with a `default`
  case sends without waiting, so a full queue can say "no" at once: that's
  backpressure (*Work queues: bounded queues and backpressure*).
- Delivery is **at-least-once**, so the consumer remembers which event IDs
  it has applied and skips repeats (*Work queues: fix the at-least-once
  consumer*).
- **Cache-aside**: read the cache; on a miss read the database and fill the
  cache. On a write, update the database, then *remove* the cached entry
  (*LRU caches: fix the stale cache-aside reads*).
- `errors.Is(err, ErrX)` finds `ErrX` even inside a wrapped error, and
  `fmt.Errorf("%w: reason", ErrX)` wraps it (*Wrapping errors*).

> **Why is it written like this?** `Submit`, `Apply` and `Lookup` are plain
> methods, and the HTTP handlers only translate HTTP to and from them. That
> split is the author's choice: the logic is tested without HTTP, and the
> handlers stay a few lines each.

### How the pieces fit

```text
CI ── POST /events ──> handleEvent ── Submit ──> queue (chan Event, bounded)
                                                    │
                                          Run (one worker goroutine)
                                                    │ Apply
                                                    v
                                   statuses map  (the database)
                                        ^           + remove from cache
                                        │ miss
dashboard ── GET /status/api ──> handleStatus ── Lookup ── cache (LRU) hit
```

| File | What's in it |
|---|---|
| `DESIGN.md` | Your design doc (task 1). |
| `exercise.go` | `Service`: `Submit`, `Apply`, `Lookup`, the worker `Run` and the HTTP handlers (tasks 2–5). |
| `cache.go` | A finished LRU `Cache` with `Get`, `Put`, `Remove` and `Len`. Read it; don't change it. |

`Service` has one mutex, `s.mu`, that guards `statuses`, `processed`,
`cache` and `dbReads`. `Apply` and `Lookup` already lock it.

## Your task

Your platform team wants this service by Friday. Write the design, then
make every test pass.

### 1. Write the design

**Goal:** fill in `DESIGN.md` so it reads like a real (short) design doc.

The test checks that `DESIGN.md`:

- contains no `TODO` anywhere (delete every placeholder line);
- has the four headings `## Requirements`, `## Estimates`, `## Design` and
  `## Trade-offs`, spelled exactly like that;
- has at least 25 words in each of those sections;
- has at least one number in **Estimates**;
- mentions the words `queue` and `cache` in **Design**;
- has at least two bullet points (lines starting with `- `) in
  **Trade-offs**.

What to write in each section:

| Section | Answer these |
|---|---|
| Requirements | What must it do? How fast, how fresh, how reliable? |
| Estimates | How many events a day and at peak? How many reads per second? How much storage? |
| Design | Which pieces exist, and how does one event, and one read, flow through them? |
| Trade-offs | For each choice: what you picked, what you gave up, and why that's OK. |

Make up plausible numbers: say 300 services that each deploy 10 times a
day. Show the arithmetic, as in the *System design fundamentals* guide.

**Think it through:**
- Which happens more often here, writes (deploy events) or reads (dashboards
  refreshing)? What does that tell you about the cache?
- What happens to CI when the queue is full? Why is that better than
  queueing forever?

### 2. Accept events with backpressure

**Goal:** make `Submit(ev)` validate the event, then queue it without ever
waiting.

Check in this order:

1. If `ID`, `Service` or `Version` is empty, or `State` is not one of
   `"started"`, `"succeeded"` or `"failed"`, return an error that wraps
   `ErrInvalidEvent`. The wording after it is up to you.
2. If `s.queue` has room, send `ev` and return `nil`.
3. Otherwise return `ErrQueueFull`.

```go
s := New(2, 100)                                  // queue holds 2
s.Submit(Event{ID: "e1", Service: "api", Version: "1.4.2", State: "started"}) // => nil
s.Submit(Event{ID: "e2", Service: "api", Version: "1.4.2", State: "done"})    // => invalid deploy event: ...
s.Submit(Event{ID: "e3", Service: "web", Version: "0.9.0", State: "failed"})  // => nil
s.Submit(Event{ID: "e4", Service: "db", Version: "16.2", State: "started"})   // => event queue is full
```

**Think it through:**
- Which `select` case lets a send fail instead of blocking?
- An invalid event is invalid whether or not the queue is full. Which check
  comes first?

### 3. Apply each event once

**Goal:** make `Apply(ev)` store the event's status, skip event IDs it has
already applied, and remove the service from the cache.

- If `s.processed` already has `ev.ID`, do nothing.
- Otherwise set `s.statuses[ev.Service]` to a `Status` built from the
  event's `Service`, `Version` and `State`, record `ev.ID` in
  `s.processed`, and remove `ev.Service` from `s.cache`.
- Leave every other service's cached entry alone.

```text
Apply evt-1 (api 1.4.2 succeeded)   statuses[api] = 1.4.2 succeeded
Apply evt-2 (api 1.5.0 succeeded)   statuses[api] = 1.5.0 succeeded
Apply evt-1 again (redelivered)     statuses[api] = 1.5.0 succeeded   (skipped)
```

**Think it through:**
- Without the check, what would the redelivered `evt-1` do to `api`'s
  status?
- Why remove the cached entry instead of putting the new status in the
  cache?

### 4. Read through the cache

**Goal:** make `Lookup(service)` answer from the cache when it can, and
from `s.statuses` when it can't.

- Cache hit: return the cached status and `true`. Don't touch
  `s.statuses`.
- Cache miss: add 1 to `s.dbReads`, then read `s.statuses`. If the service
  is there, put it in the cache and return it with `true`. If not, return
  the zero `Status` and `false`, and cache nothing.

```go
// statuses has api; the cache is empty
s.Lookup("api")     // => {api 1.4.2 succeeded}, true   dbReads 1
s.Lookup("api")     // => {api 1.4.2 succeeded}, true   dbReads 1 (cache hit)
s.Lookup("billing") // => {}, false                     dbReads 2
s.Lookup("billing") // => {}, false                     dbReads 3
```

**Think it through:**
- Which `Cache` methods do you need, and in which order?
- Why might caching "not found" cause trouble a minute later?

### 5. Serve it over HTTP

**Goal:** make `handleEvent` and `handleStatus` translate HTTP requests to
`Submit` and `Lookup` calls.

`handleEvent` (`POST /events`):

| Situation | Reply |
|---|---|
| The body isn't a JSON `Event` | `400 Bad Request` |
| `Submit` returns an error wrapping `ErrInvalidEvent` | `400 Bad Request` |
| `Submit` returns `ErrQueueFull` | header `Retry-After: 1`, then `503 Service Unavailable` |
| `Submit` returns `nil` | `202 Accepted` |

`handleStatus` (`GET /status/{service}`): read the service name with
`r.PathValue("service")` and call `Lookup`. If it isn't found, reply
`404 Not Found`. Otherwise set `Content-Type: application/json` and write
the `Status` as JSON with status 200.

```text
POST /events {"id":"e1","service":"api","version":"1.4.2","state":"started"}   -> 202
GET  /status/api       (after the worker ran)  -> 200 {"service":"api","version":"1.4.2","state":"started"}
GET  /status/billing                           -> 404
POST /events {"state":"done", ...}                                             -> 400
```

**Think it through:**
- `202 Accepted` means "queued, not done yet". Why is that the honest
  answer here?
- Headers must be set before the status code is written. Where does
  `Retry-After` go?

## Learn more

- [Google SRE book: Handling Overload](https://sre.google/sre-book/handling-overload/)
- [Azure Architecture Center: Cache-Aside pattern](https://learn.microsoft.com/en-us/azure/architecture/patterns/cache-aside)
- [MDN: 202 Accepted](https://developer.mozilla.org/en-US/docs/Web/HTTP/Reference/Status/202)
- [Design Docs at Google](https://www.industrialempathy.com/posts/design-docs-at-google/)
