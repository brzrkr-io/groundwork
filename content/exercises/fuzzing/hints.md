## The property is "no panic"

You don't compare anything inside `f.Fuzz`. Calling `ParseTarget(s)` is the
whole check: if it panics, the fuzz target fails and shows you the input.

## Task 1, nearly done

```go
f.Add("web-1:8080")
f.Add("[::1]:443")
f.Fuzz(func(t *testing.T, s string) {
	ParseTarget(s)
})
```

## Read the panic

The output names the corpus file (`FuzzParseTarget/empty_target`) and the
line in `exercise.go` that panicked. `index out of range [0] with length 0`
means `s[0]` on an empty string. `slice bounds out of range [:-1]` means
`strings.Index` found no `]` and returned `-1`.

## Task 2, nearly done

```go
if strings.HasPrefix(s, "[") {
	end := strings.Index(s, "]")
	if end < 0 {
		return "", 0, fmt.Errorf("target %q has no closing ]", s)
	}
	// ... the rest stays the same
```
