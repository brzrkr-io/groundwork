# Numbers: fix the metrics math

The monitoring helpers are giving odd answers: an error rate that is always
zero and a disk that is somehow 125% free. Find out why and fix them, after
one warm-up with `%`.

**New here:** order of operations, parentheses.

## Learn

### Recap

- `int` holds whole numbers; `float64` holds fractions.
- `int / int` drops the fractional part: `7 / 2` is 3.
- `%` gives the remainder: `7 % 2` is 1.

### Order of operations

`*`, `/` and `%` happen before `+` and `-`. Operators at the same level run
left to right. Parentheses go first of all.

```go
10 - 4 / 2   // => 8, because 4 / 2 happens first
(10 - 4) / 2 // => 3
```

Order matters with integer division too, because each `/` throws away its
fraction straight away:

```go
1 / 4 * 100 // => 0   (1 / 4 is already 0)
1 * 100 / 4 // => 25
```

> **Why is it written like this?** **Go requires** these precedence rules;
> they're the same as school maths. **Convention:** add parentheses whenever
> a reader might wonder. They cost nothing and make your meaning clear.

### Finding the bug in a formula

For a "fix it" task, check the starter by hand with the first example. Do
one operator at a time, in Go's order, and write each step down. The step
where your number goes wrong is the bug.

## Your task

Fix the monitoring helpers. Each fix is in the `return` line of one function.

### 1. Pick a shard

**Goal:** make `Shard` return the remainder of `userID` divided by `shards`.

```go
Shard(17, 5) // => 2
Shard(10, 5) // => 0
Shard(3, 8)  // => 3
```

The log store is split into `shards` pieces (*shards*). A user's logs go to
the shard with that number.

**Think it through:**
- Which operator gives a remainder?

### 2. Fix the error rate

**Goal:** make `ErrorRatePercent` return the percentage of requests that
failed as a whole number, rounded down, using only `int` arithmetic.

```go
ErrorRatePercent(5, 200)  // => 2   (2.5% rounded down)
ErrorRatePercent(50, 200) // => 25
ErrorRatePercent(1, 3)    // => 33
ErrorRatePercent(0, 100)  // => 0
```

The starter returns 0 for almost everything. `requests` is never 0.

**Think it through:**
- Work out the starter's line for `(5, 200)` step by step. What is `5 / 200` in `int`?
- Could the same operations run in a different order?

### 3. Fix the free disk percentage

**Goal:** make `DiskFreePercent` return total minus used, divided by total,
times 100.

```go
DiskFreePercent(200, 150) // => 25
DiskFreePercent(64, 16)   // => 75
DiskFreePercent(100, 0)   // => 100
```

The starter has the right operators but returns nonsense. `totalGiB` is
never 0.

**Think it through:**
- In the starter, which operator runs first?
- Which part must happen before the division?

## Learn more

- [Go spec: Operator precedence](https://go.dev/ref/spec#Operator_precedence)
- [Go spec: Integer operators](https://go.dev/ref/spec#Integer_operators)
- [Go by Example: Values](https://gobyexample.com/values)
