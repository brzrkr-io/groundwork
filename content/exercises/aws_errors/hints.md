## Search the chain, not the top

The error the SDK returns is a `*smithy.OperationError` on the outside. The
part you care about is deeper in the chain, so every task starts with
`errors.AsType`. Don't forget to add the imports listed in the README.

## Task 1: find the API error

```go
if apiErr, ok := errors.AsType[smithy.APIError](err); ok {
	return apiErr.ErrorCode()
}
return ""
```

`errors.AsType` returns `ok == false` for a `nil` error, so you need no
separate `nil` check.

## Task 2: code or fault

Replace the starter's line. Find the API error the same way as in task 1,
and return `false` if there is none. Then one line does the rest:

```go
return isThrottling(apiErr.ErrorCode()) || apiErr.ErrorFault() == smithy.FaultServer
```

## Task 3: almost the answer

```go
if err != nil {
	if _, ok := errors.AsType[*types.NoSuchKey](err); ok {
		return []byte(DefaultSettings), nil
	}
	return nil, fmt.Errorf("load settings s3://%s/%s: %w", bucket, key, err)
}
```
