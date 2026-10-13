## Task 1: numbers have no quotes

`"8081"` is a string. Remove the quotes so `ProbePort` is the number `8081`.

## Task 2: one character

Only the constant's value changes. `HealthURL` itself is already right.

## Task 3: multiply the constants

Use the constant names, not the number 30, so the value updates itself if the
interval or threshold changes:

```go
DownAfterSeconds = CheckIntervalSeconds * FailureThreshold
```
