## Task 1: Write the problem

Start from what people saw (incidents, evicted pods, manual clean-ups), then explain the cause with the numbers: how much log data, how big the disk, when eviction starts.

### Closer

Do the maths for the reader: 6 GB a day for 14 days is 84 GB on a 100 GB disk, and eviction starts at 85 GB. Then the impact: 3 incidents, 2 hours, checkout down. Mention the second pain too: old logs are stuck on single nodes.

### Answer

```markdown
## Problem

Our 40 Kubernetes nodes each have a 100 GB disk. Every node writes about
6 GB of application logs a day, and logrotate keeps 14 days of them, so
logs alone can take 84 GB. Once a disk passes 85% full, the kubelet starts
evicting pods. That caused 3 incidents last month; the longest took 2 hours
to recover. Nobody can search logs older than a few days without logging
in to the node that wrote them.
```

## Task 2: List the constraints

Look for the facts that say what you may **not** do: the audit rule, the budget. Then add what must stay the same for other teams, and how it must roll out.

### Closer

Turn each fact into a checkable rule: "Audit logs must be kept for 90 days and stay searchable." "Node costs may grow at most 10%." One per `- ` line.

### Answer

```markdown
## Constraints

- Audit logs must be kept for 90 days and stay searchable.
- Node costs must not double; the budget allows about 10% more.
- Applications must not change how they write logs.
- The change must roll out one node pool at a time.
```

## Task 3: Compare the options

Three natural options: keep fewer days on the node, ship logs off the node, or buy bigger disks. Weigh each against your constraints.

### Closer

Use the shape from the README for each: `### Option A: ...`, a description sentence, then `- Pro:` and `- Con:` lines. Shorter local retention breaks the 90-day rule; bigger disks cost five times as much; shipping logs costs a little and adds a moving part.

### Answer

```markdown
### Option A: Keep fewer days on the node

Change logrotate to keep 3 days of logs instead of 14, so at most 18 GB of
each disk holds logs.

- Pro: One config change, live in a day, costs nothing.
- Con: Logs older than 3 days are gone, which breaks the 90-day audit rule.

### Option B: Ship every log to object storage

Turn on Fluent Bit for all logs instead of only errors, send them to the
S3-compatible bucket with a 90-day lifecycle rule, and keep 2 days locally.

- Pro: Meets the audit rule, and every log becomes searchable in one place.
- Con: Storage and network cost about 6% more, and the shipper is one more
  thing to monitor.
```

## Task 4: Make the decision

Pick the option that meets every constraint, name it with its letter, give the reason with "because", then describe how the change rolls out and how it rolls back.

### Closer

"We choose Option B because ..." followed by the constraint it alone meets. For the rollback: which config files change, and what does reverting them do?

### Answer

```markdown
## Decision

We choose Option B because it is the only option that meets the 90-day
audit constraint within budget, and it uses a shipper we already run. We
roll it out one node pool at a time, starting with staging, and watch disk
use and shipper errors for a week before the next pool. To roll back, we
revert the Fluent Bit and logrotate config to the current versions.
```
