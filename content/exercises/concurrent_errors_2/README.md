# Errors from goroutines: stop on first failure

When a multi-region deploy fails in one region, you usually want to stop the
others straight away instead of finishing a doomed rollout. In this exercise
you build a small *group*: run tasks together, keep the first error, and
cancel everyone else. Real projects use `golang.org/x/sync/errgroup` for this.
After building one yourself, you'll know exactly what it does.

## Learn

### Recap

Collect errors by giving each goroutine its own slot, then `errors.Join`
them after `wg.Wait()`. A context created with `context.WithCancel` tells
every goroutine watching `ctx.Done()` to stop when you call `cancel()`.
`sync.Once` runs a function exactly once, even when many goroutines call `Do`
at the same time.

### The plan

```text
WithContext(parent) ──▶ group + ctx (a child of parent)

g.Go(task A) ──▶ goroutine ── error! ──▶ once.Do: remember err, cancel()
g.Go(task B) ──▶ goroutine ── sees ctx.Done() ──▶ returns early
g.Go(task C) ──▶ goroutine ── sees ctx.Done() ──▶ returns early

g.Wait() ──▶ wg.Wait(), cancel(), return the first error
```

The pieces:

```go
type Group struct {
	cancel context.CancelFunc // cancels the group's ctx
	wg     sync.WaitGroup     // counts running tasks
	once   sync.Once          // makes sure only the first error is kept
	err    error              // the first error
}
```

- **Only the first error.** After the first failure, the other tasks usually
  fail with `context.Canceled`. That error is noise. `once.Do` makes sure
  only the first one is saved, and that it is saved safely when several tasks
  fail at once.
- **Cancel on failure.** The same `once.Do` also calls `g.cancel()`, so every
  task watching the group's `ctx` stops early.
- **Cancel in Wait too.** Once everything is done, the context is no longer
  needed, so `Wait` cancels it to release its resources (like `defer
  cancel()` usually does).
- `g.err` is written inside `once.Do` and read only after `wg.Wait()`, so it
  needs no mutex.

## Your task

### 1. Create the group and wait for it

`WithContext(ctx)` should create a cancelable child of `ctx`, and return a
`*Group` holding its `cancel` function, plus the child context.

`g.Wait()` should wait for every task started with `g.Go`, then call
`g.cancel()`, then return `g.err`.

```go
g, ctx := WithContext(context.Background())
ctx.Err()  // => nil: still running
g.Wait()   // => nil (no tasks)
ctx.Err()  // => context.Canceled: Wait cancelled it
```

### 2. Run tasks and keep the first error

`g.Go(f)` should run `f` in a new goroutine that `g.Wait()` waits for. If
`f` returns an error, use `g.once.Do` to save it in `g.err` and call
`g.cancel()`. Only the first error is kept.

```go
g, ctx := WithContext(context.Background())
g.Go(failsAfter1sWith(errA))
g.Go(failsAfter2sWith(errB))
g.Wait() // => errA (the first failure), after 2s
```

### 3. Deploy every region, stop on the first failure

`DeployRegions(ctx, regions, deploy)` should deploy to every region at the
same time using your group. Pass each `deploy` call the **group's** context,
so the remaining deploys are cancelled as soon as one fails. Return the first
error, or `nil`.

```go
// eu-west-1 fails after 1s; the others take 5s but watch ctx
DeployRegions(ctx, []string{"us-east-1", "eu-west-1", "ap-south-1"}, deploy)
// => the eu-west-1 error, after 1s; the other two deploys were cancelled
```

## Learn more

- [sync.Once](https://pkg.go.dev/sync#Once)
- [errgroup.WithContext](https://pkg.go.dev/golang.org/x/sync/errgroup#WithContext)
- [Go blog: Go Concurrency Patterns: Context](https://go.dev/blog/context)
