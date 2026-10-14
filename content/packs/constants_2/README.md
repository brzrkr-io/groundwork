# Constants: fix the health check config

The load balancer's health check settings live in constants. One doesn't
compile, one has the wrong value, and one hasn't been filled in.

**New here:** fixing an error at the declaration, constant expressions, `*`.

## Learn

### Recap

- `const` declares a value that never changes; `const ( ... )` groups them.
- A constant with no written type is *untyped*: it gets its type where used.
- You can't assign to a constant. You change its value where it's declared.

### Quotes make a string

`8081` is a number. `"8081"` is a string that happens to contain digits. Go
won't quietly turn one into the other:

| Written | Type | Can `+` do |
|---|---|---|
| `8081` | number (`int`) | adds: `8081 + 1` is `8082` |
| `"8081"` | `string` | joins: `"8081" + "1"` is `"80811"` |

So returning a string constant from a function whose `func` line says `int`
is a compile error:

```text
./exercise.go:12:9: cannot use Timeout (untyped string constant "30") as int value in return statement
```

The message names the constant (`Timeout`), says what it is (a string) and
what was needed (an `int`).

### Fix the value, not the line that uses it

The error points at the line that **uses** the constant: the `return`. But
that line is fine. The problem is the value, which lives on the line that
**declares** the constant, inside the `const ( ... )` block.

As in *Reading compiler errors*: the line in the message is where Go noticed
the problem, not always where you fix it. Search the file for the name to
find where it's declared.

> **Why is it written like this?** **Go requires** a function's `return` value
> to match the type after `()` in the `func` line. That type is part of the
> function's promise to its callers, so you fix the value, not the promise.

### Constants built from constants

A constant's value can be an *expression*, as long as it only uses other
constants and literals. Go works it out once, when it compiles. `*`
multiplies numbers.

```go
const (
	BatchSize  = 50
	MaxPending = BatchSize * 4 // 200
)
```

Change `BatchSize` and `MaxPending` follows automatically.

> **Why is it written like this?** **Go requires** constant expressions to be
> computable at compile time, so they can't use variables or function
> parameters. **Your choice:** deriving one setting from others, instead of
> typing the result, keeps them from drifting apart.

## Your task

Fix the health check settings. The starter does **not compile**, so every test
shows ✗ until you finish task 1. All three fixes are inside the `const ( ... )`
block; you don't need to change the functions.

### 1. Fix the probe port

**Goal:** make `HealthPort` return the number 8081 by fixing the `ProbePort`
constant.

```go
HealthPort() // => 8081
```

The error points at the `return` in `HealthPort` (line 23). The thing to fix
is on line 8, where `ProbePort` is declared.

**Think it through:**
- What type does `HealthPort` promise to return? What type is `ProbePort` now?
- What makes `"8081"` a string instead of a number?

### 2. Fix the health path

**Goal:** change the value of `HealthPath` so it is `/healthz`.

```go
HealthURL("api.internal")
// => "http://api.internal/healthz"

HealthURL("10.0.3.17")
// => "http://10.0.3.17/healthz"
```

`HealthPath` is declared on line 11. `HealthURL` already uses it.

**Think it through:**
- This value is text. What stays around it?

### 3. Compute the time to mark a target down

**Goal:** set `DownAfterSeconds` to an expression that multiplies
`CheckIntervalSeconds` by `FailureThreshold`.

```go
DownAfterSeconds // => 30
```

The load balancer checks a target every 10 seconds; after 3 failed checks in
a row, it marks the target down. Use the two constant names, not the number
`30`, so it stays right if either one changes.

**Think it through:**
- Which line declares `DownAfterSeconds`? Only its value changes.
- Look at the `MaxPending` example above.

## Learn more

- [The Go Blog: Constants](https://go.dev/blog/constants)
- [Go spec: Constant expressions](https://go.dev/ref/spec#Constant_expressions)
- [AWS: Health checks for Application Load Balancer target groups](https://docs.aws.amazon.com/elasticloadbalancing/latest/application/target-group-health-checks.html)
