# append and copy

Host lists change all the time: a canary goes first, a new node joins, a
report needs a frozen copy of the list. `append` grows a slice and `copy`
duplicates one. Together they cover most of the slice work in ops tools.

## Learn

### append adds to the end

`append` takes a slice and one or more values, and gives back a slice with the
values added at the end:

```go
hosts := []string{"web-1"}
hosts = append(hosts, "web-2")
hosts = append(hosts, "web-3", "web-4") // several values at once
// hosts is now ["web-1" "web-2" "web-3" "web-4"]
```

Always assign the result, usually back to the same variable. `append` does not
change the slice you pass in; it *returns* the longer one. Go even refuses to
compile a bare `append(hosts, "web-2")` whose result you throw away.

`append` works on a `nil` slice too, so you don't need `make` to start one.

### Append a whole slice with ...

To add every element of another slice, put `...` after it:

```go
web := []string{"web-1", "web-2"}
db := []string{"db-1"}
all := append(web, db...)
// all is ["web-1" "web-2" "db-1"]
```

Read `db...` as "all the elements of `db`". Without the dots, Go would try to
add the whole slice as one element, which does not compile. (You'll see where
the `...` comes from in *Variadic functions*.)

To put something *in front*, start a new slice with it and append the rest:

```go
order := append([]string{"lb-1"}, web...)
// order is ["lb-1" "web-1" "web-2"]
```

### len and cap

A slice has a length, `len(s)`, the number of elements in it, and a
*capacity*, `cap(s)`, the room it has before it needs more memory. When
`append` runs out of room, it allocates a bigger block (roughly double), copies
the elements over and returns a slice pointing at the new block. That is why
you must use the returned slice. For now, that's all you need about `cap`.

### Building a result in a loop

The everyday pattern: start with an empty slice and append inside a loop.

```go
var upper []string // nil: length 0, ready for append
for i := range len(hosts) {
	upper = append(upper, strings.ToUpper(hosts[i]))
}
```

If `hosts` is empty, the loop never runs and `upper` stays empty.

### copy, make, and slices.Clone

Assigning a slice (`b := a`) does **not** copy the elements. Both variables
see the same elements, so changing `a[0]` also changes `b[0]`. To get an
independent copy, make a slice of the right length and `copy` into it:

```go
backup := make([]string, len(hosts))
n := copy(backup, hosts) // copy(destination, source)
// n is the number of elements copied
```

`copy` copies only as many elements as fit: the smaller of `len(dst)` and
`len(src)`. A destination of length 0 receives nothing.

The standard library has a shortcut in the `slices` package. Add it to the
`import` block at the top of the file, then call `slices.Clone`:

```go
import "slices"

backup := slices.Clone(hosts) // same result as make + copy
```

The tests in this exercise also use `slices.Equal(a, b)`, which reports whether
two slices hold the same elements in the same order. It treats a `nil` slice
and an empty slice as equal.

## Your task

Your deploy tool rolls a release out to a list of hosts. Three helpers that
build host lists are unfinished.

### 1. Put the canary first

`DeployOrder(canary, hosts)` returns the order to deploy in: the `canary` host
first, then every host in `hosts` in the same order. Leave `hosts` unchanged.

```go
DeployOrder("web-0", []string{"web-1", "web-2"})
// => ["web-0" "web-1" "web-2"]

DeployOrder("web-0", nil)
// => ["web-0"]
```

### 2. Add the domain

`WithDomain(hosts, domain)` returns a new slice where each host name has a dot
and the domain added. Build it with `append` inside a loop. If `hosts` is empty
or `nil`, return a slice with no elements (`nil` is fine).

```go
WithDomain([]string{"web-1", "db-1"}, "prod.internal")
// => ["web-1.prod.internal" "db-1.prod.internal"]

WithDomain(nil, "prod.internal")
// => []
```

### 3. Take a snapshot

`Snapshot(hosts)` returns an independent copy of `hosts`. It must hold the same
names in the same order, and changing `hosts` afterwards must not change the
snapshot. The test checks this by editing `hosts` after calling `Snapshot`.
Use `make` + `copy`, or `slices.Clone` (remember to import `"slices"`).

```go
hosts := []string{"web-1", "web-2"}
snap := Snapshot(hosts)
hosts[0] = "web-9"
// snap is still ["web-1" "web-2"]
```

## Learn more

- [Go by Example: Slices](https://gobyexample.com/slices)
- [Go blog: Go Slices: usage and internals](https://go.dev/blog/slices-intro)
- [Go spec: Appending to and copying slices](https://go.dev/ref/spec#Appending_and_copying_slices)
- [Package slices: Clone](https://pkg.go.dev/slices#Clone)
