## Task 1: Fix the untagged-bucket bug

Inside `if err != nil`, check for the `NoSuchTagSet` error code **before** the
existing `return`. If it matches, the bucket simply has no tags; anything else
is still a real error.

### Closer

Get the API error with `errors.AsType[smithy.APIError](err)` and compare
`apiErr.ErrorCode()` with `"NoSuchTagSet"`. On a match, return an empty map
(`map[string]string{}`) and `nil`. Add `"errors"` and
`"github.com/aws/smithy-go"` to the imports.

### Answer

```go
if err != nil {
	if apiErr, ok := errors.AsType[smithy.APIError](err); ok && apiErr.ErrorCode() == "NoSuchTagSet" {
		return map[string]string{}, nil
	}
	return nil, fmt.Errorf("get tags for bucket %q: %w", bucket, err)
}
```

## Task 2: Check the Env value

Read `tags["Env"]` once. You only add a line when it is set (an empty one was
already reported as missing) **and** not in `AllowedEnvs`.

### Closer

`slices.Contains(AllowedEnvs, env)` tells you whether the value is allowed.
Combine both conditions with `&&`, and use `%q` in `fmt.Sprintf` to put quotes
around the value. Add `"slices"` to the imports.

### Answer

```go
if env := tags["Env"]; env != "" && !slices.Contains(AllowedEnvs, env) {
	lines = append(lines, fmt.Sprintf("%s: Env %q is not prod, staging or dev", bucket, env))
}
```

## Task 3: Sort the report

The lines are plain strings, so the standard alphabetical sort is all you
need, just before returning.

### Closer

The `slices` package has a function that sorts a slice in place; it needs no
compare function for strings.

### Answer

```go
slices.Sort(lines)
return lines, nil
```
