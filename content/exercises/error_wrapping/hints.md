## Task 1: Wrap ErrNotFound in the store

Replace the `errors.New` call with an error that both names the config and keeps `ErrNotFound` inside it, so callers can still find it.

### Closer

`fmt.Errorf` with the `%w` verb wraps the error you pass for it. Use `%q` for the name: `fmt.Errorf("config %q: %w", ...)`.

### Answer

```go
return "", fmt.Errorf("config %q: %w", name, ErrNotFound)
```

`%w` keeps a link to `ErrNotFound`, so `errors.Is(err, ErrNotFound)` is true.

## Task 2: Wrap errors in LoadConfig

Both calls throw their error away with `_`. Give each one a name, check it, and return early with the same `load config: ` prefix.

### Closer

```go
text, err := s.Get(name)
if err != nil {
	return nil, fmt.Errorf("load config: %w", err)
}
```

Do the same for `parseLines(text)`.

### Answer

```go
text, err := s.Get(name)
if err != nil {
	return nil, fmt.Errorf("load config: %w", err)
}
values, err := parseLines(text)
if err != nil {
	return nil, fmt.Errorf("load config: %w", err)
}
return values, nil
```

Using `%w` (not `%v`) keeps `ErrNotFound` and the `*ParseError` reachable through the chain.

## Task 3: Pick a response with errors.Is and errors.AsType

The error you get may be wrapped several times. Use the `errors` package helpers that search the whole chain, in the README's order.

### Closer

`errors.Is(err, ErrNotFound)` checks for a specific value. `perr, ok := errors.AsType[*ParseError](err)` looks for a type and gives you the typed error when `ok` is true.

### Answer

```go
if errors.Is(err, ErrNotFound) {
	return "404 not found"
}
if perr, ok := errors.AsType[*ParseError](err); ok {
	return fmt.Sprintf("400 bad config on line %d", perr.Line)
}
```

Put these between the `nil` check and the final `"500 internal error"` line.
