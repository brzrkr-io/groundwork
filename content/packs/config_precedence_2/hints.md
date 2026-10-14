## Task 1: Put the layers in the right order

Each layer overwrites what came before, so the layer applied last wins. Read
`Load` top to bottom: which block runs last right now?

### Closer

Move the `fs.Visit(...)` block so it comes **after** the `s.ApplyEnv(lookup)`
call, just before `return s, nil`.

### Answer

```go
if err := s.ApplyEnv(lookup); err != nil {
	return Settings{}, err
}
fs.Visit(func(f *flag.Flag) {
	// ... the existing switch, unchanged ...
})
return s, nil
```

## Task 2: Let HTTPS_PROXY="" disable the proxy

`lookup` returns two values, and the starter throws the second one away. That
second value is the only way to tell "set to empty" from "not set".

### Closer

Keep `ok` and use it as the condition instead of checking `v != ""`:

```go
if v, ok := lookup("HTTPS_PROXY"); ok {
```

### Answer

```go
if v, ok := lookup("HTTPS_PROXY"); ok {
	s.Proxy = v
	s.Source["proxy"] = "env"
}
```

## Task 3: Parse the interval and compression variables

Both variables follow the same shape: look the variable up, parse the string,
return an error if parsing fails, then store the value and its source.

### Closer

`time.ParseDuration(v)` returns a `time.Duration` and an error;
`strconv.ParseBool(v)` returns a `bool` and an error. Wrap failures like
`fmt.Errorf("invalid BACKUP_INTERVAL %q: %w", v, err)`.

### Answer

```go
if v, ok := lookup("BACKUP_INTERVAL"); ok {
	d, err := time.ParseDuration(v)
	if err != nil {
		return fmt.Errorf("invalid BACKUP_INTERVAL %q: %w", v, err)
	}
	s.Interval = d
	s.Source["interval"] = "env"
}
if v, ok := lookup("BACKUP_COMPRESS"); ok {
	b, err := strconv.ParseBool(v)
	if err != nil {
		return fmt.Errorf("invalid BACKUP_COMPRESS %q: %w", v, err)
	}
	s.Compress = b
	s.Source["compress"] = "env"
}
```
