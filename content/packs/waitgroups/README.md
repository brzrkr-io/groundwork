# WaitGroups

A `sync.WaitGroup` answers one question: "have all my goroutines finished
yet?" You met `wg.Go` in *Goroutines*. Here you see the counter underneath it,
the older `Add`/`Done` style you'll find in existing code, and how to share
one WaitGroup with a helper.

**New here:** `wg.Add`/`wg.Done`, passing `*sync.WaitGroup`. (`wg.Wait` is from *Goroutines*.)

## Learn

### A counter you can wait on

A WaitGroup holds a counter:

| Call | What it does |
|---|---|
| `wg.Go(f)` | add 1, run `f` in a new goroutine, subtract 1 when `f` returns |
| `wg.Wait()` | block until the counter is back to 0 |

```go
var wg sync.WaitGroup
for _, svc := range services {
	wg.Go(func() {
		restart(svc)
	})
}
wg.Wait() // counter is 0: every restart has finished
```

```text
counter:  0 ─Go─▶ 1 ─Go─▶ 2 ─Go─▶ 3 ···done▶ 2 ···done▶ 1 ···done▶ 0
caller:   start  start  start  Wait ······ blocked ··········· ▶ carries on
```

> **Why is it written like this?** `var wg sync.WaitGroup` with no value is
> enough: Go guarantees the zero value is a ready, empty counter.
> `func() { … }` is a function with no name, written right where it's used;
> `wg.Go` requires a `func()` with no parameters and no results. Calling
> `wg.Wait()` **after** the loop, not inside it, is what lets the goroutines
> run at the same time.

### The older Add/Done form

`wg.Go` is new (Go 1.25). Older code does the same by hand:

```go
wg.Add(1) // one more goroutine to wait for
go func() {
	defer wg.Done() // subtract 1 when this goroutine finishes
	restart(svc)
}()
```

`wg.Add(1)` must run **before** `go`. If it runs inside the goroutine, the
caller can reach `Wait` while the counter is still 0 and return too early.
`wg.Go` can't have that bug, so use it in new code.

### Pass a WaitGroup by pointer

A WaitGroup must not be copied once in use. A helper that starts goroutines
for someone else's WaitGroup takes a *pointer*, `*sync.WaitGroup`:

```go
func startScan(wg *sync.WaitGroup, dir string) {
	wg.Go(func() {
		scan(dir)
	})
}

startScan(&wg, "/var/log") // & passes the address of the caller's wg
```

> **Why is it written like this?** Go requires the `*` in the parameter type
> and the `&` at the call. You still write `wg.Go(...)` on a pointer: Go
> follows the pointer for you. Passing a copy compiles, but the helper counts
> on a copy the caller's `Wait` never sees; `go vet` reports it as `passes
> lock by value`.

### Results by index

As in *Goroutines: fix the maintenance script*, make a slice before the loop
and let each goroutine write only its own slot `results[i]`. After `Wait`, the
slots are in input order, however the goroutines finished.

## Your task

You're writing a multi-region deploy tool. In the tests every deploy, lookup
or backup takes 1 to 3 (fake) seconds, and each function must take only as
long as the slowest one.

### 1. Deploy every region at once

**Goal:** make `DeployAll(regions, deploy)` run `deploy(region)` for every
region at the same time, and return only once all of them have finished.

```go
DeployAll([]string{"us-east-1", "eu-west-1", "ap-south-1"}, deploy)
// returns after 1s (not 3s), once all three deploys are done
```

**Think it through:**
- Where does the WaitGroup need to be declared so the loop and the line after
  it can both use it?
- Which call goes inside the loop, and which goes after it?

### 2. Resolve hostnames in order

**Goal:** make `ResolveAll(hosts, resolve)` resolve every host at the same
time and return the addresses in the **same order as `hosts`**.

The lookups finish in a different order from `hosts`, but the result must not.

```go
// api.internal takes 3s, db.internal takes 1s
ResolveAll([]string{"api.internal", "db.internal"}, resolve)
// => []string{"10.0.1.5", "10.0.2.9"}, after 3s
```

**Think it through:**
- The slice `addrs` already has one slot per host. Which slot does each
  goroutine write?
- What must happen before `return addrs`, so every slot is filled in?

### 3. Start a backup from a helper

**Goal:** make `startBackup(wg, db, backup)` run `backup(db)` in a new
goroutine that `BackupAll`'s `wg.Wait()` waits for.

`BackupAll` is written for you: it calls `startBackup(&wg, db, backup)` for
each database, then `wg.Wait()`. The starter `startBackup` counts the backup
but runs it right away, so the backups happen one after another.

```go
BackupAll([]string{"orders", "users"}, backup)
// returns after 1s (not 2s), once both backups are done
```

**Think it through:**
- `wg` here is a pointer to `BackupAll`'s WaitGroup. Which of its methods
  starts *and* counts a goroutine?
- That one call does the work of `wg.Add(1)`, `backup(db)` *and*
  `wg.Done()`. Which starter lines are left once you've written it?

## Learn more

- [sync.WaitGroup](https://pkg.go.dev/sync#WaitGroup)
- [Go by Example: WaitGroups](https://gobyexample.com/waitgroups)
- [go vet: copylocks](https://pkg.go.dev/golang.org/x/tools/go/analysis/passes/copylock)
