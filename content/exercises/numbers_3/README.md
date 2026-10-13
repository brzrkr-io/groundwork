# Numbers: plan the capacity

Before a launch, someone has to answer three questions: how much traffic is
that per day, how many nodes do we need, and what will it cost? You'll write
a function for each.

## Learn

### Recap

Use `int` for counts and `float64` for fractions like prices. `int / int`
rounds down, and `(a + b - 1) / b` rounds up. Untyped constants, as in
*Constants*, fit in with either number type:

```go
const SecondsPerMinute = 60

minutes := 3
price := 0.25
minutes * SecondsPerMinute // => 180, an int
price * SecondsPerMinute   // => 15, a float64
```

## Your task

Write the three capacity planning helpers. Each one starts as `return 0`.

### 1. Requests per day

Change `RequestsPerDay` so it turns a steady rate in requests per second into
requests per day. A day has 60 × 60 × 24 seconds.

```go
RequestsPerDay(1)  // => 86400
RequestsPerDay(10) // => 864000
```

### 2. Nodes needed

Each replica asks for `millicoresPerReplica` millicores of CPU and each node
offers `millicoresPerNode`. Change `NodesNeeded` so it returns how many nodes
the replicas need in total, rounded up. Work out the total CPU first, then
divide.

```go
NodesNeeded(10, 500, 2000) // => 3   (5000 millicores, 2.5 nodes)
NodesNeeded(4, 500, 2000)  // => 1   (2000 millicores, exactly 1 node)
NodesNeeded(3, 1500, 4000) // => 2   (4500 millicores)
NodesNeeded(0, 500, 2000)  // => 0
```

### 3. Monthly cost

Cloud prices are per hour, and cost reports count a month as 730 hours.
Declare a constant `HoursPerMonth` set to 730, then change `MonthlyCost` so it
returns the hourly price times the number of instances times
`HoursPerMonth`:

```go
MonthlyCost(0.125, 4)  // => 365
MonthlyCost(0.5, 2)    // => 730
MonthlyCost(0.0625, 1) // => 45.625
```

## Learn more

- [Go spec: Arithmetic operators](https://go.dev/ref/spec#Arithmetic_operators)
- [Kubernetes: Resource management for pods and containers](https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/)
- [AWS: EC2 On-Demand pricing](https://aws.amazon.com/ec2/pricing/on-demand/)
