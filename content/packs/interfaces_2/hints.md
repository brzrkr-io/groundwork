## Task 1: Fix HTTPCheck

Put the interface and `HTTPCheck`'s methods side by side. `Name` matches; look closely at the name of the other method.

### Closer

The interface asks for `Check() bool`, but `HTTPCheck` has `Checked() bool`. Tip: adding `var _ HealthChecker = HTTPCheck{}` makes the compiler name the problem.

### Answer

```go
// Check reports whether the last request returned 200.
func (h HTTPCheck) Check() bool {
	return h.StatusCode == 200
}
```

The method name now matches the interface exactly, so `HTTPCheck` satisfies `HealthChecker`.

## Task 2: Fix TCPCheck

This time the method name is right. Compare its result type with the interface.

### Closer

`Check` must return a `bool`, not a `string`. `c.Open` is already a `bool` that holds the answer.

### Answer

```go
// Check reports whether the port is open.
func (c TCPCheck) Check() bool {
	return c.Open
}
```

Both name and result type must match for a type to satisfy an interface.

## Task 3: Fix FailingChecks

The loop keeps a check when `c.Check()` is `true`, which means healthy. You want the opposite.

### Closer

`!` flips a `bool`: `if !c.Check() { ... }`.

### Answer

```go
for _, c := range checks {
	if !c.Check() {
		failing = append(failing, c.Name())
	}
}
```

Now only checks that report `false` (unhealthy) are collected.
