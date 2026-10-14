## Task 1: Decide when to scale up

Each bullet in the README is one small comparison. All of them must hold, so
join them with "and".

### Closer

The three parts are `cpuPercent > 75`, `replicas < maxReplicas` and
`!cooldown`. Join them with `&&`.

### Answer

```go
return cpuPercent > 75 && replicas < maxReplicas && !cooldown
```

## Task 2: Decide when to scale down

This is the mirror image of task 1: low CPU instead of high, and above the
minimum instead of below the maximum.

### Closer

Use `cpuPercent < 25`, `replicas > minReplicas` and `!cooldown`, joined with
`&&`.

### Answer

```go
return cpuPercent < 25 && replicas > minReplicas && !cooldown
```

## Task 3: Decide when the service is stable

Don't repeat the rules. Call your two functions from tasks 1 and 2: stable
means "not scale up" and "not scale down".

### Closer

`ShouldScaleUp(cpuPercent, replicas, maxReplicas, cooldown)` is a `bool`.
Put `!` in front to get "should not scale up". Do the same for
`ShouldScaleDown` (it takes `minReplicas`), and join the two with `&&`.

### Answer

```go
return !ShouldScaleUp(cpuPercent, replicas, maxReplicas, cooldown) &&
	!ShouldScaleDown(cpuPercent, replicas, minReplicas, cooldown)
```
