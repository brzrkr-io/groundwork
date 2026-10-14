# Naming things: fix the misleading names

A bad name slows the reader down. A *misleading* name is worse: the reader
believes it and writes a bug. Each bug in this exercise was written by
someone who trusted a name. Fix the name, and the bug becomes obvious.

**New here:** units in names, names that hide side effects, `time.Millisecond`.

## Learn

### Recap

- A name says what a value is, or what a function does, in the team's words.
- Rename with your editor's **Rename Symbol**, then fix the comment by hand,
  as in *Naming things*.

### Put the unit in the name

`timeout := 30` could be 30 seconds or 30 milliseconds. The code compiles
either way, so a wrong guess becomes a bug in production: a retry that
waits 8 minutes instead of half a second.

When a plain number carries a unit, say the unit in the name:
`timeoutSeconds`, `sizeBytes`, `delayMillis`. Then a mismatch looks wrong
on the page:

```go
time.Duration(sizeBytes) * time.Second // bytes times seconds? suspicious
```

Even better, use a type that carries its own unit. A `time.Duration` knows
it's a duration, so `30 * time.Second` can't be mistaken for milliseconds.

To turn an `int` of milliseconds into a `time.Duration`, multiply by
`time.Millisecond`, just as you multiply by `time.Second` for seconds.

> **Why is it written like this?** **Go requires** the conversion
> `time.Duration(n)` before multiplying: `n * time.Second` with an `int`
> variable `n` doesn't compile, because Go never mixes types silently. The
> `Millis` suffix is **convention** (`Ms` is also common); the exact word is
> **your choice**, as long as the team uses one.

### A name is a promise

`filterHealthy` could mean "keep the healthy ones" or "filter out the
healthy ones". Readers pick the first meaning. If the code does the second,
every caller who didn't read the body has a bug.

Prefer names that can only be read one way: say **what you get back**
(`unhealthyNodes`), not **what the function does on the way** (`filter`,
`process`, `handle`).

### Lookups mustn't change things

A function called `ownerOf`, `getConfig` or `isReady` sounds like a
question. Callers expect to ask it as often as they like with no effect.
If it secretly writes something (to a map, a file, a counter), that write
is a *side effect* the name hides.

```go
// countErrors counts the lines with "ERROR" in them... and also deletes the log.
```

You have two honest fixes: rename it so the side effect is in the name
(`claimOwner`, `loadOrCreateConfig`), or remove the side effect so the name
is true.

Maps are a sneaky case. A map is a reference to shared data, so a function
that writes to a map parameter changes the caller's map too.

## Your task

Three small helpers in the deploy bot have names that lie. The tests show
the bug each one caused.

### 1. Put the unit in the delay's name

**Goal:** rename `delay` to `delayMillis`, then fix `BackoffDelay` so it
returns the right `time.Duration`.

```go
BackoffDelay(0)  // => 500ms
BackoffDelay(3)  // => 4s
BackoffDelay(20) // => 30s (the cap)
```

`delayMillis` itself is correct: it returns milliseconds. Rename it in
its comment too.

**Think it through:**
- After the rename, read the `return` line in `BackoffDelay` out loud. Does the unit match?
- Which `time` constant means one millisecond?

### 2. Rename the helper that returns the opposite

**Goal:** rename `filterHealthy` to `unhealthyNodes`, give it a comment that
says what it returns, then fix `ServingCount`.

`ServingCount` returns how many nodes can serve traffic: the healthy ones.

```go
// web-1 healthy, web-2 unhealthy, web-3 healthy, web-4 healthy
ServingCount(nodes) // => 3
DrainList(nodes)    // => ["web-2"]
```

The new comment starts with `unhealthyNodes`. `DrainList` is already
correct; don't change what it returns.

**Think it through:**
- Read the loop in `filterHealthy`. Which nodes does it keep?
- Once you know how many nodes are unhealthy, how do you get the healthy count from all the nodes?

### 3. Make the lookup a lookup

**Goal:** change `ownerOf` so it only reads `owners`, never writes to it.

`PageTargets` must still send services with no owner to `"platform"`, and
the caller's map must be exactly the same after the call:

```go
owners := map[string]string{"api": "payments", "db": "data"}
PageTargets(owners, []string{"api", "cache", "db", "queue"})
// => ["payments", "platform", "data", "platform"]
// owners is still {"api": "payments", "db": "data"}
```

**Think it through:**
- Which line in `ownerOf` writes to the map?
- Without that line, what should `ownerOf` return for a missing service?

## Learn more

- [time.Duration and its constants](https://pkg.go.dev/time#Duration)
- [Go Code Review Comments: Variable names](https://go.dev/wiki/CodeReviewComments#variable-names)
- [Go blog: Go maps in action](https://go.dev/blog/maps)
