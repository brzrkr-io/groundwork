## Task 1: convert both sides

The function must return a `float64`, so divide two `float64`s. Don't write
`float64(totalMiB / pods)`: that divides as ints first and loses the `.5`.

## Task 2: Itoa, not string

Add `import "strconv"`, then replace `string(port)` with
`strconv.Itoa(port)`.

## Task 3: divide, then convert

`budget / costPerReplica` is a `float64` like 3.33. Converting it with
`int(...)` drops the fraction.

## Almost the answer

```go
return float64(totalMiB) / float64(pods)

return "port " + strconv.Itoa(port)

return int(budget / costPerReplica)
```
