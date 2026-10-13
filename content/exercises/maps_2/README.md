# Maps: fix the config bugs

Your deploy tool keeps its settings in maps: replica counts per service and
environment variables per container. Three of its helpers misuse maps in
classic ways. Find and fix each bug.

## Learn

### Recap

`m[k]` gives the zero value when `k` is missing, and `v, ok := m[k]` tells
you whether it was really there. `delete(m, k)` removes a key, and does
nothing if it's missing. A map passed to a function is shared with the
caller. A `nil` map (declared with `var` and never made) can be read but
panics when you write to it. Create maps with `make` or a literal first.

### Copying a map

`a := b` does not copy a map. Both names point at the same map, so writing
through one changes the other. To make a real copy, create a new map and copy
every entry across with `maps.Copy(dst, src)` from the `maps` package. It
writes each key and value of `src` into `dst`, replacing any value `dst`
already had for that key:

```go
import "maps"

copied := make(map[string]string)
maps.Copy(copied, original)
```

`maps.Copy` does the same as a `for k, v := range original { copied[k] = v }`
loop, which you'll still see in older code. Like that loop, it writes into
`dst`, so `dst` must be a real map, not `nil`. A `nil` `src` is fine: there is
just nothing to copy.

## Your task

Fix the three helpers in `exercise.go`.

### 1. Respect scaled-down services

`Replicas` returns how many copies of `service` to run. If the service isn't
in the map, the default is `2`. But a service that is listed with `0` has been
scaled down on purpose, and must get `0`. Right now it gets `2`, because the
code can't tell "listed as 0" from "not listed".

```go
replicas := map[string]int{"api": 5, "batch": 0}

Replicas(replicas, "api")    // => 5
Replicas(replicas, "batch")  // => 0
Replicas(replicas, "worker") // => 2
```

### 2. Count only what was really removed

`Unset` deletes every key in `keys` from `env` and returns how many keys it
actually removed. Keys that weren't there don't count. A key whose value is
the empty string `""` *is* there, so it counts. If the same key is given
twice, only the first one removes something. Right now `Unset` counts every
key it was given.

```go
env := map[string]string{"DEBUG": "1", "LOG_LEVEL": "info", "PORT": "8080"}

Unset(env, "DEBUG", "TRACE") // => 1, env is now {"LOG_LEVEL": "info", "PORT": "8080"}
Unset(env)                   // => 0, env is unchanged
```

### 3. Merge without crashing

`Merge` returns a **new** map holding every entry from `defaults`, with
`overrides` replacing or adding entries. Right now it crashes with
`panic: assignment to entry in nil map`. Fix it so it works. Don't change
`defaults` or `overrides`: the tests check that both stay as they were.

```go
defaults := map[string]string{"LOG_LEVEL": "info", "PORT": "8080"}
overrides := map[string]string{"LOG_LEVEL": "debug", "REGION": "eu-west-1"}

Merge(defaults, overrides)
// => map[string]string{"LOG_LEVEL": "debug", "PORT": "8080", "REGION": "eu-west-1"}

Merge(nil, nil)
// => map[string]string{}
```

A panic stops the whole test run, so while `Merge` crashes, the test output
ends with the panic message and a list of lines it passed through. The panic
happens inside `maps.Copy`, but look for `exercise.go` in that list to find
the line of your code that called it.

## Learn more

- [Go blog: Go maps in action](https://go.dev/blog/maps)
- [Go spec: Index expressions (maps)](https://go.dev/ref/spec#Index_expressions)
- [Package maps: Copy](https://pkg.go.dev/maps#Copy)
- [Kubernetes: Define environment variables for a container](https://kubernetes.io/docs/tasks/inject-data-application/define-environment-variable-container/)
