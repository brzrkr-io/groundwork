# WaitGroups

A `sync.WaitGroup` answers one question: "have all my goroutines finished
yet?" You met `wg.Go` in *Goroutines*. This exercise shows the counter
underneath it, the older `Add`/`Done` style you'll still see in existing code,
and why `wg.Go` is the safer choice.

## Learn

### A counter you can wait on

A WaitGroup holds a counter:

| Call | What it does |
|---|---|
| `wg.Go(f)` | add 1, run `f` in a new goroutine, subtract 1 when `f` returns |
| `wg.Wait()` | block until the counter is back to 0 |

```go
var wg sync.WaitGroup
for _, region := range regions {
	wg.Go(func() {
		deploy(region)
	})
}
wg.Wait() // counter is 0: every deploy has finished
```

```text
counter:  0 ─Go─▶ 1 ─Go─▶ 2 ─Go─▶ 3 ···done▶ 2 ···done▶ 1 ···done▶ 0
main:     start  start  start  Wait ······ blocked ··········· ▶ carries on
```

### The older Add/Done form

`wg.Go` is new (Go 1.25). You'll see this in existing code, and it does the
same thing by hand:

```go
wg.Add(1) // one more goroutine to wait for
go func() {
	defer wg.Done() // subtract 1 when this goroutine finishes
	deploy(region)
}()
```

Your editor suggests rewriting it as `wg.Go`, and for good reason. `wg.Add(1)`
must happen **before** `go`. If it runs inside the new goroutine, `main` can
reach `Wait` while the counter is still 0, and `Wait` returns while the
deploys are still running. That bug *usually* works on your laptop and fails
in production. `wg.Go` can't have it, so use `wg.Go` in new code.

### Pass a WaitGroup by pointer

A WaitGroup must not be copied after you start using it. If a helper function
needs one, pass a pointer:

```go
func startDeploy(wg *sync.WaitGroup, region string) {
	wg.Go(func() {
		deploy(region)
	})
}
```

If you pass `sync.WaitGroup` by value, the helper adds to a copy, and your
`Wait` never sees it. `go vet` reports this as `passes lock by value`.

### Results by index

As in *Goroutines: fix the maintenance script*, give each goroutine its own
slot in a slice made before the loop. After `Wait`, every slot is filled in,
in the same order as the input, however the goroutines finished.

## Your task

You're writing a multi-region deploy tool. Every deploy, DNS lookup or backup
in the tests takes 1 to 3 (fake) seconds.

### 1. Deploy every region at once

`DeployAll(regions, deploy)` should call `deploy(region)` for every region at
the same time and return once all of them have finished. Use `wg.Go` and
`wg.Wait`.

```go
DeployAll([]string{"us-east-1", "eu-west-1", "ap-south-1"}, deploy)
// returns after ~1s, once all three deploys are done
```

### 2. Resolve hostnames in order

`ResolveAll(hosts, resolve)` should resolve every host at the same time and
return the addresses in the **same order as `hosts`**, even though the lookups
finish in a different order.

```go
ResolveAll([]string{"api.internal", "db.internal"}, resolve)
// => []string{"10.0.1.5", "10.0.2.9"}
```

### 3. Start a backup from a helper

`BackupAll` is written for you. It calls `startBackup(&wg, db, backup)` for
each database, then `wg.Wait()`. The starter `startBackup` counts the backup
in `wg` but runs it right away, so the backups happen one after another. Fix
it: start `backup(db)` in a new goroutine that `wg.Wait()` waits for, using
the `wg` pointer it receives.

```go
BackupAll([]string{"orders", "users"}, backup)
// returns after ~1s, once both backups are done
```

## Learn more

- [sync.WaitGroup](https://pkg.go.dev/sync#WaitGroup)
- [Go by Example: WaitGroups](https://gobyexample.com/waitgroups)
- [go vet: copylocks](https://pkg.go.dev/golang.org/x/tools/go/analysis/passes/copylock)
