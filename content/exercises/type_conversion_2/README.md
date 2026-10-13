# Type conversion: fix the capacity helpers

A teammate's capacity helpers don't build. One mixes number types, one turns
a port into a strange character, and one is still empty.

## Learn

### Recap

`float64(n)` and `int(f)` convert between number types. `int(f)` chops off
the fraction instead of rounding. Convert before dividing, or `int / int`
throws the fraction away first. `strconv.Itoa(n)` turns an `int` into its
digits; `string(n)` does not.

### Errors come in stages

Go checks your code in two stages. First the compiler checks types; then
`go vet` looks for likely bugs. You only see vet's complaints once the code
compiles. So after you fix one error, a different one may appear. That's
progress, not a new problem.

Here is the compiler's message when a function that returns `float64` tries
to return an `int`:

```text
cannot use totalMiB / pods (value of type int) as float64 value in return statement
```

## Your task

Fix the three helpers. The starter does **not compile**, so every test shows
✗ until tasks 1 and 2 are done.

### 1. Fix the average memory

`AvgMemoryMiB` returns the average memory per pod. It doesn't compile. Fix it
so it returns the exact average, including any fraction:

```go
AvgMemoryMiB(1000, 4) // => 250
AvgMemoryMiB(1000, 8) // => 125
AvgMemoryMiB(3, 2)    // => 1.5
```

`pods` is never 0.

### 2. Fix the port label

Once task 1 compiles, `go vet` reports a problem in `PortLabel`. Fix it so it
returns `port` followed by a space and the port number. You'll need to add
`import "strconv"`.

```go
PortLabel(8080) // => "port 8080"
PortLabel(443)  // => "port 443"
```

### 3. Count the affordable replicas

You have a monthly `budget` and each replica costs `costPerReplica` a month.
Change `AffordableReplicas` so it returns how many whole replicas you can
afford. You can't run part of a replica, so drop any fraction:

```go
AffordableReplicas(100, 30)   // => 3   (3.33 replicas)
AffordableReplicas(90, 30)    // => 3
AffordableReplicas(250, 12.5) // => 20
AffordableReplicas(29.99, 30) // => 0
```

`costPerReplica` is never 0.

## Learn more

- [Go spec: Conversions](https://go.dev/ref/spec#Conversions)
- [Package strconv](https://pkg.go.dev/strconv)
- [Command vet: stringintconv](https://pkg.go.dev/golang.org/x/tools/go/analysis/passes/stringintconv)
