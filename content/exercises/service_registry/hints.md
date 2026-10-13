## Task 1: a switch

Use a `switch s` with one `case` per constant, and a `default` that returns
`fmt.Sprintf("Status(%d)", int(s))`. Add `"fmt"` to the imports.

## Task 2: check, then store

Use the comma-ok form to see whether the name is taken:

```go
if _, ok := r.services[s.Name]; ok {
	return fmt.Errorf("service %q is already registered", s.Name)
}
```

`s` is your own copy of the service, so you can set `s.Status` before storing
it with `r.services[s.Name] = s`. `Lookup` can be two lines:
`s, ok := r.services[name]` and `return s, ok`.

## Tasks 3 and 4: delete and filter

For `Deregister`, check with comma-ok first, then call
`delete(r.services, name)`.

For `ByTag`, range over `r.services`, `append` each service where
`slices.Contains(s.Tags, tag)` is true, then sort with `slices.SortFunc` and
`cmp.Compare(a.Name, b.Name)`.

## Task 5: write the copy back

```go
for name, s := range r.services {
	if hc.Check(s.Address) {
		s.Status = StatusHealthy
		healthy++
	} else {
		s.Status = StatusUnhealthy
	}
	r.services[name] = s
}
```
