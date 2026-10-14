# Work queues: fix the at-least-once consumer

Your deploy service reads deploy requests from a message queue. This week a
failed deploy of `api` simply vanished. Fixing that makes the queue deliver
some requests twice, so a slow `billing` deploy would run twice. Fix the
consumer so every request is deployed exactly once in effect, even when the
queue delivers it more than once.

**New here:** message brokers, ack and nack, visibility timeouts, at-least-once
delivery, idempotent consumers.

## Learn

### Recap

As in *Work queues: bounded queues and backpressure*, a queue sits between
the code that accepts work and the code that does it. There the queue was a
channel inside one process: if the process dies, the jobs die with it. Real
systems use a *message broker* (Amazon SQS, RabbitMQ, Kafka): a separate
service that keeps messages safe until a consumer says it is done with them.

### Ack and nack

A consumer receives a message, does the work, then tells the broker how it
went:

| Call | Means | The broker |
|---|---|---|
| `Receive()` | "Give me the next message" | Hands it over and marks it *in flight* |
| `Ack(id)` | "Done, forget it" | Deletes it for good |
| `Nack(id)` | "Failed, try again later" | Puts it back at the end of the queue |

RabbitMQ has these calls under the same names. SQS calls an ack
`DeleteMessage`, and has no nack: you set the message's visibility timeout
(next section) to 0, so it's delivered again straight away.

### Visibility timeout

What if the consumer crashes before it acks or nacks? In SQS, a received
message is hidden for a *visibility timeout* (30 seconds by default). If
nobody deletes it in time, it becomes visible again and is delivered again.
So a crashed consumer never loses work. But a consumer that is just *slow*
gets the same message twice: once now, and again when the timeout runs out.

### Delivery guarantees

When you ack decides what can go wrong:

| Order | Guarantee | Failure or crash means |
|---|---|---|
| Ack, then work | *At-most-once* | The work is lost |
| Work, then ack | *At-least-once* | The work may run twice |
| Work, then ack, and skip IDs you've done | *Effectively-once* | Nothing: duplicates are ignored |

Losing a deploy request is worse than seeing it twice, so deploy queues are
at-least-once. The *Distributed systems basics* guide explains why true
exactly-once delivery is impossible over a network.

### Idempotent consumers

Every message carries an `ID` that stays the same on every redelivery. The
consumer remembers the IDs it has finished in a set, and when a known ID
comes back, it acks it and skips the work. That makes the consumer
*idempotent*: handling a message twice has the same effect as once. It's the
same idea as *Idempotency keys: replay retried requests*, over a queue
instead of HTTP.

A `map[string]bool` works as a set. A missing key reads as `false`:

```go
seen := map[string]bool{}
seen["alert-17"] = true
seen["alert-17"] // => true
seen["alert-99"] // => false
```

> **Why is it written like this?** Go has no built-in set type, so
> `map[K]bool` is the common convention. Reading a missing key returns the
> zero value, `false`, which Go guarantees, so no `ok` check is needed.

**When** you record an ID matters as much as **whether** you do. Record it
only once the work has really succeeded, and before the ack: if the consumer
crashes between the two, the redelivery is skipped instead of run again.

### In production

This pack keeps the IDs in memory, so a restart forgets them. A real
consumer stores them durably, for example as a row in its database with the
message ID as a unique key, written in the same transaction as the work. It
deletes old IDs after a retention window longer than the broker could
possibly redeliver (SQS keeps a message for at most 14 days).

### The broker in this pack

`broker.go` is a complete, tiny SQS-like broker. It runs in one goroutine
and has no real clock. The tests use two extra methods:

| Method | Does |
|---|---|
| `Expire()` | Pretends the visibility timeout ran out: every in-flight message goes back in the queue |
| `Pending()` | How many messages are not acked yet (ready plus in flight) |

`Ack` and `Nack` do nothing for a message that isn't in flight.

### How to find bugs like these

Walk one message through `HandleNext`, line by line. After each line, ask:

- What does the broker hold now: ready, in flight, or gone?
- What does `c.done` say about this ID?
- If the deploy failed here, or the process crashed here, would the request
  still get deployed exactly once?

The first line where the answer is "no" is the bug.

## Your task

`HandleNext` in `exercise.go` receives one message and deploys it. It already
returns `false, nil` when no message is ready, and `true` plus the deploy's
error otherwise; keep that. It has three bugs, one per task.

### 1. Ack only after the deploy succeeds

**Goal:** make `HandleNext` ack a message only after its deploy succeeds,
and nack it when the deploy fails.

Right now it acks before deploying, so a failed deploy is gone for good:

```text
publish d1 (api v2); the deploy fails
HandleNext() => true, "rollout of api@v2 timed out"
Pending()    => 0   (buggy: d1 is lost)
Pending()    => 1   (fixed: d1 was nacked and will be retried)
```

On success, ack `m.ID`. On failure, nack `m.ID` and return `true` and the
error.

**Think it through:**
- Which of the three rows in the delivery guarantees table is the code now?
  Which one do you want?
- The ack moves to after the `if err := ...` block. What goes inside that
  block?

### 2. Skip messages that were already deployed

**Goal:** make `HandleNext` ack a message whose ID is already in `c.done`
and return `true, nil` without deploying it.

Now that the ack waits for the deploy, a slow deploy gets redelivered. The
consumer fills `c.done` but never reads it, so the redelivered message
deploys again:

```text
publish d1 (billing v14); the deploy is slower than the visibility timeout
HandleNext() => deploys billing@v14; d1 is redelivered meanwhile
HandleNext() => deploys billing@v14 again   (buggy)
HandleNext() => skips d1 and acks it         (fixed)
```

Only the `ID` decides what's a duplicate. Two messages with different IDs
but the same service and version are two requests, and both deploy.

**Think it through:**
- The check goes before the deploy. What do you look up in `c.done`?
- A skipped duplicate still has to be acked. Why, and what happens if it
  isn't?

### 3. Record a deploy only after it succeeds

**Goal:** make `HandleNext` add `m.ID` to `c.done` only after the deploy
succeeded, before the ack.

After task 2, a failed deploy is retried, but the retry is skipped as a
"duplicate": the ID was recorded before the deploy even ran.

```text
publish d1 (api v2); the first deploy fails
HandleNext() => fails, d1 nacked
HandleNext() => skips d1 as a duplicate, api v2 never deployed   (buggy)
HandleNext() => deploys api@v2                                   (fixed)
```

**Think it through:**
- Which line marks `m.ID` as done, and which point in `HandleNext` is the
  first one where the deploy is known to have worked?

## Learn more

- [Amazon SQS visibility timeout](https://docs.aws.amazon.com/AWSSimpleQueueService/latest/SQSDeveloperGuide/sqs-visibility-timeout.html)
- [Amazon SQS at-least-once delivery](https://docs.aws.amazon.com/AWSSimpleQueueService/latest/SQSDeveloperGuide/standard-queues-at-least-once-delivery.html)
- [RabbitMQ: Consumer acknowledgements and publisher confirms](https://www.rabbitmq.com/docs/confirms)
- [Wikipedia: Idempotence](https://en.wikipedia.org/wiki/Idempotence)
