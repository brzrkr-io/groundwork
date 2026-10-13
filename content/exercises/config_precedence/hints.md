## Task 1: Apply the config file

The `f.Port` block in `ApplyFile` is your pattern. A pointer field is `nil`
when the key is missing from the JSON, so only copy the value when it isn't.

### Closer

`*f.LogLevel` reads the value behind the pointer. Check for `nil` first, then
set both the field and its source:

```go
if f.LogLevel != nil {
	s.LogLevel = ...
	s.Source["log_level"] = "file"
}
```

### Answer

```go
if f.LogLevel != nil {
	s.LogLevel = *f.LogLevel
	s.Source["log_level"] = "file"
}
if f.Region != nil {
	s.Region = *f.Region
	s.Source["region"] = "file"
}
```

## Task 2: Apply the environment

`lookup` works like `os.LookupEnv`: it returns the value and an `ok` that says
whether the variable is set. Write one `if` per variable and only act when `ok`
is true.

### Closer

```go
if v, ok := lookup("APP_REGION"); ok {
	s.Region = v
	s.Source["region"] = "env"
}
```

For `APP_PORT`, convert `v` with `strconv.Atoi` first and return an error with
`fmt.Errorf("invalid APP_PORT %q: %w", v, err)` if that fails.

### Answer

```go
if v, ok := lookup("APP_PORT"); ok {
	port, err := strconv.Atoi(v)
	if err != nil {
		return fmt.Errorf("invalid APP_PORT %q: %w", v, err)
	}
	s.Port = port
	s.Source["port"] = "env"
}
if v, ok := lookup("APP_LOG_LEVEL"); ok {
	s.LogLevel = v
	s.Source["log_level"] = "env"
}
if v, ok := lookup("APP_REGION"); ok {
	s.Region = v
	s.Source["region"] = "env"
}
return nil
```

## Task 3: Let flags have the final say

`fs.Visit` calls your function once for each flag that was actually set on the
command line, so default flag values never sneak in. Copy the value from `cli`
for each one.

### Closer

`fs.Visit` takes a `func(f *flag.Flag)`. Switch on `f.Name` (the name without
the dash):

```go
fs.Visit(func(f *flag.Flag) {
	switch f.Name {
	case "port":
		s.Port = cli.port
		s.Source["port"] = "flag"
	}
})
```

Careful: the flag is `log-level`, but the source key is `log_level`.

### Answer

```go
fs.Visit(func(f *flag.Flag) {
	switch f.Name {
	case "port":
		s.Port = cli.port
		s.Source["port"] = "flag"
	case "log-level":
		s.LogLevel = cli.logLevel
		s.Source["log_level"] = "flag"
	case "region":
		s.Region = cli.region
		s.Source["region"] = "flag"
	}
})
return s, nil
```
