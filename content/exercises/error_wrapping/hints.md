## Task 1: %w keeps the original

`fmt.Errorf` with `%w` builds the message and wraps the error you pass in:

```go
return "", fmt.Errorf("config %q: %w", name, ErrNotFound)
```

## Task 2: one check per step

Each call that can fail gets its own `if err != nil` check, and each check
wraps with the same prefix:

```go
text, err := s.Get(name)
if err != nil {
	return nil, fmt.Errorf("load config: %w", err)
}
```

Do the same for `parseLines(text)`, then return the values and `nil`.

## Task 3: errors.Is, then errors.AsType

```go
if errors.Is(err, ErrNotFound) {
	return "404 not found"
}
if perr, ok := errors.AsType[*ParseError](err); ok {
	return fmt.Sprintf("400 bad config on line %d", perr.Line)
}
```
