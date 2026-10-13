# Slice gotchas: two bugs from one buffer

Two bug reports came in for the rollout tool: a canary wave that eats the next
host in line, and a restart command that targets the wrong deployment. Both
come from slices sharing one backing array.

## Learn

### Recap

A slice is a header (pointer, `len`, `cap`) over a backing array, and
sub-slices share that array. When a slice has spare capacity, `append` writes
into the next box of the array, even if another slice is using it. The full
slice expression `s[low:high:max]` caps the capacity at `max - low`, so with
`max == high` the next `append` has to copy into a new array.

```text
hosts := []string{"web-1", "web-2", "web-3"}
canary := hosts[:1]       // len 1, cap 3  -> append writes over web-2
canary := hosts[:1:1]     // len 1, cap 1  -> append copies
```

## Your task

Find and fix the sharing bug in each function. Fix only the slicing, not the
rest of the logic.

### 1. Split the canary wave

`Split(hosts, n)` returns the first `n` hosts as the canary wave and the
remaining hosts as `rest` (`n` is between 1 and `len(hosts)`). Callers often
add a host to the canary wave afterwards:

```go
hosts := []string{"web-1", "web-2", "web-3", "web-4"}
canary, rest := Split(hosts, 1)
canary = append(canary, "web-9")
// canary => ["web-1" "web-9"]
// rest   => ["web-2" "web-3" "web-4"]  (right now it says "web-9" first!)
// hosts  => ["web-1" "web-2" "web-3" "web-4"]
```

Make sure appending to `canary` can never change `rest` or `hosts`. The test
does exactly the `append` above and then checks `rest` and `hosts`.

### 2. Build both restart commands

`RestartCommands(namespace, deployment)` builds two `kubectl` commands from one
shared prefix: one for the canary deployment (the name plus `-canary`) and one
for the main deployment. Right now both come out pointing at the main
deployment.

```go
RestartCommands("prod", "api")
// canary => "kubectl rollout restart -n prod deployment/api-canary"
// main   => "kubectl rollout restart -n prod deployment/api"
```

## Learn more

- [Go spec: Full slice expressions](https://go.dev/ref/spec#Slice_expressions)
- [Go blog: Arrays, slices (and strings): The mechanics of 'append'](https://go.dev/blog/slices)
- [kubectl rollout restart](https://kubernetes.io/docs/reference/kubectl/generated/kubectl_rollout/kubectl_rollout_restart/)
