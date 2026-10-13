# Constants: fix the health check config

The load balancer's health check settings live in constants. One doesn't
compile, one has the wrong value, and one hasn't been filled in.

## Learn

### Recap

`const` declares a value that never changes, and a `const ( ... )` block
groups related constants. A constant with no written type is *untyped*: it
takes its type from where it is used. The compiler stops you from assigning
to a constant.

### Quotes make a string

`8081` is a number. `"8081"` is a string that happens to contain digits. Go
won't quietly turn one into the other, so returning a string constant from a
function that returns `int` is a compile error:

```text
cannot use ProbePort (untyped string constant "8081") as int value in return statement
```

The message names the constant, says what it is (an untyped string constant)
and what the code needed (an `int`).

### Constants built from constants

A constant's value can be an expression, as long as it only uses other
constants and literals. Go works it out once, when it compiles. `*` multiplies
numbers.

```go
const (
	ShortTimeout = 5
	LongTimeout  = ShortTimeout * 6 // 30
)
```

Change `ShortTimeout` and `LongTimeout` follows automatically.

## Your task

Fix the health check settings. The starter does **not compile**, so every test
shows ✗ until you finish task 1.

### 1. Fix the probe port

`HealthPort` returns `ProbePort`, but the code doesn't compile. Read the
error and fix the **constant** so `HealthPort` returns the number 8081:

```go
HealthPort() // => 8081
```

### 2. Fix the health path

The load balancer probes `/healthz`, but `HealthPath` says `/health`. Fix the
constant so `HealthURL` returns the right URL:

```go
HealthURL("api.internal")
// => "http://api.internal/healthz"

HealthURL("10.0.3.17")
// => "http://10.0.3.17/healthz"
```

### 3. Compute the time to mark a target down

The load balancer checks a target every `CheckIntervalSeconds` (10) seconds.
After `FailureThreshold` (3) failed checks in a row, it marks the target
down. Set `DownAfterSeconds` to an expression that multiplies those two
constants, so it is 30 now and stays right if either one changes:

```go
DownAfterSeconds // => 30
```

## Learn more

- [The Go Blog: Constants](https://go.dev/blog/constants)
- [Go spec: Constant expressions](https://go.dev/ref/spec#Constant_expressions)
- [AWS: Health checks for Application Load Balancer target groups](https://docs.aws.amazon.com/elasticloadbalancing/latest/application/target-group-health-checks.html)
