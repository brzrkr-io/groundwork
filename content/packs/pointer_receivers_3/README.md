# Pointer receivers: a rolling deploy

A *rolling update* moves a service to a new version a few replicas at a time,
so most of the service keeps running while the rest restarts. Kubernetes
Deployments work this way. Here you build a small tracker for one.

**New here:** nothing new; practice pointer receivers.

## Learn

### Recap

A method that changes its value needs a pointer receiver, `(r *Rollout)`.
With a value receiver the changes land on a copy. If any method of a type
needs a pointer receiver, give every method of that type one (Go
convention).

Inside a method you can call the type's other methods on the receiver:

```go
func (d *Disk) Fill() {
	for d.Free() > 0 {
		d.Write(1)
	}
}
```

### Reading the starter

`Pause` is already written. Use its receiver line as a model:

```go
func (r *Rollout) Pause() {
	r.Paused = true
}
```

The other three methods were sketched with value receivers. Fixing the
receiver is part of each task: by the end, every `Rollout` method has a
pointer receiver. Only change the `(r Rollout)` part of each line; the name,
parameters and result type stay as they are.

### Numbers inside strings

`fmt.Sprintf` builds a string from a pattern; `%d` is replaced by an
integer:

```go
fmt.Sprintf("disk %d of %d", 1, 4) // "disk 1 of 4"
```

> **Why is it written like this?** `%d` is fixed by the `fmt` package. Any
> text around it is your choice, and is copied as is, spaces included.

## Your task

`Replicas` is how many copies of the service run. `Updated` is how many of
them already run the new version. An operator can pause the rollout at any
time.

### 1. Report the status

**Goal:** `Status` returns a string describing the rollout. Give it a pointer
receiver.

Check these in order; the first match wins:

| Situation | Status |
|---|---|
| `Paused` is `true` | `"paused at <Updated>/<Replicas>"` |
| `Updated` equals `Replicas` | `"done"` |
| anything else | `"rolling out <Updated>/<Replicas>"` |

`<Updated>` and `<Replicas>` stand for the field values. Add `"fmt"` to the
imports.

```go
r := &Rollout{Service: "api", Version: "v2", Replicas: 5, Updated: 2}
r.Status() // => "rolling out 2/5"

r.Pause()
r.Status() // => "paused at 2/5"
```

A paused rollout that is finished still says `"paused at 3/3"`, not
`"done"`.

**Think it through:**
- Which parts of the string are fixed text, and which come from fields?
- Why does the order of the checks matter?

### 2. Advance by one batch

**Goal:** `Advance(batch)` updates up to `batch` more replicas and returns
how many it updated. Give it a pointer receiver.

- Paused, or `batch` is `0` or less: update nothing, return `0`.
- Never update more replicas than are left. With 1 left and `batch` 3,
  update 1 and return 1. A finished rollout returns `0`.

```go
r := &Rollout{Service: "api", Version: "v2", Replicas: 5}
r.Advance(2) // => 2, r.Updated == 2
r.Advance(2) // => 2, r.Updated == 4
r.Advance(2) // => 1, r.Updated == 5
r.Advance(2) // => 0, r.Updated == 5
```

**Think it through:**
- How many replicas are left? Which two fields tell you?
- The built-in `min(a, b)` picks the smaller number. Which two numbers?

### 3. Run the whole rollout

**Goal:** `Run(batch)` keeps calling `Advance(batch)` until it returns `0`,
and returns the status after each step. Give it a pointer receiver.

- After every call that updated at least one replica, append `Status()` to a
  slice.
- If the very first call returns `0`, return an empty slice (`nil` is fine).

```go
r := &Rollout{Service: "api", Version: "v2", Replicas: 5}
r.Run(2)
// => ["rolling out 2/5", "rolling out 4/5", "done"], r.Updated == 5

p := &Rollout{Service: "api", Version: "v2", Replicas: 5, Updated: 1, Paused: true}
p.Run(2)
// => [], p.Updated == 1
```

**Think it through:**
- You don't know how many steps there will be. Which kind of `for` loop
  repeats until something happens?
- Look at the `Fill` example in *Recap*: calling other methods on `r`.

## Learn more

- [A Tour of Go: Pointer receivers](https://go.dev/tour/methods/4)
- [Go FAQ: Should I define methods on values or pointers?](https://go.dev/doc/faq#methods_on_values_or_pointers)
- [fmt.Sprintf](https://pkg.go.dev/fmt#Sprintf)
- [Kubernetes: Performing a rolling update](https://kubernetes.io/docs/tutorials/kubernetes-basics/update/update-intro/)
