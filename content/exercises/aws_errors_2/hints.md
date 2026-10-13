## Task 1: Fix IsNotFound

Keep the `return` line as it is. Only the line that finds `apiErr` has to
change: a type assertion looks at the outermost error only, and the SDK's
outermost error is never the API error.

### Closer

Replace the type assertion `err.(smithy.APIError)` with
`errors.AsType[smithy.APIError](err)`, and add `"errors"` to the imports.

### Answer

```go
apiErr, ok := errors.AsType[smithy.APIError](err)
return ok && apiErr.ErrorCode() == "InvalidInstanceID.NotFound"
```

`errors.AsType` walks the whole chain, so it finds the API error under the
SDK's wrappers.

## Task 2: Finish CanTerminate

Find the API error the same way as in task 1. Two codes get special answers;
everything else, including errors that aren't AWS errors, ends at the same
wrapped-error `return`.

### Closer

Inside `if apiErr, ok := errors.AsType[smithy.APIError](err); ok { ... }`,
use a `switch apiErr.ErrorCode()` with `case "DryRunOperation":` and
`case "UnauthorizedOperation":`. After the `if`, wrap with
`fmt.Errorf("dry-run terminate %s: %w", id, err)`.

### Answer

```go
if apiErr, ok := errors.AsType[smithy.APIError](err); ok {
	switch apiErr.ErrorCode() {
	case "DryRunOperation":
		return true, nil
	case "UnauthorizedOperation":
		return false, nil
	}
}
return false, fmt.Errorf("dry-run terminate %s: %w", id, err)
```
