## Task 1: Check length and prefix

Call `Truncate` once inside `f.Fuzz`, then write one `if` with `t.Errorf` for each property. The fuzz function takes both arguments, `msg string` and `limit int`.

### Closer

Seeds take both values: `f.Add("disk full", -1)`. The length check only applies when `limit >= 0`. `strings.HasPrefix(msg, got)` tells you whether `got` is the start of `msg`.

### Answer

```go
f.Add("disk 92% full on db-1", 10)
f.Add("disk full", -1)
f.Fuzz(func(t *testing.T, msg string, limit int) {
	got := Truncate(msg, limit)
	if limit >= 0 && len(got) > limit {
		t.Errorf("Truncate(%q, %d) = %q, which is %d bytes, more than %d", msg, limit, got, len(got), limit)
	}
	if !strings.HasPrefix(msg, got) {
		t.Errorf("Truncate(%q, %d) = %q, which is not the start of the message", msg, limit, got)
	}
})
```

## Task 2: Check that characters stay whole

The fuzzer also makes broken UTF-8 input, and you can't expect clean output from that. So only check the result when the input was valid.

### Closer

`utf8.ValidString(s)` reports whether `s` is valid UTF-8. Combine two calls with `&&`.

### Answer

```go
if utf8.ValidString(msg) && !utf8.ValidString(got) {
	t.Errorf("Truncate(%q, %d) = %q, which cuts a character in half", msg, limit, got)
}
```

## Task 3: Fix Truncate

Cutting at `limit` bytes can land in the middle of a character like `é`, which takes 2 bytes. When that happens, step the cut back until it sits at the start of a character.

### Closer

Start with `cut := limit` and loop while `cut > 0 && !utf8.RuneStart(msg[cut])`, moving `cut` back by one. Then return `msg[:cut]`. Treat a negative limit as `0` first.

### Answer

```go
func realTruncate(msg string, limit int) string {
	limit = max(limit, 0)
	if len(msg) <= limit {
		return msg
	}
	cut := limit
	for cut > 0 && !utf8.RuneStart(msg[cut]) {
		cut--
	}
	return msg[:cut]
}
```

`msg[limit]` exists here because this code only runs when `len(msg) > limit`.
