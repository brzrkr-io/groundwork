## Task 1: Fix the average memory

Read the compiler error. The function must return a `float64`, but `totalMiB / pods` is an int.

### Closer

Convert both sides to `float64` before dividing. Don't write `float64(totalMiB / pods)`: that divides as ints first and loses the `.5`.

### Answer

```go
return float64(totalMiB) / float64(pods)
```

## Task 2: Fix the port label

`string(port)` doesn't give you the digits; it gives the character with that number. Which `strconv` function turns an int into its digits?

### Closer

Add `import "strconv"`, then replace `string(port)` with `strconv.Itoa(port)`.

### Answer

```go
return "port " + strconv.Itoa(port)
```

## Task 3: Count the affordable replicas

Divide the budget by the cost to get a `float64` like 3.33. Then turn it into an `int`, which drops the fraction.

### Closer

Divide first, then convert the result: `int(...)` around `budget / costPerReplica`.

### Answer

```go
return int(budget / costPerReplica)
```
