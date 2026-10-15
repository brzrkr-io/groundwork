# Maps: fix the config bugs

Your deploy tool keeps its settings in maps: replica counts per service and
environment variables per container. Three of its helpers misuse maps in
classic ways. Find and fix each bug.

**New here:** the `maps` package and `maps.Copy`; why a `nil` map panics on write.

## Learn

### Recap: reading a map

`m[k]` gives the value for key `k`. If `k` is missing, you get the
*zero value* instead (`0` for `int`, `""` for `string`). No error, no crash.

So `m[k]` alone can't tell "stored as 0" from "not there". The *comma ok*
form can:

```go
ages := map[string]int{"newborn": 0}

age, ok := ages["newborn"] // age is 0, ok is true:  the key is there
age, ok = ages["ghost"]    // age is 0, ok is false: the key is missing
```

> **Why is it written like this?** Go requires the two names on the left and
> the comma between them. Calling the second one `ok` is convention; any name
> works. If you only need `ok`, write `_` for the value: Go requires every
> declared variable to be used, and `_` means "throw this away".

### Recap: deleting

`delete(m, k)` removes key `k`. If `k` isn't there, it does nothing. It
doesn't tell you whether it removed anything, so check *before* you delete.

### nil maps

A map declared with `var` and never created is `nil`. You can read from a
`nil` map (every key is missing), but writing to it panics:

```go
var scores map[string]int // nil: no map exists yet
_ = scores["x"]           // fine, gives 0
scores["x"] = 1           // panic: assignment to entry in nil map
```

Create a real map first, with `make(map[KeyType]ValueType)` or a literal
like `map[string]int{}`.

> **Why is it written like this?** `var` only declares a name; it doesn't
> build a map. `make` and literals both build one. Either works; `make` is
> common when you start empty.

### Copying a map

`a := b` does **not** copy a map. Both names point at the same map, so
writing through one changes the other.

To copy, build a new map and copy every entry into it. `maps.Copy(dst, src)`
from the `maps` package writes each key and value of `src` into `dst`,
replacing any value `dst` already had for that key:

```go
import "maps"

stock := map[string]int{"apples": 3}
backup := map[string]int{}
maps.Copy(backup, stock) // backup is now {"apples": 3}
```

| | `dst` | `src` |
|---|---|---|
| Is written to? | yes | no |
| May be `nil`? | **no**: it panics | yes: nothing to copy |

`maps.Copy` does the same as a `for k, v := range src { dst[k] = v }` loop,
which you'll still see in older code.

## Your task

Fix the three helpers in `exercise.go`. Each one already runs; it just gives
the wrong result or crashes.

### 1. Respect scaled-down services

**Goal:** make `Replicas` return `0` for a service that is listed with `0`.

`Replicas(replicas, service)` returns the count stored for `service`. If the
service is **not in the map**, it returns the default `2`. A service listed
with `0` was scaled down on purpose and must get `0`. Right now it gets `2`.

```go
replicas := map[string]int{"api": 5, "batch": 0}

Replicas(replicas, "api")    // => 5
Replicas(replicas, "batch")  // => 0
Replicas(replicas, "worker") // => 2
```

**Think it through:**
- The `if` checks `n == 0`. Is "the value is 0" the same question as "the key is missing"?
- Which form of map read from *Recap: reading a map* answers "is the key there?"

### 2. Count only what was really removed

**Goal:** make `Unset` return how many of `keys` were actually in `env`.

`Unset(env, keys...)` deletes every key in `keys` from `env`. It returns the
number it really removed:

- A key that wasn't in `env` doesn't count.
- A key whose value is the empty string `""` *is* there, so it counts.
- If the same key is given twice, only the first one counts.
- With no keys at all, it returns `0`.

Right now `Unset` counts every key it was given.

```go
env := map[string]string{"DEBUG": "1", "LOG_LEVEL": "info", "PORT": "8080"}

Unset(env, "DEBUG", "TRACE") // => 1, env is now {"LOG_LEVEL": "info", "PORT": "8080"}
Unset(env)                   // => 0, env is unchanged
```

**Think it through:**
- `removed++` runs on every loop pass. When should it run?
- Why can't you check the value against `""`? (Look at the third bullet.)
- `delete` doesn't report anything, so the check has to come before it.

### 3. Merge without crashing

**Goal:** make `Merge` return a new map instead of panicking.

`Merge(defaults, overrides)` returns a **new** map with every entry from
`defaults`, then `overrides` replacing or adding entries. Either argument may
be `nil`. It must not change `defaults` or `overrides`: the tests check that
both stay as they were.

```go
defaults := map[string]string{"LOG_LEVEL": "info", "PORT": "8080"}
overrides := map[string]string{"LOG_LEVEL": "debug", "REGION": "eu-west-1"}

Merge(defaults, overrides)
// => map[string]string{"LOG_LEVEL": "debug", "PORT": "8080", "REGION": "eu-west-1"}

Merge(nil, nil)
// => map[string]string{}   (an empty map, not nil)
```

Right now it crashes with `panic: assignment to entry in nil map`. A panic
stops the test run and prints the chain of calls that led to it. The panic
happens inside `maps.Copy`, but the bug is in your code: find the
`exercise.go` line in that list.

**Think it through:**
- Which argument of `maps.Copy` gets written to? Check the table above.
- Look at the line where `merged` is declared. Does it build a map, or only a name?
- Fix the declaration, not the `maps.Copy` lines.

## Learn more

- [Go blog: Go maps in action](https://go.dev/blog/maps)
- [Go spec: Index expressions (maps)](https://go.dev/ref/spec#Index_expressions)
- [Package maps: Copy](https://pkg.go.dev/maps#Copy)
- [Kubernetes: Define environment variables for a container](https://kubernetes.io/docs/tasks/inject-data-application/define-environment-variable-container/)
