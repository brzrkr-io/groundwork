## Task 1: Fix the defaults bug

The flag definitions are fine. Look at when `*duration` and `*author` are read:
before or after `fs.Parse` has filled them in?

### Closer

Move the whole `opts := Options{...}` block so it comes after the `fs.Parse`
check.

### Answer

```go
if err := fs.Parse(args); err != nil {
	return Options{}, err
}
opts := Options{
	For:    *duration,
	Author: *author,
}
return opts, nil
```

## Task 2: Add a repeatable -label flag

Make a map before the flag, and have an `fs.Func` closure add one pair to it
each time `-label` appears. Put the map in `Options` after parsing.

### Closer

Start with `labels := map[string]string{}`. Inside the function, use
`key, value, ok := strings.Cut(s, "=")` and reject `!ok`, an empty key or an
empty value. Add `"errors"` and `"strings"` to the imports.

### Answer

```go
labels := map[string]string{}
fs.Func("label", "match alerts with this key=value label (repeatable)", func(s string) error {
	key, value, ok := strings.Cut(s, "=")
	if !ok || key == "" || value == "" {
		return errors.New("label must look like key=value")
	}
	labels[key] = value
	return nil
})
```

Then add `Labels: labels` to `Options`.

## Task 3: Add a validated -env flag

Use a plain variable that starts at the default, and an `fs.Func` that only
changes it for an allowed value.

### Closer

`env := "dev"`, then inside the function a `switch s` with
`case "dev", "staging", "prod":` that sets `env = s` and returns `nil`.
Anything else returns the error.

### Answer

```go
env := "dev"
fs.Func("env", "environment: dev, staging or prod (default dev)", func(s string) error {
	switch s {
	case "dev", "staging", "prod":
		env = s
		return nil
	}
	return errors.New("env must be dev, staging or prod")
})
```

Then add `Env: env` to `Options`.
