# Pointer receivers: a rolling deploy

A *rolling update* moves a service to a new version a few replicas at a time,
so most of the service keeps running while the rest restarts. Kubernetes
Deployments work this way. In this exercise you build a small tracker for one.

## Learn

### Recap

A method that changes its value needs a pointer receiver, `(r *Rollout)`;
with a value receiver the changes land on a copy. Go adds the `&` for you
when you call a pointer method on a variable. If any method of a type needs a
pointer receiver, give every method of that type one, as in
*Pointer receivers*. Inside a method you can call the type's other methods on
the receiver: `r.Status()`.

### Reading the starter

`Pause` is already written. Use it as a model:

```go
func (r *Rollout) Pause() {
	r.Paused = true
}
```

The other three methods are stubs, and they were sketched with value
receivers. Choosing the right receiver is part of each task: by the end,
every `Rollout` method should have a pointer receiver.

## Your task

`Replicas` is how many copies of the service run. `Updated` is how many of
them already run the new version. An operator can pause the rollout at any
time.

### 1. Report the status

Complete `Status` so it describes the rollout. Check these in order:

| Situation | Status |
|---|---|
| `Paused` is `true` | `"paused at <Updated>/<Replicas>"` |
| `Updated` equals `Replicas` | `"done"` |
| anything else | `"rolling out <Updated>/<Replicas>"` |

```go
r := &Rollout{Service: "api", Version: "v2", Replicas: 5, Updated: 2}
r.Status() // => "rolling out 2/5"

r.Pause()
r.Status() // => "paused at 2/5"
```

You'll format numbers into the string, so add `import "fmt"` to the file.

### 2. Advance by one batch

Complete `Advance(batch)` so it updates the next `batch` replicas and returns
how many it updated:

- If the rollout is paused, or `batch` is `0` or less, update nothing and
  return `0`.
- Never update more replicas than are left. If 1 replica is left and `batch`
  is 3, update 1 and return 1. A finished rollout returns `0`. The built-in
  `min(a, b)` returns the smaller of two numbers, which fits this rule well.

```go
r := &Rollout{Service: "api", Version: "v2", Replicas: 5}
r.Advance(2) // => 2, r.Updated == 2
r.Advance(2) // => 2, r.Updated == 4
r.Advance(2) // => 1, r.Updated == 5
r.Advance(2) // => 0, r.Updated == 5
```

### 3. Run the whole rollout

Complete `Run(batch)` so it keeps calling `Advance(batch)` until a call
returns `0`. After every call that updated at least one replica, append
`Status()` to a slice. Return the slice. If the very first call returns `0`,
return an empty slice (`nil` is fine).

```go
r := &Rollout{Service: "api", Version: "v2", Replicas: 5}
r.Run(2)
// => ["rolling out 2/5", "rolling out 4/5", "done"], r.Updated == 5

p := &Rollout{Service: "api", Version: "v2", Replicas: 5, Updated: 1, Paused: true}
p.Run(2)
// => [], p.Updated == 1
```

## Learn more

- [A Tour of Go: Pointer receivers](https://go.dev/tour/methods/4)
- [Go FAQ: Should I define methods on values or pointers?](https://go.dev/doc/faq#methods_on_values_or_pointers)
- [Kubernetes: Performing a rolling update](https://kubernetes.io/docs/tutorials/kubernetes-basics/update/update-intro/)
