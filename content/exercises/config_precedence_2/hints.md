## Task 1: read Load top to bottom

Each layer overwrites what came before, so the layer applied **last** wins.
Which block runs last in `Load` right now? Flags should have the final say.

## Task 2: the ok you threw away

`lookup` returns two values, and the starter discards the second one with
`_`. That `ok` is the only way to tell "set to empty" from "not set":

```go
if v, ok := lookup("HTTPS_PROXY"); ok {
```

## Task 3: parse, check, apply

Follow the same shape for both variables:

```go
if v, ok := lookup("BACKUP_INTERVAL"); ok {
	d, err := time.ParseDuration(v)
	if err != nil {
		return fmt.Errorf("invalid BACKUP_INTERVAL %q: %w", v, err)
	}
	s.Interval = d
	s.Source["interval"] = "env"
}
```

For `BACKUP_COMPRESS`, use `strconv.ParseBool(v)` and the key `"compress"`.
