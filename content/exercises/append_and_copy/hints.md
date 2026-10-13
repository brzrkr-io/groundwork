## Task 1: Put the canary first

`order` already holds the canary. You need to add every host from `hosts` to
the end of it with `append`, and keep what `append` gives back.

### Closer

To append all elements of a slice, put `...` after it, and assign the result:
`order = append(order, ...)`.

### Answer

```go
order = append(order, hosts...)
```

`hosts...` adds each host as its own element, and `append` builds a new result
without changing `hosts`.

## Task 2: Add the domain

Loop over `hosts` and, for each one, append the host name plus `"."` plus
`domain` to `out`. `out` starts as `nil`, which is fine for `append`.

### Closer

Use `for i := range len(hosts)`, and inside the loop assign the result:
`out = append(out, hosts[i] + ...)`.

### Answer

```go
for i := range len(hosts) {
	out = append(out, hosts[i]+"."+domain)
}
```

With no hosts the loop never runs, so `out` stays `nil`, which the task allows.

## Task 3: Take a snapshot

`return hosts` hands back the same elements, so an edit to `hosts` shows up in
the "copy" too. You need a new slice with its own elements.

### Closer

Either make a slice of the same length and `copy` into it, or use the shortcut
from the `slices` package (add `import "slices"` at the top):

```go
snap := make([]string, len(hosts))
copy(snap, hosts)
```

### Answer

```go
import "slices"

// in Snapshot:
return slices.Clone(hosts)
```

`slices.Clone` allocates new elements, so later changes to `hosts` don't reach
the snapshot.
