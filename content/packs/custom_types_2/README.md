# Custom types: fix the deploy phase enum

A deploy dashboard tracks every rollout through its phases. Someone wrote the
`Phase` enum in a hurry, and the dashboard is showing nonsense. Fix it.

**New here:** nothing; practice with `iota`, comparisons and arithmetic on custom types.

## Learn

### Recap

As in *Custom types and enums*, `type Phase int` declares a new type whose
underlying type is `int`. Inside `const ( ... )`, `iota` counts up from `0`,
one per line. A line with only a name repeats the line above.

A variable you never set holds the zero value, `0`. That is whichever
constant sits on the first line.

### Order matters

With `iota`, the order of the lines **is** the meaning. Swap two lines and
their numbers swap too, with no compile error.

```go
const (
	Small Size = iota // 0
	Large             // 1
	Medium            // 2  ← oops: Medium > Large now
)
```

Only the **first** line needs the type and `= iota`. The lines below copy it.

> **Why is it written like this?** Repeating the line above is a Go rule for
> `const` blocks. Writing `= iota` only once is convention; writing it on
> every line would also work, but nobody does.

### Comparisons and ranges

Because the values are ordered numbers, `<`, `<=`, `>` and `>=` work. To ask
"strictly between A and C", use `x > A && x < C`. `<=` would include `C`
itself.

### Arithmetic keeps the type

Adding a plain number to a custom type gives back the custom type:

```go
s := Small
bigger := s + 1 // bigger is a Size, not an int
```

### An unknown zero value

Many Go enums start with an `Unknown` (or `Unspecified`) constant. A phase
nobody set then reads as "unknown" instead of quietly looking like a real
phase.

> **Why is it written like this?** Putting `Unknown` first is convention, not
> a Go rule. Go only decides that the zero value is `0`.

## Your task

The pipeline runs `Pending` → `Building` → `Testing` → `Deploying` → `Done`.
All three fixes are in `exercise.go`, at the `TODO` markers.

### 1. Fix the constant order

**Goal:** reorder the lines of the `const` block so each constant gets the value in this table.

| Constant | Value |
|---|---|
| `Unknown` | `0` |
| `Pending` | `1` |
| `Building` | `2` |
| `Testing` | `3` |
| `Deploying` | `4` |
| `Done` | `5` |

Right now a `Phase` nobody set looks like `Pending`, and `Deploying` comes
before `Testing`. After the fix:

```go
var p Phase
p == Unknown // => true
```

**Think it through:**
- Which line must be first so it gets `0`?
- Which line should carry `Phase = iota` once you've moved things?

### 2. Fix IsActive

**Goal:** make `IsActive` return `true` for `Building`, `Testing` and `Deploying` only.

At the moment a finished deploy still counts as active, so the dashboard
never clears. `Unknown`, `Pending` and `Done` must return `false`.

```go
Testing.IsActive() // => true
Pending.IsActive() // => false
Done.IsActive()    // => false
```

**Think it through:**
- The line is almost right. Which of the two comparisons lets `Done` in?

### 3. Stop Next at the ends

**Goal:** make `Next` return the phase after `p`, except that `Done` and `Unknown` stay where they are.

- `Done` is the last phase, so `Done.Next()` returns `Done`.
- `Unknown` can't move forward, so `Unknown.Next()` returns `Unknown`.
- Every other phase moves one step along the pipeline.

```go
Pending.Next()   // => Building
Deploying.Next() // => Done
Done.Next()      // => Done
Unknown.Next()   // => Unknown
```

The tests print phases as numbers, so `Done` shows up as `5` in a failure
message.

**Think it through:**
- The existing `return` is right for most phases. What must happen *before* it for the two special ones?
- What should you return when the phase must not move?

## Learn more

- [Go spec: Iota](https://go.dev/ref/spec#Iota)
- [Go by Example: Enums](https://gobyexample.com/enums)
- [Effective Go: Constants](https://go.dev/doc/effective_go#constants)
