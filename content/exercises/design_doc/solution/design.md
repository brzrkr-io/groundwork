# Design: stop log files from filling node disks

Author: you · Status: draft · Reviewers: platform team

## Problem

Our 40 Kubernetes nodes each have a 100 GB disk. Every node writes about
6 GB of application logs a day, and logrotate keeps 14 days of them, so
logs alone can take 84 GB. Once a disk passes 85% full, the kubelet starts
evicting pods. That caused 3 incidents last month; the longest took 2 hours
to recover, and checkout was down for 25 minutes of it. On-call engineers
now clean disks by hand, and nobody can search logs older than a few days
without logging in to the node that wrote them.

## Constraints

- Audit logs must be kept for 90 days and stay searchable.
- Node costs may grow by at most 10%.
- Applications must not change how they write logs.
- The change must roll out one node pool at a time.

## Options

### Option A: Keep fewer days on the node

Change logrotate to keep 3 days of logs instead of 14, so at most 18 GB of
each disk holds logs.

- Pro: One config change, live in a day, costs nothing.
- Con: Logs older than 3 days are gone, which breaks the 90-day audit rule.

### Option B: Ship every log to object storage

Turn on Fluent Bit, which already runs on every node, for all logs instead
of only errors. It sends them to our S3-compatible bucket, where a
lifecycle rule deletes them after 90 days. Nodes keep 2 days locally.

- Pro: Meets the audit rule, and every log becomes searchable in one place.
- Pro: Disk use for logs drops to about 12 GB per node.
- Con: Storage and network cost about 6% more, and the shipper is one more
  thing to monitor.

### Option C: Bigger disks

Replace the 100 GB disks with 500 GB disks and leave logging as it is.

- Pro: No change to logging, so no risk of losing logs.
- Con: Disk cost goes up five times, and logs are still stuck on each node.

## Decision

We choose Option B because it is the only option that meets the 90-day
audit constraint within budget, and it uses a shipper we already run. We
roll it out one node pool at a time, starting with staging, and watch disk
use and shipper errors for a week before the next pool. To roll back, we
revert the Fluent Bit and logrotate config to the current versions; no
application changes, so nothing else needs undoing.
