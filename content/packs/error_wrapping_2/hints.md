## Task 1: Fix the broken chain in FetchSecret

The message is already right; only the formatting verb is wrong. Which verb
keeps the original error reachable instead of just copying its text?

### Closer

`%v` copies the text, `%w` wraps the error. Change only the verb on the line
under the TODO: `fmt.Errorf("fetch secret %q: ...", path, err)`.

### Answer

```go
return "", fmt.Errorf("fetch secret %q: %w", path, err)
```

`%w` prints the same text but keeps `err` in the chain for `errors.Is` and
`errors.AsType`.

## Task 2: Fix ShouldRetry

`==` only compares the outermost error. You need a function from the `errors`
package that searches the whole chain for a sentinel value.

### Closer

`errors.Is(err, target)` walks the chain and returns `false` for `nil`, so the
whole body is one call.

### Answer

```go
return errors.Is(err, ErrTimeout)
```

`errors.Is` compares values, so a new error with the same text still returns
`false`.

## Task 3: Fix Explain

The type assertion `err.(*AccessDeniedError)` only looks at the outermost
error. Swap it for the `errors` function that searches the chain for a type.

### Closer

`errors.AsType[*AccessDeniedError](err)` returns the same two values as the
type assertion: the typed error and an `ok` bool. Only the `if` line changes.

### Answer

```go
if denied, ok := errors.AsType[*AccessDeniedError](err); ok {
	return fmt.Sprintf("ask an admin to let role %q read %q", denied.Role, denied.Path)
}
```

`errors.AsType` unwraps layer by layer until it finds an `*AccessDeniedError`.
