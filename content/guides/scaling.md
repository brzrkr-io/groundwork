# Scaling: caches, queues and load balancers

**In one sentence:** scaling is making a system handle more load by adding resources, and the tools for it (load balancers, caches, queues, replicas, shards and CDNs) each buy capacity at the price of some new problem.

**Why it matters:** "it's slow" and "it fell over at peak" are the most common incident tickets a platform engineer sees. The fix is rarely "buy a bigger server". Knowing which tool takes load off which part, and what each one breaks in return, is how you choose the right one at 3 a.m.

**You'll learn:**

- Vertical versus horizontal scaling, and why stateless services scale out easily
- How load balancers spread requests, and when each algorithm fits
- Where caches sit, how they go stale, and how to stop a stampede
- How queues absorb bursts, push back when full, and deliver at least once
- How databases scale with read replicas and sharding, and what CDNs do
- What autoscaling can and can't fix

## Find the bottleneck first

A system is as fast as its slowest part. If the database is at 100% CPU, adding web servers makes things **worse**: more servers send more queries to the same struggling database.

So before scaling anything, find the *bottleneck*: the resource that runs out first. Look at your metrics during the slow period:

| Symptom | Likely bottleneck |
|---|---|
| App CPU near 100% | Application servers |
| App CPU low, requests slow, many waiting on the DB | Database |
| Connection errors, "too many connections" | Connection limits (DB, file descriptors) |
| Memory climbing until OOM kill | Unbounded queue or cache, a leak |
| Fine in one region, slow in another | Network distance |

The *System design fundamentals* guide shows how to estimate load before it arrives. This guide is about what to do once you know where it hurts.

## Vertical and horizontal scaling

There are two ways to add capacity.

- *Vertical scaling* (scaling **up**): a bigger machine. 4 CPUs become 32.
- *Horizontal scaling* (scaling **out**): more machines. 2 instances become 10.

| | Vertical | Horizontal |
|---|---|---|
| Effort | Change one instance type | Needs a load balancer, stateless code |
| Ceiling | The biggest machine you can buy | Very high |
| Failure | Still one machine: a single point of failure | Lose one of ten, lose 10% |
| Downtime to grow | Often a restart | None: add instances |
| Good for | Databases, quick fixes | Web services, workers |

Vertical scaling is underrated: it's one setting and no code change. Use it to buy time. But it has a ceiling and keeps the single point of failure, so services that must grow scale out.

## Stateless services

Horizontal scaling only works if **any instance can serve any request**. That's what *stateless* means: the instance keeps nothing between requests that another instance would need.

State hides in surprising places:

| Hidden state | Problem when scaled out | Move it to |
|---|---|---|
| Login sessions in memory | The next request hits another instance and you're logged out | A shared store (Redis) or signed cookies |
| Uploaded files on local disk | The other instances can't see them | Object storage (S3) |
| A counter in a Go variable | Each instance counts separately | A database or a metrics system |
| A cron job inside the app | Runs once per instance: ten times | A separate scheduled job |

The state doesn't disappear; it moves to a service built to share it. Then app instances become interchangeable *cattle*: Kubernetes can kill, move and add them freely. This is why a Kubernetes Deployment's pods are meant to be stateless, and databases run differently.

## Load balancing

A *load balancer* sits in front of identical instances and spreads requests across them. You've used them: an nginx `upstream` block, a Kubernetes Service, an AWS Application Load Balancer.

```text
                 +--> api-1 (healthy)
clients --> LB --+--> api-2 (healthy)
                 +--> api-3 (failed health check: skipped)
```

It does two jobs:

1. **Spread the load**, so no instance is overwhelmed while others idle.
2. **Route around failure**: it runs *health checks* (say `GET /healthz` every 5 s) and stops sending traffic to instances that fail them. This is what makes N+1 redundancy work.

Load balancers work at two levels. A *layer 4* balancer forwards TCP connections without reading them: fast, but it can't route by URL. A *layer 7* balancer understands HTTP: it can route `/api` and `/static` to different services, add headers and retry.

### Algorithms

