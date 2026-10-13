# Goroutines: fix the maintenance script

A teammate's maintenance script uses goroutines, but it is still slow, and it
can't report what its goroutines found. Time to fix it.

## Learn

### Recap

`go f()` starts `f` in a new goroutine and carries on without waiting. A
`sync.WaitGroup` waits for goroutines: start each one with `wg.Go(func() {
... })`, then call `wg.Wait()` once, after you have started them all.

### Getting results back

`go` throws away a function's return value, so `go ping(host)` gives you
nothing back. A simple way to collect results is a slice made **before**
starting the goroutines, with one slot per input. Goroutine `i` writes only
to `results[i]`:

```go
results := make([]bool, len(hosts))
var wg sync.WaitGroup
for i, host := range hosts {
	wg.Go(func() {
		results[i] = ping(host) // each goroutine owns one slot
	})
}
wg.Wait()       // after this, every slot is filled in
return results  // same order as hosts
```

```text
hosts:    [ "db-1" , "db-2" , "db-3" ]
              │        │        │
          goroutine goroutine goroutine
              ▼        ▼        ▼
results:  [ true   , false  , true   ]
```

This is safe because no two goroutines touch the same slot, and nobody reads
the slice until `wg.Wait()` returns. Reading it **before** `Wait` returns would
be a bug: some slots might not be written yet.

## Your task

The script drains Kubernetes nodes before a kernel upgrade, then pings the
database hosts. Every drain or ping takes about 1 second.

### 1. Fix the slow drain

`DrainNodes(nodes, drain)` already uses a WaitGroup, but draining three nodes
takes 3 seconds instead of 1. Find the bug and fix it so that all drains run at
the same time and `DrainNodes` returns once all of them have finished.

```go
DrainNodes([]string{"node-a", "node-b", "node-c"}, drain)
// should take ~1s, not ~3s
```

### 2. Ping hosts concurrently

`PingAll(hosts, ping)` returns one `bool` per host, in the same order as
`hosts`. It works, but it pings one host at a time. Make it ping every host
at the same time, store each answer in its own slot of `results`, and return
only after every ping has finished.

```go
PingAll([]string{"db-1", "db-2", "db-3"}, ping)
// => []bool{true, false, true}   (after ~1s, not ~3s)
```

## Learn more

- [Go by Example: WaitGroups](https://gobyexample.com/waitgroups)
- [sync.WaitGroup.Go](https://pkg.go.dev/sync#WaitGroup.Go)
- [Go blog: Fixing for loops in Go 1.22](https://go.dev/blog/loopvar-preview)
