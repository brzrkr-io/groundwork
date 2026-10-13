# Custom types: fix the deploy phase enum

A deploy dashboard tracks every rollout through its phases. Someone wrote the
`Phase` enum in a hurry, and the dashboard is showing nonsense. Fix it.

## Learn

### Recap

As in *Custom types and enums*, `type Phase int` declares a new type whose
underlying type is `int`. Inside a `const ( ... )` block, `iota` counts up
from `0`, one per line, and a line with only a name repeats the line above.
A variable you never set holds the zero value, which is the first constant
in the block.

### Order matters

With `iota`, the order of the lines **is** the meaning. Swap two lines and
their numbers swap too, with no compile error. Every comparison like
`p > Pending` and every bit of arithmetic like `p + 1` depends on that order.

### Arithmetic keeps the type

Adding an untyped number to a custom type gives back the custom type:

```go
p := Building
next := p + 1 // next is a Phase, not an int
```

### An unknown zero value

Many Go enums start with an `Unknown` (or `Unspecified`) constant. A phase
that was never set then reads as "unknown" instead of quietly looking like a
real phase.

## Your task

The pipeline runs `Pending` → `Building` → `Testing` → `Deploying` → `Done`.

### 1. Fix the constant order

Reorder the constants in `exercise.go` so they get these values:

| Constant | Value |
|---|---|
| `Unknown` | `0` |
| `Pending` | `1` |
| `Building` | `2` |
| `Testing` | `3` |
| `Deploying` | `4` |
| `Done` | `5` |

Right now a `Phase` nobody set looks like `Pending`, and `Deploying` comes
before `Testing`.

```go
var p Phase
p == Unknown // => true
```

### 2. Fix IsActive

`IsActive` reports whether work is in progress. It must return `true` for
`Building`, `Testing` and `Deploying` only. At the moment a finished deploy
still counts as active, so the dashboard never clears.

```go
Testing.IsActive() // => true
Pending.IsActive() // => false
Done.IsActive()    // => false
```

### 3. Stop Next at the ends

`Next` returns the phase after `p`. `Done` is the last phase, so
`Done.Next()` returns `Done`. An `Unknown` phase can't move forward, so
`Unknown.Next()` returns `Unknown`. Every other phase moves one step along
the pipeline.

```go
Pending.Next()   // => Building
Deploying.Next() // => Done
Done.Next()      // => Done
Unknown.Next()   // => Unknown
```

The tests print phases as numbers, so `Done` shows up as `5` in a failure
message.

## Learn more

- [Go spec: Iota](https://go.dev/ref/spec#Iota)
- [Go by Example: Enums](https://gobyexample.com/enums)
- [Effective Go: Constants](https://go.dev/doc/effective_go#constants)
