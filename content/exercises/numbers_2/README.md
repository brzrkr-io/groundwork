# Numbers: fix the metrics math

The monitoring helpers are giving odd answers: an error rate that is always
zero and a disk that is somehow 125% free. Find out why and fix them, after
one warm-up with `%`.

## Learn

### Recap

`int` holds whole numbers and `float64` holds fractions. `+ - * /` do what
you'd expect, except that `int / int` drops the fractional part: `7 / 2` is 3.
`%` gives the remainder: `7 % 2` is 1.

### Order of operations

`*`, `/` and `%` happen before `+` and `-`. Operators at the same level run
left to right. When in doubt, add parentheses: they cost nothing and make
your meaning clear.

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

## Your task

Fix the monitoring helpers.

### 1. Pick a shard

Your log store is split into `shards` pieces (*shards*). A user's logs go to
the shard numbered by the remainder of the user ID divided by the shard
count. Change `Shard` to return it:

```go
Shard(17, 5) // => 2
Shard(10, 5) // => 0
Shard(3, 8)  // => 3
```

### 2. Fix the error rate

`ErrorRatePercent` returns the percentage of requests that failed, as a whole
number rounded down. Right now it returns 0 for almost everything. Fix it
using only `int` arithmetic:

```go
ErrorRatePercent(5, 200)  // => 2   (2.5% rounded down)
ErrorRatePercent(50, 200) // => 25
ErrorRatePercent(1, 3)    // => 33
ErrorRatePercent(0, 100)  // => 0
```

`requests` is never 0.

### 3. Fix the free disk percentage

`DiskFreePercent` should return the free space as a percentage of the disk:
total minus used, divided by total, times 100. It returns nonsense instead.
Fix it:

```go
DiskFreePercent(200, 150) // => 25
DiskFreePercent(64, 16)   // => 75
DiskFreePercent(100, 0)   // => 100
```

`totalGiB` is never 0.

## Learn more

- [Go spec: Operator precedence](https://go.dev/ref/spec#Operator_precedence)
- [Go spec: Integer operators](https://go.dev/ref/spec#Integer_operators)
- [Go by Example: Values](https://gobyexample.com/values)
