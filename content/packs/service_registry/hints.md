## Task 1: Name the statuses

Choose a word based on the value of `s`. A `switch` with one case per constant, plus a fallback, fits well.

### Closer

Use `switch s` with `case StatusUnknown:`, `case StatusHealthy:`, `case StatusUnhealthy:`, and a `default` that returns `fmt.Sprintf("Status(%d)", int(s))`. Add `"fmt"` to the imports.

### Answer

```go
switch s {
case StatusUnknown:
	return "unknown"
case StatusHealthy:
	return "healthy"
case StatusUnhealthy:
	return "unhealthy"
default:
	return fmt.Sprintf("Status(%d)", int(s))
}
```

`%d` with `int(s)` prints the number; `%v` on `s` would call `String` again forever.

## Task 2: Register and look up services

`Register` checks two problems (empty name, name already taken) before storing anything. `Lookup` is a plain map lookup.

### Closer

Use `errors.New` for the empty name, and the comma-ok form to see whether the name is taken:

```go
if _, ok := r.services[s.Name]; ok {
	return fmt.Errorf("service %q is already registered", s.Name)
}
```

`s` is your own copy, so set `s.Status` before storing it with `r.services[s.Name] = s`.

### Answer

```go
func (r *Registry) Register(s Service) error {
	if s.Name == "" {
		return errors.New("service name is empty")
	}
	if _, ok := r.services[s.Name]; ok {
		return fmt.Errorf("service %q is already registered", s.Name)
	}
	s.Status = StatusUnknown
	r.services[s.Name] = s
	return nil
}

func (r *Registry) Lookup(name string) (Service, bool) {
	s, ok := r.services[name]
	return s, ok
}
```

Add `"errors"` and `"fmt"` to the imports.

## Task 3: Deregister a service

First find out whether the name is there, so you know what to return. Then remove it.

### Closer

Check with comma-ok, `if _, ok := r.services[name]; !ok`, then call `delete(r.services, name)`.

### Answer

```go
if _, ok := r.services[name]; !ok {
	return false
}
delete(r.services, name)
return true
```

## Task 4: Find services by tag

Range over all services and keep the ones whose tags include `tag`. Then sort what you kept by name, because map order is random.

### Closer

`slices.Contains(s.Tags, tag)` tells you whether a service has the tag. Sort with `slices.SortFunc(found, func(a, b Service) int { ... })` and `cmp.Compare(a.Name, b.Name)`. Add `"cmp"` and `"slices"` to the imports.

### Answer

```go
var found []Service
for _, s := range r.services {
	if slices.Contains(s.Tags, tag) {
		found = append(found, s)
	}
}
slices.SortFunc(found, func(a, b Service) int {
	return cmp.Compare(a.Name, b.Name)
})
return found
```

## Task 5: Health-check everything

Loop over the map, ask `hc.Check` about each address, and update the status. The loop variable is a copy, so the change only sticks if you store it back.

### Closer

Range with both key and value: `for name, s := range r.services`. Set `s.Status`, count the healthy ones, then write `r.services[name] = s`.

### Answer

```go
healthy := 0
for name, s := range r.services {
	if hc.Check(s.Address) {
		s.Status = StatusHealthy
		healthy++
	} else {
		s.Status = StatusUnhealthy
	}
	r.services[name] = s
}
return healthy
```

Changing an existing key while ranging over a map is allowed in Go.
