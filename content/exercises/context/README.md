# Context and cancellation

A `context.Context` carries a "stop now" signal through your program. When a
user presses Ctrl+C, a request times out, or a deploy is aborted, the context
is how every goroutine involved finds out it should give up.

## Learn

### The shape of it

```go
import "context"

ctx, cancel := context.WithCancel(context.Background())
defer cancel()

go worker(ctx) // the worker watches ctx

// later, when the work is no longer wanted:
cancel() // every goroutine watching ctx is told to stop
```

- `context.Background()` is the empty starting context. It is never
  cancelled. Use it in `main`, in tests, and at the top of a request.
- `context.WithCancel(parent)` returns a new *child* context, plus a
  `cancel` function that cancels it.
- `context.WithTimeout(parent, d)` is the same, but it also cancels itself
  after `d`.
- Cancelling a parent cancels all its children:

```text
Background
   └── WithTimeout(30s)          deploy
          ├── WithCancel         upload artifact
          └── WithTimeout(5s)    health check
cancel the deploy ⇒ the upload and the health check are cancelled too
```

Always `defer cancel()` right after creating a context, even with
`WithTimeout`. It frees the context's timer as soon as you are done.
`go vet` warns if you forget.

### Noticing cancellation

| Method | Gives you |
|---|---|
| `ctx.Done()` | a channel that is **closed** when the context is cancelled |
| `ctx.Err()` | `nil` while running, then why it ended: `context.Canceled` or `context.DeadlineExceeded` |

`ctx.Done()` is just a channel, so it goes in a `select` next to your other
cases:

```go
select {
case <-time.After(d):
	return nil // waited the full time
case <-ctx.Done():
	return ctx.Err() // stopped early: Canceled or DeadlineExceeded
}
```

Compare errors with `errors.Is(err, context.Canceled)`, as in *Wrapping
errors*.

### Conventions

- A function that might take a while takes `ctx context.Context` as its
  **first** parameter, named `ctx`.
- Pass the context you were given down to the functions you call. Don't
  store it in a struct.
- When a context is cancelled, stop quickly and return `ctx.Err()` (or an
  error that wraps it).

### About the tests

The tests use `testing/synctest`, so timeouts fire instantly in real time
while the fake clock still moves by the right amount.

## Your task

You're building the waiting and looping parts of a deploy agent.

### 1. Sleep, but wake up when cancelled

`Sleep(ctx, d)` waits for `d` and returns `nil`. If `ctx` is cancelled first,
it returns `ctx.Err()` straight away.

```go
Sleep(context.Background(), 5*time.Second) // => nil, after 5s

ctx, cancel := context.WithCancel(context.Background())
// cancel() is called after 2s
Sleep(ctx, 5*time.Second) // => context.Canceled, after 2s
```

### 2. Stop a worker loop

`RunWorker(ctx, jobs, handle)` calls `handle(job)` for each job it receives
from `jobs`, in order. It stops in two ways:

- `jobs` is closed: return `nil`.
- `ctx` is cancelled: return `ctx.Err()` without waiting for more jobs.

```go
// jobs carries "deploy api", "deploy web", then is closed
RunWorker(ctx, jobs, handle) // handles both, => nil

// jobs carries "deploy api", "deploy web", stays open; ctx is cancelled at 1s
RunWorker(ctx, jobs, handle) // handles both, => context.Canceled at 1s
```

### 3. Give a check a deadline

`CheckWithTimeout(timeout, check)` creates a context from
`context.Background()` that times out after `timeout`, calls `check` with
it, and returns whatever `check` returns. Cancel the context when you are
done.

```go
// check waits for ctx.Done() and returns ctx.Err()
CheckWithTimeout(3*time.Second, check) // => context.DeadlineExceeded, after 3s

// check returns nil straight away
CheckWithTimeout(3*time.Second, check) // => nil
```

## Learn more

- [context package](https://pkg.go.dev/context)
- [Go blog: Go Concurrency Patterns: Context](https://go.dev/blog/context)
- [Go by Example: Context](https://gobyexample.com/context)
