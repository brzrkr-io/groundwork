# Goroutines: fix the maintenance script

A teammate's maintenance script uses goroutines, but it is still slow, and it
can't report what its goroutines found. Time to fix it.

**New here:** collecting goroutine results in a slice; mostly practice.

## Learn

### Recap

`go f()` starts `f` in a new goroutine and carries on without waiting. A
`sync.WaitGroup` waits for goroutines: start each with `wg.Go(func() { ... })`,
then call `wg.Wait()` **once, after** you have started them all.

> **Why is it written like this?** Where `wg.Wait()` sits is your choice, and
> it decides the behaviour. Inside the loop, each goroutine must finish before
> the next one starts, so nothing overlaps.

### Getting results back

`go` throws away a function's return value. A simple way to collect results:
make a slice **before** starting the goroutines, with one slot per input.
Goroutine `i` writes only to slot `i`:

```go
sizes := make([]int, len(disks))
var wg sync.WaitGroup
for i, disk := range disks {
	wg.Go(func() {
		sizes[i] = usage(disk) // this goroutine owns slot i
	})
}
wg.Wait()
fmt.Println(sizes) // same order as disks
```

```text
disks:  [ "sda" , "sdb" , "sdc" ]
            │       │       │
         goroutine goroutine goroutine
            ▼       ▼       ▼
sizes:  [  40   ,  91   ,  12   ]
```

This is safe because no two goroutines touch the same slot, and nobody reads
the slice until `wg.Wait()` returns. Reading it earlier is a bug: some slots
might not be written yet.

> **Why is it written like this?** Go requires the slice to exist before the
> goroutines write to it, so `make` comes first. Using the loop index `i` as
> the slot number is the author's choice; it keeps results in input order.

## Your task

The script drains Kubernetes nodes before a kernel upgrade, then pings the
database hosts. Every drain or ping takes about 1 second.

### 1. Fix the slow drain

**Goal:** make `DrainNodes(nodes, drain)` run all drains at the same time and
return once all of them have finished.

It already uses a WaitGroup, but three nodes take 3 seconds instead of 1.
Find the bug; the `TODO(1)` comment sits right above it.

```go
DrainNodes([]string{"node-a", "node-b", "node-c"}, drain)
// should take ~1s, not ~3s
```

**Think it through:**
- Which line makes the loop stop and wait for each drain?
- Does that line need to change, or only move? Where should it be?

### 2. Ping hosts concurrently

**Goal:** make `PingAll(hosts, ping)` ping every host at the same time and
return one `bool` per host, in the same order as `hosts`.

Store each answer in its own slot of the existing `results` slice, and return
only after every ping has finished. Three hosts should take ~1s, not ~3s.

```go
PingAll([]string{"db-1", "db-2", "db-3"}, ping)
// => []bool{true, false, true}   (after ~1s)
```

**Think it through:**
- The line that writes `results[i]` stays the same. What do you wrap it in?
- What do you need to declare before the loop, and call after it?

## Learn more

- [Go by Example: WaitGroups](https://gobyexample.com/waitgroups)
- [sync.WaitGroup.Go](https://pkg.go.dev/sync#WaitGroup.Go)
- [Go blog: Fixing for loops in Go 1.22](https://go.dev/blog/loopvar-preview)
