## Task 1: Pick a shard

The shard number is the remainder of the user ID divided by the shard count.
Which operator gives you a remainder?

### Closer

`%` gives the remainder: `17 % 5` is 2. Use it with `userID` and `shards`.

### Answer

```go
return userID % shards
```

## Task 2: Fix the error rate

`errors / requests` is 0 whenever `errors` is smaller than `requests`, and 0
times 100 is still 0. Try changing the order of the operations.

### Closer

Multiply `errors` by 100 **before** you divide by `requests`:
`errors * 100 / ...`. Operators at the same level run left to right.

### Answer

```go
return errors * 100 / requests
```

## Task 3: Fix the free disk percentage

Right now `usedGiB/totalGiB*100` runs first, and the result is subtracted from
`totalGiB`. The free space (total minus used) has to be worked out first.

### Closer

Put parentheses around the subtraction so it happens first:
`(totalGiB - usedGiB) / ...`.

### Answer

```go
return (totalGiB - usedGiB) / totalGiB * 100
```
