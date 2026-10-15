# Numbers: plan the capacity

Before a launch, someone has to answer three questions: how much traffic is
that per day, how many nodes do we need, and what will it cost? You'll write
a function for each.

**New here:** nothing; practice with untyped constants in arithmetic.

## Learn

### Recap

- `int` for counts, `float64` for fractions like prices.
- `int / int` rounds down; `(a + b - 1) / b` rounds up.
- A local variable made with `:=` can hold a step of the calculation.

### Untyped constants fit either type

As in *Constants*, a constant with no written type takes its type from where
it's used. So one constant works with `int` and `float64`:

```go
const SecondsPerMinute = 60

minutes := 3
price := 0.25
minutes * SecondsPerMinute // => 180, an int
price * SecondsPerMinute   // => 15, a float64
```

> **Why is it written like this?**
> - **Go requires** constants to be declared with `const Name = value`, and a
>   constant at the top level (outside any function) can be used by every
>   function in the file.
> - **Convention:** declare a constant just above the function that uses it,
>   or at the top of the file, with a comment starting with its name.
> - **Your choice:** a named constant instead of a bare number like `60`
>   tells the reader what the number means.

## Your task

Write the three capacity planning helpers. Each one starts as `return 0`.

### 1. Requests per day

**Goal:** make `RequestsPerDay` turn requests per second (`rps`) into
requests per day.

```go
RequestsPerDay(1)  // => 86400
RequestsPerDay(10) // => 864000
```

A day has 60 × 60 × 24 seconds. In Go, `×` is written `*`.

**Think it through:**
- How many seconds in a day, and what do you multiply by it?

### 2. Nodes needed

**Goal:** make `NodesNeeded` return how many nodes all the replicas need,
rounded up. Work out the total CPU first, then divide.

```go
NodesNeeded(10, 500, 2000) // => 3   (5000 millicores, 2.5 nodes)
NodesNeeded(4, 500, 2000)  // => 1   (2000 millicores, exactly 1 node)
NodesNeeded(3, 1500, 4000) // => 2   (4500 millicores)
NodesNeeded(0, 500, 2000)  // => 0
```

Each replica asks for `millicoresPerReplica` of CPU; each node offers
`millicoresPerNode`.

**Think it through:**
- Total CPU is two parameters multiplied. Would a variable for it help?
- Then use the ceiling trick from *Numbers and arithmetic*. What plays `a` and `b`?

### 3. Monthly cost

**Goal:** declare a constant `HoursPerMonth` set to 730, then make
`MonthlyCost` return the hourly price times the instances times
`HoursPerMonth`.

```go
MonthlyCost(0.125, 4)  // => 365
MonthlyCost(0.5, 2)    // => 730
MonthlyCost(0.0625, 1) // => 45.625
```

Cloud prices are per hour, and cost reports count a month as 730 hours. The
constant goes directly under the `// HoursPerMonth ...` comment below
`TODO(3)` (line 21), outside the function.

**Think it through:**
- A constant is a number here, so does it need quotes?
- Inside `MonthlyCost`, use the constant by name.

## Learn more

- [Go spec: Arithmetic operators](https://go.dev/ref/spec#Arithmetic_operators)
- [Kubernetes: Resource management for pods and containers](https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/)
- [AWS: EC2 On-Demand pricing](https://aws.amazon.com/ec2/pricing/on-demand/)
