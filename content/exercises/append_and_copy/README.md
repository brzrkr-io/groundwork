# append and copy

Host lists change all the time: a canary goes first, a new node joins, a
report needs a frozen copy of the list. `append` grows a slice and `copy`
duplicates one. Together they cover most slice work in ops tools.

**New here:** `append`, `...` after a slice, `len` vs `cap`, `copy`,
`slices.Clone`.

## Learn

### append adds to the end

`append` takes a slice and one or more values. It gives back a slice with the
values added at the end:

```go
tags := []string{"prod"}
tags = append(tags, "eu")
tags = append(tags, "web", "v2") // several values at once
// tags is now ["prod" "eu" "web" "v2"]
```

`append` works on a `nil` slice too, so you don't need `make` to start one.

> **Why is it written like this?** Go requires you to use the result:
> `append` *returns* the longer slice, and a bare `append(tags, "eu")` does not
> compile. Assigning it back to the same variable is the convention.

### Append a whole slice with ...

To add every element of another slice, put `...` right after it:

```go
eu := []string{"fra", "ams"}
us := []string{"iad"}
regions := append(eu, us...)
// regions is ["fra" "ams" "iad"]
```

Read `us...` as "all the elements of `us`". Without the dots, Go would try to
add the whole slice as one element, which does not compile. (You'll see where
`...` comes from in *Variadic functions*.)

> **Why is it written like this?** Go requires the three dots, with no space,
> after the slice.

### len and cap

`len(s)` is how many elements a slice holds. `cap(s)`, the *capacity*, is how
much room it has before it needs more memory. When `append` runs out of room,
it moves the elements to a bigger block and returns a slice pointing there.
That is another reason to always use the returned slice.

### Building a result in a loop

Start with an empty slice and append inside a loop:

```go
var loud []string // nil: length 0, ready for append
for i := range len(alerts) {
	loud = append(loud, strings.ToUpper(alerts[i]))
}
```

If `alerts` is empty, the loop never runs and `loud` stays `nil`.

### Copies: copy and slices.Clone

`b := a` does **not** copy the elements. Both names share them, so changing
`a[0]` also changes `b[0]`. For an independent copy, make a slice with the
right length and `copy` into it:

```go
saved := make([]int, len(ports))
n := copy(saved, ports) // copy(destination, source)
// n is the number of elements copied
```

`copy` only fills slots that exist: it copies the smaller of `len(dst)` and
`len(src)`.

The `slices` package has a one-call shortcut, `slices.Clone(s)`, which returns
a new slice with its own elements. To use it, add `import "slices"` near the
top of the file, below the `package` line.

> **Why is it written like this?** Go requires an `import` for every package
> you use; `append`, `copy`, `len` and `make` are built in and need none.
> Destination-first order in `copy` is fixed by Go.

The tests also use `slices.Equal(a, b)`: same elements in the same order. It
counts a `nil` slice and an empty slice as equal.

## Your task

Your deploy tool rolls a release out to a list of hosts. Three helpers that
build host lists are unfinished.

### 1. Put the canary first

**Goal:** make `DeployOrder(canary, hosts)` return the `canary` host first,
then every host in `hosts`, in the same order.

`order` already starts with the canary. Don't change `hosts`. With no hosts,
the result is just the canary.

```go
DeployOrder("web-0", []string{"web-1", "web-2"})
// => ["web-0" "web-1" "web-2"]

DeployOrder("web-0", nil)
// => ["web-0"]
```

**Think it through:**
- You're adding a whole slice, not one value. What goes after its name?
- Where does the result of `append` have to go?

### 2. Add the domain

**Goal:** make `WithDomain(hosts, domain)` return each host name with a dot
and the domain added.

Build it with `append` inside a loop. The pieces: the host name (from
`hosts`), the fixed text `"."`, and the `domain` parameter, with no spaces. If
`hosts` is empty or `nil`, return a slice with no elements (`nil` is fine).

```go
WithDomain([]string{"web-1", "db-1"}, "prod.internal")
// => ["web-1.prod.internal" "db-1.prod.internal"]

WithDomain(nil, "prod.internal")
// => []
```

**Think it through:**
- `out` already exists. What should each loop step add to it?
- Which parts are parameters (no quotes) and which is fixed text (in quotes)?

### 3. Take a snapshot

**Goal:** make `Snapshot(hosts)` return an independent copy of `hosts`.

Same names, same order. Changing `hosts` afterwards must not change the
snapshot; the test checks this by editing `hosts[0]` after the call. Use
`make` + `copy`, or `slices.Clone` (import `"slices"` first).

```go
hosts := []string{"web-1", "web-2"}
snap := Snapshot(hosts)
hosts[0] = "web-9"
// snap is still ["web-1" "web-2"]
```

**Think it through:**
- Why does `return hosts` share the elements instead of copying them?
- If you use `copy`, how long must the destination be?

## Learn more

- [Go by Example: Slices](https://gobyexample.com/slices)
- [Go blog: Go Slices: usage and internals](https://go.dev/blog/slices-intro)
- [Go spec: Appending to and copying slices](https://go.dev/ref/spec#Appending_and_copying_slices)
- [Package slices: Clone](https://pkg.go.dev/slices#Clone)
