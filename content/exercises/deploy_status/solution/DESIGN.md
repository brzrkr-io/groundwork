# Design: deploy-status service

## Requirements

Functional: CI pipelines send deploy events (service, version, state).
Anyone can ask for the latest deploy of one service. The states are
`started`, `succeeded` and `failed`, and anything else is rejected.

Non-functional: accepting an event must be fast (under 50 ms) even during a
deploy rush, a redelivered event must never change the answer, and a status
may be up to a second or two out of date. Losing the cache must not lose
data.

## Estimates

- 300 services, about 10 deploys each per day, 2 events per deploy:
  6,000 events a day, or 0.07 per second on average.
- Peak: a monorepo merge deploys everything at once, so 600 events arrive
  within a few seconds. A queue of 1,000 events absorbs that.
- Reads: 50 open dashboards, each polling 300 services every 10 seconds:
  1,500 reads per second, far more than writes. Reads need the cache.
- Storage: 300 statuses of about 200 bytes is 60 KB. Remembering event IDs
  (about 40 bytes each) for 30 days is 6,000 × 30 × 40 bytes, about 7 MB.

## Design

`POST /events` validates the event and puts it on a bounded queue, then
replies `202 Accepted` straight away. When the queue is full it replies
`503` with `Retry-After: 1`, so CI retries later instead of piling up work.

One worker takes events off the queue and applies them to the database.
It skips event IDs it has already applied, because delivery is
at-least-once, and it removes the service from the cache after each write.

`GET /status/{service}` reads through an LRU cache: a hit never touches the
database, and a miss reads the database and fills the cache.

## Trade-offs

- Queue instead of writing in the request: CI gets a fast answer and bursts
  are smoothed out, but a status can lag an event by a moment, and events
  still in the queue are lost if the process crashes. A durable queue (SQS)
  would fix that at the cost of another dependency.
- Invalidate instead of updating the cache on write: the next read pays one
  database read, but the cache can never hold a value the database doesn't.
- One worker instead of many: events for a service are applied in order and
  the code stays simple, which is plenty for 0.07 events per second.
- The cache is per process: two replicas could disagree for a moment. That
  is fine for a dashboard; a shared cache (Redis) is the next step if not.
