## Task 1: Make FieldError a proper error

Both methods are one line. `Error` joins the field name and the inner
error's message; `Unwrap` hands back the inner error.

### Closer

`Error` builds a string from `e.Field`, `": "` and `e.Err.Error()`. `Unwrap`
returns `e.Err`, which is what lets `errors.Is` see the category.

### Answer

```go
func (e *FieldError) Error() string {
	return e.Field + ": " + e.Err.Error()
}

func (e *FieldError) Unwrap() error {
	return e.Err
}
```

## Task 2: Validate name, port and replicas

Each rule is one `if` that appends a `*FieldError` to `errs`. Keep the order
name, port, replicas, and add the `fmt` import.

### Closer

```go
if c.Port < 1 || c.Port > 65535 {
	errs = append(errs, &FieldError{
		Field: "port",
		Err:   fmt.Errorf("%w: %d is not between 1 and 65535", ErrOutOfRange, c.Port),
	})
}
```

Name uses `c.Name == ""` and `Err: ErrRequired`; replicas works like port.

### Answer

```go
if c.Name == "" {
	errs = append(errs, &FieldError{Field: "name", Err: ErrRequired})
}
if c.Port < 1 || c.Port > 65535 {
	errs = append(errs, &FieldError{
		Field: "port",
		Err:   fmt.Errorf("%w: %d is not between 1 and 65535", ErrOutOfRange, c.Port),
	})
}
if c.Replicas < 1 || c.Replicas > 50 {
	errs = append(errs, &FieldError{
		Field: "replicas",
		Err:   fmt.Errorf("%w: %d is not between 1 and 50", ErrOutOfRange, c.Replicas),
	})
}
```

## Task 3: Validate the image and the environment variables

The image check goes between name and port: empty first, then "no tag".
For env, finish `ValidEnvName` by checking each character, then loop over the
keys in sorted order.

### Closer

For the image, check `c.Image == ""`, then `else if` with
`_, tag, found := strings.Cut(c.Image, ":"); !found || tag == ""`. In
`ValidEnvName`, return `false` for an empty name, then range over the runes
and test `r >= 'A' && r <= 'Z' || r >= '0' && r <= '9' || r == '_'`. Loop the
env with `slices.Sorted(maps.Keys(c.Env))`.

### Answer

```go
// between the name and port checks
if c.Image == "" {
	errs = append(errs, &FieldError{Field: "image", Err: ErrRequired})
} else if _, tag, found := strings.Cut(c.Image, ":"); !found || tag == "" {
	errs = append(errs, &FieldError{
		Field: "image",
		Err:   fmt.Errorf("%w: %q has no tag", ErrInvalid, c.Image),
	})
}

// after the replicas check
for _, key := range slices.Sorted(maps.Keys(c.Env)) {
	if !ValidEnvName(key) {
		errs = append(errs, &FieldError{
			Field: "env." + key,
			Err:   fmt.Errorf("%w: use only A-Z, 0-9 and _", ErrInvalid),
		})
	}
}
```

```go
func ValidEnvName(name string) bool {
	if name == "" {
		return false
	}
	for _, r := range name {
		ok := r >= 'A' && r <= 'Z' || r >= '0' && r <= '9' || r == '_'
		if !ok {
			return false
		}
	}
	return true
}
```

Sorting the keys makes the report come out in the same order every time.

## Task 4: List the fields with problems

First get a list of errors: the joined list if `err` has `Unwrap() []error`,
otherwise just `err` on its own. Then pull a `*FieldError` out of each one.
Task 5 needs the same steps, so a small helper pays off.

### Closer

Start with `list := []error{err}`, and replace it when
`err.(interface{ Unwrap() []error })` succeeds. For each `e`, use
`fe, ok := errors.AsType[*FieldError](e)`. Return `nil` early for a `nil`
error.

### Answer

```go
func fieldErrors(err error) []*FieldError {
	if err == nil {
		return nil
	}
	list := []error{err}
	if joined, ok := err.(interface{ Unwrap() []error }); ok {
		list = joined.Unwrap()
	}
	var found []*FieldError
	for _, e := range list {
		if fe, ok := errors.AsType[*FieldError](e); ok {
			found = append(found, fe)
		}
	}
	return found
}

func FieldNames(err error) []string {
	var names []string
	for _, fe := range fieldErrors(err) {
		names = append(names, fe.Field)
	}
	return names
}
```

## Task 5: Filter by category

This is `FieldNames` with one extra condition. Reuse the helper from task 4
and only keep a field when its error matches `kind`.

### Closer

Loop over `fieldErrors(err)` and append `fe.Field` only when
`errors.Is(fe, kind)` is true.

### Answer

```go
var names []string
for _, fe := range fieldErrors(err) {
	if errors.Is(fe, kind) {
		names = append(names, fe.Field)
	}
}
return names
```

`errors.Is` can see the category because of your `Unwrap` from task 1.
