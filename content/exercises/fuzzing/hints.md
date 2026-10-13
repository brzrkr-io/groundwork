## Task 1: Write the fuzz target

The property you check is only "`ParseTarget` never panics". You don't compare anything: calling the function is the whole check. Seeds give the fuzzer good starting inputs.

### Closer

Add each seed with `f.Add("web-1:8080")`. Then `f.Fuzz(func(t *testing.T, s string) { ... })`, and inside it just call `ParseTarget(s)`.

### Answer

```go
func FuzzParseTarget(f *testing.F) {
	f.Add("web-1:8080")
	f.Add("[::1]:443")
	f.Fuzz(func(t *testing.T, s string) {
		ParseTarget(s)
	})
}
```

Your target fails right after this. That's expected: it replays the saved crash inputs.

## Task 2: Fix the crashes

Read the panic. It names the saved input and the line in `exercise.go`. `index out of range [0] with length 0` means `s[0]` on an empty string. `slice bounds out of range [:-1]` means `strings.Index` found no `]`.

### Closer

Replace `s[0] == '['` with `strings.HasPrefix(s, "[")`, which is simply false for `""`. After `end := strings.Index(s, "]")`, return an error if `end < 0`.

### Answer

```go
if strings.HasPrefix(s, "[") {
	end := strings.Index(s, "]")
	if end < 0 {
		return "", 0, fmt.Errorf("target %q has no closing ]", s)
	}
	host = s[1:end]
	// ... the rest stays the same
```

An empty string then takes the other branch, finds no `:` and returns the "no port" error.
