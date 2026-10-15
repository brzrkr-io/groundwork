# Type conversion: fix the capacity helpers

A teammate's capacity helpers don't build. One mixes number types, one turns
a port into a strange character, and one is still empty.

**New here:** errors that come in stages (compiler, then vet).

## Learn

### Recap

- `float64(n)` and `int(f)` convert between number types.
- `int(f)` chops off the fraction instead of rounding.
- Convert before dividing, or `int / int` throws the fraction away first.
- `strconv.Itoa(n)` turns an `int` into its digits; `string(n)` does not.

### Errors come in stages

Go checks your code in two stages. First the compiler checks types; then
`go vet` looks for likely bugs. You only see vet's complaints once the code
compiles. So after you fix one error, a different one may appear. That's
progress, not a new problem.

Here's the compiler's message when a function that returns `float64` tries
to return an `int`:

```text
./exercise.go:7:9: cannot use totalMiB / pods (value of type int) as float64 value in return statement
```

It names the expression (`totalMiB / pods`), its type (`int`) and the type
the `func` line promised (`float64`).

> **Why is it written like this?** **Go requires** the `return` value to match
> the result type in the `func` line. Keep the `func` line as it is: other
> code (and the tests) rely on it. Change the expression instead.

## Your task

Fix the three helpers. The starter does **not compile**, so every test shows
✗ until tasks 1 and 2 are done.

### 1. Fix the average memory

**Goal:** make `AvgMemoryMiB` (line 7) return the exact average memory per
pod, including any fraction.

```go
AvgMemoryMiB(1000, 4) // => 250
AvgMemoryMiB(1000, 8) // => 125
AvgMemoryMiB(3, 2)    // => 1.5
```

`pods` is never 0.

**Think it through:**
- `3 / 2` as `int`s is 1, but the test wants 1.5. When must you convert?

### 2. Fix the port label

**Goal:** make `PortLabel` (line 13) return `port `, then the port number as
digits.

```go
PortLabel(8080) // => "port 8080"
PortLabel(443)  // => "port 443"
```

Once task 1 compiles, `go vet` reports this line. You'll need to add
`import "strconv"` under the `package` line.

**Think it through:**
- What does `string(...)` do to an `int`? What should be used instead?

### 3. Count the affordable replicas

**Goal:** make `AffordableReplicas` return how many **whole** replicas fit in
`budget` when each costs `costPerReplica`.

```go
AffordableReplicas(100, 30)   // => 3   (3.33 replicas)
AffordableReplicas(90, 30)    // => 3
AffordableReplicas(250, 12.5) // => 20
AffordableReplicas(10, 2.5)   // => 4
AffordableReplicas(29.99, 30) // => 0
```

You can't run part of a replica, so drop any fraction. `costPerReplica` is
never 0. Replace the starter's `return 0` line: if you leave it below your own
`return`, `go vet` reports `unreachable code`.

**Think it through:**
- The parameters are `float64`, but the function returns `int`. Which conversion chops the fraction?
- Do you drop the fraction before or after dividing? With `10` and `2.5`, the answer is 4 (`10 / 2.5`), not 5 (`10 / 2`).

## Learn more

- [Go spec: Conversions](https://go.dev/ref/spec#Conversions)
- [Package strconv](https://pkg.go.dev/strconv)
- [Command vet: stringintconv](https://pkg.go.dev/golang.org/x/tools/go/analysis/passes/stringintconv)
