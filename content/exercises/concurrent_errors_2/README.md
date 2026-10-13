# Errors from goroutines: stop on first failure

When a multi-region deploy fails in one region, you want to stop the others
straight away instead of finishing a doomed rollout. Here you build a small
*group*: run tasks together, keep the first error, and cancel everyone else.
Real projects use `golang.org/x/sync/errgroup`; after this you'll know what
it does.

**New here:** a cancel-on-first-error group, `sync.Once` for "only the first".

## Learn

### Recap

- `context.WithCancel(parent)` returns a child context and a `cancel`
  function. Calling `cancel()` closes `ctx.Done()`, so everyone watching stops.
  Cancelling the parent cancels the child too.
- `sync.Once` runs a function exactly once, even when many goroutines call
  `Do` at the same time:

```go
var once sync.Once
for range 3 {
	go once.Do(func() { fmt.Println("printed once") })
}
```

### The plan

```text
WithContext(parent) ──▶ group + ctx (a child of parent)

g.Go(task A) ──▶ goroutine ── error! ──▶ once.Do: remember err, cancel()
g.Go(task B) ──▶ goroutine ── sees ctx.Done() ──▶ returns early
g.Go(task C) ──▶ goroutine ── sees ctx.Done() ──▶ returns early

g.Wait() ──▶ wg.Wait(), cancel(), return the first error
```

The `Group` struct is already written for you:

| Field | Job |
|---|---|
| `cancel` | cancels the group's `ctx` |
| `wg` | counts running tasks |
| `once` | makes sure only the first error is kept |
| `err` | the first error |

- **Only the first error.** After the first failure, the others usually
  fail with `context.Canceled`. That's noise, so `once.Do` saves only the
  first, safely, even if several fail together.
- **Cancel on failure.** The same `once.Do` also calls `g.cancel()`.
- **Cancel in Wait too.** When everything is done the context is no longer
  needed, so `Wait` cancels it to release it (like `defer cancel()`).
- `g.err` is written inside `once.Do` and read only after `wg.Wait()`, so it
  needs no mutex.

> **Why is it written like this?**
> - **Go requires** a pointer receiver (`g *Group`) here: `WaitGroup` and
>   `Once` must never be copied, and methods must change the real group.
> - **Convention:** constructors that also give a context return it last, and
>   the caller usually *shadows* its `ctx` with it: `g, ctx := WithContext(ctx)`.
> - **Author's choice:** the field names, copied from `errgroup`.

## Your task

### 1. Create the group and wait for it

**Goal:** `WithContext(ctx)` returns a `*Group` and a cancelable child of `ctx`; `g.Wait()` waits, cancels, and returns the first error.

`WithContext` stores the child's cancel function in the `cancel` field.
`Wait` does three things in order: wait for every task, call `g.cancel()`,
return `g.err`. Both have a `TODO(1)`.

```go
g, ctx := WithContext(context.Background())
ctx.Err()  // => nil: still running
g.Wait()   // => nil (no tasks)
ctx.Err()  // => context.Canceled: Wait cancelled it
```

**Think it through:**
- Which `context` function gives you both a child and a cancel function?
- How do you build a `&Group{...}` with only one field set?

### 2. Run tasks and keep the first error

**Goal:** `g.Go(f)` runs `f` in a new goroutine that `g.Wait()` waits for, and saves only the first error.

On an error, use `g.once.Do` to set `g.err` and call `g.cancel()`. Three
1-second tasks must finish together in 1 second.

```go
g, ctx := WithContext(context.Background())
g.Go(failsAfter1sWith(errA))
g.Go(failsAfter2sWith(errB))
g.Wait() // => errA (the first failure), after 2s
```

**Think it through:**
- The starter calls `f()` directly. Which field can start a goroutine *and* count it?
- What two things go inside the function you pass to `once.Do`?

### 3. Deploy every region, stop on the first failure

**Goal:** `DeployRegions(ctx, regions, deploy)` deploys every region at once with your group and returns the first error, or `nil`.

Pass each `deploy` call the **group's** context, not the caller's, so the
other deploys are cancelled as soon as one fails. If the caller cancels its
`ctx`, `DeployRegions` returns `context.Canceled` right away.

```go
// eu-west-1 fails after 1s; the others take 5s but watch ctx
DeployRegions(ctx, []string{"us-east-1", "eu-west-1", "ap-south-1"}, deploy)
// => the eu-west-1 error, after 1s; the other two deploys were cancelled
```

**Think it through:**
- Use the three pieces you built: which comes first, which in a loop, which last?
- `g.Go` wants a `func() error`, but `deploy` takes two arguments. How do you wrap it?

## Learn more

- [sync.Once](https://pkg.go.dev/sync#Once)
- [errgroup.WithContext](https://pkg.go.dev/golang.org/x/sync/errgroup#WithContext)
- [Go blog: Go Concurrency Patterns: Context](https://go.dev/blog/context)
