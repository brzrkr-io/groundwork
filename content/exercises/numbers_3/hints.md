## Task 1: multiply it out

Multiply the rate by 60, then by 60, then by 24. You can write it all in one
expression.

## Task 2: two steps

First store the total in a variable:

```go
total := replicas * millicoresPerReplica
```

Then round `total / millicoresPerNode` up with the ceiling trick from
*Numbers and arithmetic*.

## Task 3: a constant and a product

Declare the constant above the function, outside any function body:

```go
const HoursPerMonth = 730
```

Then return `hourlyPrice * instances * HoursPerMonth`.
