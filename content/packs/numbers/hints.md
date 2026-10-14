## Task 1: Total the memory

The total memory is the number of replicas times the memory each one needs.
Which operator multiplies?

### Closer

`*` multiplies two `int`s: `return replicas * ...`.

### Answer

```go
return replicas * perReplicaMiB
```

## Task 2: Round the replica count up

Plain `requestsPerSec / perReplica` rounds down. The README's ceiling trick
rounds up: add something to the requests before dividing.

### Closer

The pattern is `(a + b - 1) / b`. Here `a` is `requestsPerSec` and `b` is
`perReplica`. The parentheses matter, so the adding happens before the divide.

### Answer

```go
return (requestsPerSec + perReplica - 1) / perReplica
```

## Task 3: Work out the CPU percentage

Divide what's used by the limit to get a fraction like 0.25, then turn it into
a percentage. These are `float64`s, so nothing is thrown away.

### Closer

`usedMillicores / limitMillicores` gives the fraction. Multiply it by `100`
with `*`.

### Answer

```go
return usedMillicores / limitMillicores * 100
```
