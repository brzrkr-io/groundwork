## Task 1: copy the port block

The `f.Port` block is the pattern. A pointer field is `nil` when the key was
missing, and `*f.LogLevel` reads the value when it's not:

```go
if f.LogLevel != nil {
	s.LogLevel = *f.LogLevel
	s.Source["log_level"] = "file"
}
```

Do the same for `Region` (key `"region"`).

## Task 2: one if per variable

Call `lookup` like `os.LookupEnv`, and only act when `ok` is true:

```go
if v, ok := lookup("APP_REGION"); ok {
	s.Region = v
	s.Source["region"] = "env"
}
```

For `APP_PORT`, convert first and return an error if that fails:

```go
port, err := strconv.Atoi(v)
if err != nil {
	return fmt.Errorf("invalid APP_PORT %q: %w", v, err)
}
```

## Task 3: visit the set flags

`fs.Visit` takes a function that gets one `*flag.Flag` per flag that was set.
Switch on `f.Name` (the name without the dash) and copy the value from `cli`:

```go
fs.Visit(func(f *flag.Flag) {
	switch f.Name {
	case "port":
		s.Port = cli.port
		s.Source["port"] = "flag"
	}
})
```

Add cases for `"log-level"` and `"region"`. Careful: the flag is called
`log-level`, but the source key is `log_level`.
