## Task 1: Requests per day

You have a number of requests each second. How many seconds are in a whole day? Multiply by that.

### Closer

A minute has 60 seconds, an hour 60 minutes, a day 24 hours. You can chain `*` in one line: `rps * 60 * ...`.

### Answer

```go
return rps * 60 * 60 * 24
```

## Task 2: Nodes needed

First work out how much CPU all the replicas need together. Then divide by what one node offers, but round up: half a node still needs a whole node.

### Closer

Store the total in a variable: `total := replicas * millicoresPerReplica`. Then use the round-up trick from *Numbers and arithmetic*: add `millicoresPerNode - 1` before dividing.

### Answer

```go
total := replicas * millicoresPerReplica
return (total + millicoresPerNode - 1) / millicoresPerNode
```

Adding one less than the divisor pushes any remainder over to the next whole number.

## Task 3: Monthly cost

There are two steps: declare a constant, then multiply three things together.

### Closer

Write the constant outside any function, above `MonthlyCost`: `const HoursPerMonth = 730`. Then multiply the price, the instances and `HoursPerMonth` with `*`.

### Answer

```go
const HoursPerMonth = 730

// in MonthlyCost:
return hourlyPrice * instances * HoursPerMonth
```
