## Task 1: multiply

The total is the number of replicas times the memory for each one. `*`
multiplies.

## Task 2: use the ceiling trick

Plain `requestsPerSec / perReplica` rounds down. Add `perReplica - 1` to the
requests before dividing. The parentheses matter.

## Task 3: divide, then scale

`usedMillicores / limitMillicores` gives a fraction like 0.25. Multiply by
100 to get a percentage. These are `float64`s, so nothing is truncated.

## Almost the answer

```go
return replicas * perReplicaMiB

return (requestsPerSec + perReplica - 1) / perReplica

return usedMillicores / limitMillicores * 100
```
