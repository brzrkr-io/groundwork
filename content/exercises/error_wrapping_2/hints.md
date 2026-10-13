## Each bug is one line

Every fix is a one-line change on the line below its TODO. You do not need to
touch `read`, `DatabasePassword` or the error types.

## Which tool for which check

- Task 1: the verb that wraps is `%w`.
- Task 2: searching the chain for a sentinel value is `errors.Is`.
- Task 3: searching the chain for a type is `errors.AsType[*AccessDeniedError]`.

## Almost the answer

```go
return "", fmt.Errorf("fetch secret %q: %w", path, err)
```

```go
return errors.Is(err, ErrTimeout)
```

```go
if denied, ok := errors.AsType[*AccessDeniedError](err); ok {
```
