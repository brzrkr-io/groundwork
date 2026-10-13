## Turn each bullet into a comparison

Each bullet in the README is one small boolean expression. Join them with
`&&`, because all of them must hold.

## Task 1 and 2: the three parts

For scaling up: `cpuPercent > 75`, `replicas < maxReplicas` and `!cooldown`.
Scaling down is the mirror image: `<` 25, `>` the minimum, and `!cooldown`.

## Task 3: call the other two

`ShouldScaleUp(cpuPercent, replicas, maxReplicas, cooldown)` is a `bool`.
Put `!` in front of it to get "should not scale up".

## Almost the answer

```go
return !ShouldScaleUp(cpuPercent, replicas, maxReplicas, cooldown) &&
	!ShouldScaleDown(cpuPercent, replicas, minReplicas, cooldown)
```
