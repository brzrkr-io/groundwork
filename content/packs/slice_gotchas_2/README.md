# Slice gotchas: two bugs from one buffer

Two bug reports came in for the rollout tool: a canary wave that eats the next
host in line, and a restart command that targets the wrong deployment. Both
bugs come from two slices sharing one backing array.

**New here:** nothing; practice with the full slice expression `s[low:high:max]` from *Slice gotchas: shared memory*.

## Learn

### Recap: slices share boxes

A slice is a small header: a pointer to a *backing array*, a length (`len`)
and a capacity (`cap`). Slicing a slice (`s[1:3]`) makes a new header over the
**same** array. No data is copied.

### append writes into spare room

If a slice has `cap` bigger than `len`, `append` doesn't make a new array. It
writes into the next unused box of the existing array, even if another slice
is already using that box.

```go
ports := []int{80, 443, 8080}
first := ports[:1]          // len 1, cap 3
first = append(first, 22)   // writes into box 1
// ports is now [80 22 8080]  (443 got overwritten!)
```

### Full slice expressions limit the room

A third number in the brackets sets the capacity: `s[low:high:max]` gives
`cap = max - low`. When `max` equals `high`, there is no spare room, so the
next `append` must copy into a brand-new array.

```go
ports := []int{80, 443, 8080}
two := ports[0:2:2]         // len 2, cap 2
two = append(two, 22)       // no room: copies to a new array
// ports is still [80 443 8080]
```

> **Why is it written like this?** The colons and their order (`low:high:max`)
> are **Go required**. You may leave out `low` (it means 0), but with three
> parts `high` and `max` must both be written. Which values you pass is your
> choice.

## Your task

Find and fix the sharing bug in each function. Change only the slicing; the
rest of the logic is already right.

### 1. Split the canary wave

**Goal:** make `Split(hosts, n)` return `canary` and `rest` so that appending
to `canary` never changes `rest` or `hosts`.

`canary` is the first `n` hosts, `rest` is everything after them. `n` is
always between 1 and `len(hosts)`. The test splits, then appends `"web-9"` to
`canary` and checks all three slices:

```go
hosts := []string{"web-1", "web-2", "web-3", "web-4"}
canary, rest := Split(hosts, 1)
canary = append(canary, "web-9")
// canary => ["web-1" "web-9"]
// rest   => ["web-2" "web-3" "web-4"]   (right now it starts with "web-9"!)
// hosts  => ["web-1" "web-2" "web-3" "web-4"]
```

**Think it through:**
- Which of the two returned slices has spare capacity right after it?
- Which box does that spare capacity point at, and who else uses it?

### 2. Build both restart commands

**Goal:** make `RestartCommands(namespace, deployment)` return two different
commands: one for the canary deployment, one for the main deployment.

Both commands share the prefix `kubectl rollout restart -n <namespace>`. The
canary one ends with `deployment/<deployment>-canary`, the main one with
`deployment/<deployment>`. Words are joined with single spaces (the starter
already does that). Right now both come out pointing at the main deployment.

```go
RestartCommands("prod", "api")
// canary => "kubectl rollout restart -n prod deployment/api-canary"
// main   => "kubectl rollout restart -n prod deployment/api"
```

**Think it through:**
- `args` was made with `cap 8`. How many items does it hold before the two
  `append` calls?
- Both `append` calls start from `args`. Which box do they both write into?
- Keep `make([]string, 0, 8)` as it is; change what the two `append`s start
  from.

## Learn more

- [Go spec: Full slice expressions](https://go.dev/ref/spec#Slice_expressions)
- [Go blog: Arrays, slices (and strings): The mechanics of 'append'](https://go.dev/blog/slices)
- [kubectl rollout restart](https://kubernetes.io/docs/reference/kubectl/generated/kubectl_rollout/kubectl_rollout_restart/)