| Algorithm | How it picks | Good when | Weak when |
|---|---|---|---|
| Round robin | The next one in order: 1, 2, 3, 1, 2, 3 | Requests cost about the same | Some requests are slow |
| Weighted round robin | Round robin, bigger instances get more turns | Mixed instance sizes | Same as round robin |
| Least connections | The one with the fewest requests in progress | Request times vary a lot | Needs live counts |
| Random of two | Pick two at random, use the less busy | Many balancers, no shared state | Rarely a problem |
| Consistent hashing | The same key always goes to the same instance | Caches, sticky data | Uneven keys make hot spots |

**Why least connections helps:** with round robin, an instance stuck on three 30-second report exports still gets every third request while the others sit idle. Least connections notices the pile-up and sends new work elsewhere. You'll build round robin with health checks in *Load balancing: round robin with health checks* and fix a least-connections balancer in *Load balancing: fix the least-connections balancer*.

### Consistent hashing

Sometimes you **want** the same key to land on the same server, so that server's cache already holds it. The naive way is `hash(key) % N`. But change N from 4 to 5 and about 80% of keys move to a different server: every cache goes cold at once.

*Consistent hashing* places servers and keys on a circle of hash values. A key belongs to the first server clockwise from it:

```text
                 0
         cache-3#0 . cache-1#0
       .                     .   <- key "api" hashes here
  cache-2#1                cache-1#1   <- first server clockwise: "api" goes to cache-1
       .                     .
         cache-2#0 . cache-3#1

(each server sits at several points, "#0", "#1": its virtual nodes)
```

Add a server and it takes over only the keys between it and its neighbour: about 1/N of them. Placing each server at many points (*virtual nodes*) evens out the spread. You'll see the modulo problem in *Sharding: route keys and plan a rebalance* and build the ring in *Consistent hashing: a ring with virtual nodes*.

## Caching

A *cache* keeps a copy of an answer closer to whoever asks, so the slow source isn't asked every time. The *Distributed systems basics* guide covers staleness; here is where caches sit and how to run them.

### Layers

A request can meet a cache at every hop:

```text
browser cache -> CDN -> load balancer -> app (in-process cache) -> shared cache (Redis) -> database (its own buffer cache)
```

| Layer | Speed | Shared by | Example |
|---|---|---|---|
| In-process | Nanoseconds | One instance only | A Go map with a TTL, an LRU |
| Shared (Redis, Memcached) | About 1 ms | Every instance | Sessions, rendered pages |
| CDN | Near the user | Every user in a region | Images, JS, release downloads |

An in-process cache is fastest but each replica has its own copy, so ten replicas may hold ten different versions for a while. A shared cache is one copy, one network hop away.

### Cache-aside

The most common pattern is *cache-aside* (also called lazy loading). The app manages the cache itself:

```text
read:   look in cache -> hit? return it
                      -> miss? read the database, store it in cache, return it
write:  write the database -> delete the cache entry
```

Why **delete** on write instead of updating the cache? Two writers racing can leave the older value in the cache forever if both update it; deleting means the next read fetches the truth. Other patterns exist: *write-through* (write cache and database together) and *write-behind* (write the cache, flush to the database later, risking loss). You'll fix a broken cache-aside in *LRU caches: fix the stale cache-aside reads*.

### TTLs and invalidation

There are two ways a cached copy stops being used:

- **TTL** (time to live): the entry expires after a fixed time. Simple, and sets an upper bound on staleness. 30 seconds is fine for an instance list; 0 is right for a bank balance.
- **Invalidation**: whoever changes the data deletes the cached copy. Fresh, but every writer must remember to do it, and a missed path means stale data with no expiry.

Real systems use both: invalidate on write, plus a TTL as a safety net for the path someone forgot.

### Eviction

Memory is limited, so a full cache must throw something out. *Eviction policies*:

- **LRU** (least recently used): evict whatever hasn't been read for longest. The usual default.
- **LFU** (least frequently used): evict whatever is read least often.
- **TTL-first**: evict whatever expires soonest.

An unbounded cache is a memory leak with good intentions. You'll build an LRU in *LRU caches: evict the least recently used*.

### Hit ratio

The *hit ratio* is hits divided by all lookups. It's the number that tells you whether a cache is earning its keep:

```text
1,000 reads per second, 95% hit ratio  -> 50 reach the database
1,000 reads per second, 80% hit ratio  -> 200 reach the database (4× the load)
```

A small drop in hit ratio can multiply database load, which is why a cache restart after a deploy can cause an outage.

### Stampedes

