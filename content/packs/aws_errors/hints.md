## Task 1: Read the error code

The SDK's error is a `*smithy.OperationError` on the outside; the API error
is deeper in the chain. Search the whole chain for something that satisfies
`smithy.APIError`. Add the imports listed in the README first.

### Closer

`errors.AsType[smithy.APIError](err)` returns the API error and an `ok`
bool. If `ok`, call `.ErrorCode()` on it; otherwise return `""`.

### Answer

```go
if apiErr, ok := errors.AsType[smithy.APIError](err); ok {
	return apiErr.ErrorCode()
}
return ""
```

`errors.AsType` returns `ok == false` for a `nil` error, so you need no
separate `nil` check.

## Task 2: Decide whether to retry

There are two reasons to retry: a throttling code or a server fault. Both
come from the same API error, so find it once and check both.

### Closer

Find the API error the same way as in task 1 and return `false` if there is
none. The fault check is `apiErr.ErrorFault() == smithy.FaultServer`; join it
to the throttling check with `||`.

### Answer

```go
apiErr, ok := errors.AsType[smithy.APIError](err)
if !ok {
	return false
}
return isThrottling(apiErr.ErrorCode()) || apiErr.ErrorFault() == smithy.FaultServer
```

## Task 3: Handle errors in LoadSettings

Inside `if err != nil`, first ask whether the error is the specific
"object missing" type. Only if it isn't, wrap and return it.

### Closer

Use `errors.AsType[*types.NoSuchKey](err)`; you only need the `ok`, so write
`_, ok := ...`. Wrap with `fmt.Errorf("load settings s3://%s/%s: %w", ...)`.

### Answer

```go
if err != nil {
	if _, ok := errors.AsType[*types.NoSuchKey](err); ok {
		return []byte(DefaultSettings), nil
	}
	return nil, fmt.Errorf("load settings s3://%s/%s: %w", bucket, key, err)
}
```

`%w` keeps the original error in the chain, so callers can still find it.
