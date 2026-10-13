## Task 1: move one block

Nothing is wrong with the flag definitions. Move the `opts := Options{...}`
block so it comes after `fs.Parse`, then the `*duration` and `*author` reads
happen at the right time.

## Task 2: a map and a closure

Make the map before the flag: `labels := map[string]string{}`. Inside the
`fs.Func` function, use `strings.Cut(s, "=")` and check `ok`, `key` and
`value`. Put `Labels: labels` in `Options`.

## Task 3: start with the default

```go
env := "dev"
fs.Func("env", "dev, staging or prod", func(s string) error {
	switch s {
	case "dev", "staging", "prod":
		env = s
		return nil
	}
	return errors.New("env must be dev, staging or prod")
})
```

Then put `Env: env` in `Options`.

## Task 2 in full

```go
labels := map[string]string{}
fs.Func("label", "key=value label (repeatable)", func(s string) error {
	key, value, ok := strings.Cut(s, "=")
	if !ok || key == "" || value == "" {
		return errors.New("label must look like key=value")
	}
	labels[key] = value
	return nil
})
```
