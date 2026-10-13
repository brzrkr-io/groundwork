## Task 1: check the code before wrapping

Inside `if err != nil`, look for the `NoSuchTagSet` code **before** the
existing `return`. If it matches, return `map[string]string{}, nil`. If it
doesn't match, fall through to the wrapped error.

```go
if apiErr, ok := errors.AsType[smithy.APIError](err); ok && apiErr.ErrorCode() == "NoSuchTagSet" {
	return map[string]string{}, nil
}
```

## Task 2: one if, two conditions

Read the value once, then require both "is set" and "is not allowed":

```go
if env := tags["Env"]; env != "" && !slices.Contains(AllowedEnvs, env) {
	lines = append(lines, fmt.Sprintf("%s: Env %q is not prod, staging or dev", bucket, env))
}
```

## Task 3: one line

`slices.Sort(lines)` sorts strings alphabetically, in place. Put it just
before `return lines, nil`.