A popular entry expires, and in the next millisecond 500 requests miss together and all hit the database. That's a *cache stampede*. Fixes:

- **Request collapsing**: one request loads, the rest wait for its answer (*TTL caching: collapse concurrent loads*).
- **Jittered TTLs**: 60 s ± 10% so entries don't all expire in the same second.
- **Refresh ahead**: reload a hot entry shortly before it expires.

## Queues and async work

Some work doesn't need to happen while the user waits: building an image, sending a notification, resizing a log archive. Put it on a *queue* and let *workers* do it at their own pace.

```text
API (producer) --> [ queue: job job job ... ] --> worker 1
                                              --> worker 2   (consumers)
```

What a queue buys you:

- **Fast responses**: the API replies `202 Accepted` at once.
- **Burst absorption**: 1,000 jobs arrive in a second; three workers chew through them over the next minute.
- **Decoupling**: the API keeps accepting work while workers restart or deploy.

### Backpressure

A queue with no limit only hides overload: it grows until memory runs out, and every job waits longer. A *bounded queue* has a fixed size, and when it's full it *pushes back*:

| When full | Effect |
|---|---|
| Reject at once (`503` + `Retry-After`) | Callers slow down; nothing is silently delayed for an hour |
| Block the producer for a while | Natural slow-down, but the producer can hang without a timeout |
| Drop the oldest jobs | Fine for metrics, wrong for deploys |

That is *backpressure*: overload becomes a visible signal at the edge instead of invisible delay inside. You'll build one in *Work queues: bounded queues and backpressure*.

