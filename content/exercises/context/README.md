# Context and cancellation

A `context.Context` carries a "stop now" signal through your program. When a
user presses Ctrl+C, a request times out, or a deploy is aborted, the context
is how every goroutine involved finds out it should give up.

**New here:** `context.Context`, `WithCancel`, `WithTimeout`, `ctx.Done()`,
`ctx.Err()`.

## Learn

### The shape of it

```go
ctx, cancel := context.WithCancel(context.Background())
defer cancel()

go watchLogs(ctx) // the goroutine watches ctx

cancel() // later: everything watching ctx is told to stop
```

| Function | Gives you |
|---|---|
| `context.Background()` | the empty starting context; never cancelled. Use it in `main`, tests, and at the top of a request. |
| `context.WithCancel(parent)` | a child context, plus a `cancel` function |
| `context.WithTimeout(parent, d)` | the same, but it also cancels itself after `d` |

Cancelling a parent cancels all its children:

```text
Background
   └── WithTimeout(30s)          deploy
          ├── WithCancel         upload artifact
          └── WithTimeout(5s)    health check
cancel the deploy ⇒ the upload and the health check stop too
```

> **Why is it written like this?** `WithCancel` and `WithTimeout` return two
> values, so Go requires two names on the left: `ctx, cancel :=`. Calling
> `defer cancel()` on the very next line is convention, even with
> `WithTimeout`: it frees the timer as soon as you're done, and `go vet`
> warns if you forget.

### Noticing cancellation

| Method | Gives you |
|---|---|
| `ctx.Done()` | a channel that is **closed** when the context ends |
| `ctx.Err()` | `nil` while running, then why it ended: `context.Canceled` or `context.DeadlineExceeded` |

`ctx.Done()` is a channel, so it goes in a `select` next to other channels.
`time.After(d)` is a channel that receives once, after `d`. Here a backup
waits for a "ready" signal, but gives up after a minute or on cancel:

```go
select {
case <-ready:
	startBackup()
case <-time.After(time.Minute):
	return errors.New("never became ready")
case <-ctx.Done():
	return ctx.Err()
}
```

A receive can also tell you whether a channel was closed:
`msg, ok := <-ch` sets `ok` to `false` once `ch` is closed and empty.

> **Why is it written like this?** `select` waits until *one* case is ready
> and runs only that one; Go requires each `case` to be a channel send or
> receive. The order of cases is your choice.

### Conventions

- A function that might take a while takes `ctx context.Context` as its
  **first** parameter, named `ctx`.
- Pass the context you were given down. Don't store it in a struct.
- When cancelled, stop quickly and return `ctx.Err()`.

The tests use `testing/synctest`: timeouts fire instantly, but the fake clock
still moves by the right amount.

## Your task

You're building the waiting and looping parts of a deploy agent.

### 1. Sleep, but wake up when cancelled

**Goal:** make `Sleep(ctx, d)` wait for `d` and return `nil`, or return
`ctx.Err()` straight away if `ctx` is cancelled first.

```go
Sleep(context.Background(), 5*time.Second) // => nil, after 5s

ctx, cancel := context.WithCancel(context.Background())
// cancel() is called after 2s
Sleep(ctx, 5*time.Second) // => context.Canceled, after 2s
```

**Think it through:**
- `time.Sleep` can't be interrupted. Which two channels are you waiting on?
- What does each one return when it wins?

### 2. Stop a worker loop

**Goal:** make `RunWorker(ctx, jobs, handle)` call `handle(job)` for each job
from `jobs`, in order, until it has to stop.

It stops in two ways:

- `jobs` is closed: return `nil`.
- `ctx` is cancelled: return `ctx.Err()` at once, without waiting for more
  jobs.

```go
// jobs carries "deploy api", "deploy web", then is closed
RunWorker(ctx, jobs, handle) // handles both, => nil

// jobs carries "deploy api", "deploy web", stays open; ctx is cancelled at 1s
RunWorker(ctx, jobs, handle) // handles both, => context.Canceled at 1s
```

**Think it through:**
- A `for range jobs` loop can't notice `ctx`. What loop runs until you
  `return`?
- How do you tell "got a job" apart from "`jobs` was closed"?

### 3. Give a check a deadline

**Goal:** make `CheckWithTimeout(timeout, check)` call `check` with a context
that times out after `timeout`, and return what `check` returns.

Build the context from `context.Background()`, and cancel it when you're
done.

```go
// check waits for ctx.Done() and returns ctx.Err()
CheckWithTimeout(3*time.Second, check) // => context.DeadlineExceeded, after 3s

// check returns nil straight away
CheckWithTimeout(3*time.Second, check) // => nil
```

**Think it through:**
- Which function in the first table cancels itself after a while?
- What's the parent, and which of its two results do you pass to `check`?

## Learn more

- [context package](https://pkg.go.dev/context)
- [Go blog: Go Concurrency Patterns: Context](https://go.dev/blog/context)
- [Go by Example: Context](https://gobyexample.com/context)
