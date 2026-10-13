## Task 1: give the error a name

Replace `_` with `err`, then add an `if err != nil` that returns
`0, fmt.Errorf("replicas %q is not a number", s)`.

## Task 2: zero value and style

Return `0`, not `n`. Build the message with `fmt.Errorf` and `%d`, all
lowercase, no `!`.

## Task 3: return inside the if

```go
n, err := ParseReplicas(input)
if err != nil {
	return "", fmt.Errorf("scale: %v", err)
}
```

Then the final `return` lines can use `nil` as the error.