A rule of thumb for sizing (*Little's law*): the average number of jobs in the system equals arrival rate × time each spends there. 50 jobs per second, each taking 2 seconds, means about 100 jobs in flight, so you need about 100 workers' worth of capacity or the queue grows without end.

### At-least-once and idempotent consumers

Queues like SQS, RabbitMQ and Kafka deliver each message **at least once**. A worker takes a message, does the work, then *acknowledges* it. If the worker crashes before the ack, the queue hands the message to someone else, and it may run twice.

So consumers must be *idempotent*: processing a message twice has the same effect as once. The usual way is to record the IDs of processed messages and skip repeats. Two ordering rules matter:

1. **Ack after the work succeeds**, never before, or a crash loses the job.
2. **Record the ID after the work succeeds**, or a failed job is skipped as a "duplicate" on retry.

Messages that fail again and again go to a *dead-letter queue* for a human to look at, instead of blocking the queue forever. You'll fix a consumer that breaks both rules in *Work queues: fix the at-least-once consumer*.

## Scaling the database

App servers are easy to scale because they're stateless. The database holds the state, so it's usually the last bottleneck.

### Read replicas

Most services read far more than they write. A *read replica* is a copy of the database that follows the *primary* and serves reads:

```text
writes --> primary --(replication)--> replica 1 --> reads
                                  --> replica 2 --> reads
```

Replicas multiply read capacity and give you a standby for failover. The catch is *replication lag*: a replica can be a few milliseconds to seconds behind. A user who saves a setting and reloads from a replica may see the old value. Send "read your own write" queries to the primary.

### Sharding

When **writes** or the data size outgrow one machine, split the data: each *shard* is a separate database holding part of the rows. A *shard key* decides where each row lives:

```text
shard = hash(service name) % 4

"api"     -> shard 3
"billing" -> shard 0
"web"     -> shard 1
"worker"  -> shard 3
```

Choosing the shard key is the big decision:

- **Spread**: a key with many values (customer ID) spreads well; a key with few (country) makes *hot shards*.
- **Locality**: queries that need one key's data hit one shard. Queries across all shards (reports, joins) get slow and complicated.
- **Rebalancing**: adding a shard means moving data. With `% N`, most keys move; consistent hashing moves about 1/N.

Shard last. It adds permanent complexity to every query, backup and migration. Bigger machines, replicas, caching and archiving old data come first. You'll route keys and count what moves in *Sharding: route keys and plan a rebalance*.

## CDNs

A *content delivery network* (CDN) is a cache spread across hundreds of locations near users. A user in Sydney gets your release binary from a server in Sydney instead of crossing the Pacific to `us-east-1`.

- **Static content** (images, JS bundles, release downloads, container image layers) is the classic use: it rarely changes, and versioned file names (`app.3f9c1.js`) make invalidation unnecessary.
- **Origin protection**: the CDN absorbs most traffic, so the *origin* (your servers) only sees cache misses.
- **Purging**: when content must change now, you ask the CDN to *purge* it, which takes seconds to minutes to reach every location.

`Cache-Control` headers from your service tell the CDN what it may cache and for how long, the same TTL idea one layer out.

## Autoscaling and its limits

*Autoscaling* adds and removes instances automatically from a metric: Kubernetes' Horizontal Pod Autoscaler (HPA) scales pods on CPU or custom metrics, and AWS Auto Scaling groups do the same for VMs.

```text
target: 60% CPU
load rises -> CPU 90% -> HPA adds pods -> CPU back to 60%
load falls -> CPU 20% -> HPA removes pods (slowly, to avoid flapping)
```

It's great for daily traffic curves. Its limits:

- **It's slow.** Noticing the metric, starting a VM or pulling an image and passing readiness takes from 30 seconds to several minutes. A spike that arrives in 5 seconds is over (or has already taken you down) before help arrives. Keep headroom, and keep queues for bursts.
- **It scales the wrong thing.** If the database is the bottleneck, more app pods just open more connections to it. Each pod's connection pool × number of pods can exceed the database's connection limit.
- **It needs a good metric.** CPU is a poor signal for I/O-bound services; queue length or requests per pod are often better.
- **It needs limits.** A `maxReplicas` stops a bug or an attack from scaling you into a huge bill; a `minReplicas` keeps N+1 at night.
- **It needs stateless, fast-starting services.** A service that takes 3 minutes to warm its cache scales badly.

Autoscaling handles predictable growth. Bursts are handled by queues, headroom and rate limits.

## Common mistakes

- **Scaling the wrong tier.** Adding app servers when the database is the bottleneck makes things worse. Find the bottleneck first.
- **Hidden state in "stateless" services.** In-memory sessions, local files and in-process cron jobs break as soon as there are two instances.
- **Caches with no bound and no TTL.** That's a memory leak plus stale data with no expiry.
- **Updating the cache on write.** Racing writers can leave an old value cached forever; delete the entry instead.
- **Unbounded queues.** They hide overload until the process runs out of memory.
- **Acknowledging before the work is done.** A crash between ack and work loses the job silently.
- **Sharding too early.** It's permanent complexity; replicas, caching and bigger machines come first.
- **Trusting autoscaling with spikes.** It reacts in minutes, not seconds.

## Check yourself

1. A web service's pods sit at 15% CPU, but requests are slow and the database is at 95% CPU. Will adding more web pods help? Why or why not?
2. Why does a service that stores login sessions in a Go map break when you scale it to three replicas behind a load balancer?
3. A load balancer uses round robin. One of three backends is busy with several slow 30-second report exports. What goes wrong, and which algorithm handles it better?
4. A cluster routes cache keys with `hash(key) % 4`. Roughly what share of keys change server when a fifth server is added, and what technique reduces that?
5. In the cache-aside pattern, what should a write do to the cache, and why not simply update the cached value?
6. A worker acknowledges a queue message as soon as it receives it, then crashes halfway through the job. What happens to the job?
7. A deploy event queue receives 20 events per second, and each takes 0.5 seconds to process. On average, how many events are in flight, and what does that tell you about the number of workers?
8. Why can't Kubernetes' Horizontal Pod Autoscaler protect a service from a traffic spike that arrives within five seconds?

## Answers

1. No. The database is the bottleneck. More web pods send more queries and open more connections to it, which makes it slower. Reduce database load (caching, better queries, read replicas) or scale the database.
2. Each replica has its own map. A user who logs in on replica 1 and whose next request lands on replica 2 appears logged out. Move the sessions to a shared store or a signed cookie so every replica sees them.
3. Round robin keeps sending every third request to the busy backend, so those requests queue behind the slow exports while the other backends sit idle. Least connections sends new requests to the backend with the fewest in progress.
4. About 80% (4 out of 5 keys), so nearly every cache goes cold at once. Consistent hashing moves only about 1/5 of the keys, to the new server.
5. Write the database first, then delete the cache entry. If two writers both update the cache, the slower one can leave an older value there forever; deleting means the next read fetches the current value from the database.
6. The job is lost: the queue thinks it was done and never redelivers it. Acknowledge only after the work succeeds, so a crash leads to redelivery instead.
7. By Little's law, 20 × 0.5 = 10 events in flight on average. You need at least 10 workers' worth of capacity (plus headroom) or the queue grows forever.
8. Autoscaling first has to notice the metric, then start new pods, pull images and pass readiness checks, which takes from half a minute to several minutes. The spike has hit (and maybe passed) by then. Headroom, queues and rate limits handle bursts.

## Key terms

| Term | Meaning |
|---|---|
| Bottleneck | The resource that runs out first and limits the whole system |
| Vertical scaling | Handling more load with a bigger machine |
| Horizontal scaling | Handling more load with more machines |
| Stateless service | A service where any instance can handle any request |
| Load balancer | A component that spreads requests across identical instances |
| Health check | A periodic probe a load balancer uses to skip broken instances |
| Round robin | Picking backends in turn, one after another |
| Least connections | Picking the backend with the fewest requests in progress |
| Consistent hashing | Mapping keys to servers on a ring so adding one moves only about 1/N of keys |
| Virtual nodes | Placing each server at many points on a hash ring to even out the spread |
| Cache-aside | The app reads the cache, falls back to the database, and deletes cache entries on write |
| Cache invalidation | Removing a cached copy because the source changed |
| Eviction | Throwing out a cache entry to make room, for example the least recently used |
| Hit ratio | The share of cache lookups that find their answer in the cache |
| Cache stampede | Many requests reloading the same expired entry at once |
| Bounded queue | A queue with a fixed size that refuses or delays work when full |
| Backpressure | A busy component pushing back on new work instead of piling it up |
| Little's law | Items in a system = arrival rate × time each spends there |
| Dead-letter queue | Where messages that keep failing are parked for a human |
| Read replica | A copy of a database that follows the primary and serves reads |
| Replication lag | How far a replica is behind the primary |
| Sharding | Splitting data across several databases by a shard key |
| Hot shard | A shard that gets far more load than the others |
| CDN | A network of caches near users that serves content for your origin |
| Autoscaling | Adding and removing instances automatically based on a metric |

## Watch

- [CS75 (Summer 2012) Lecture 9 Scalability Harvard Web Development David Malan](https://www.youtube.com/watch?v=-W9F__D3oY4) · George Schott · 106 min: a classic Harvard lecture that walks from one server to vertical and horizontal scaling, load balancers, sticky sessions, caching and database replication; long, so watch it in parts: every idea in this guide is in it.
- [What is LOAD BALANCING? ⚖️](https://www.youtube.com/watch?v=K0Ta65OqQkY) · Gaurav Sen · 14 min: why spreading load needs more than `% N`, building up to consistent hashing.
- [Consistent Hashing | Algorithms You Should Know #1](https://www.youtube.com/watch?v=UF9Iqmg94tk) · ByteByteGo · 8 min: the hash ring and virtual nodes, animated.
- [Cache Systems Every Developer Should Know](https://www.youtube.com/watch?v=dGAgxozNWFE) · ByteByteGo · 6 min: every layer a request can be cached at, from the browser to the database.
- [What is a MESSAGE QUEUE and Where is it used?](https://www.youtube.com/watch?v=oUJbuFMyBDk) · Gaurav Sen · 10 min: why queues decouple producers from workers, and what happens when a worker dies mid-job.

## Go deeper

- [Horizontal Pod Autoscaling](https://kubernetes.io/docs/concepts/workloads/autoscaling/horizontal-pod-autoscale/) (Kubernetes docs): how the HPA picks a replica count, and its stabilization window against flapping.
- [Cache-Aside pattern](https://learn.microsoft.com/en-us/azure/architecture/patterns/cache-aside) (Azure Architecture Center): the read and write paths, with the consistency issues spelled out.
- [Load Balancing in the Datacenter](https://sre.google/sre-book/load-balancing-datacenter/) (Google SRE book): how Google spreads load across backends, including why plain round robin falls short.
- [Consistent hashing](https://en.wikipedia.org/wiki/Consistent_hashing) (Wikipedia): the original idea, its history at Akamai, and where it's used today.
- [Little's law](https://en.wikipedia.org/wiki/Little%27s_law) (Wikipedia): the one formula behind sizing queues and worker pools.
