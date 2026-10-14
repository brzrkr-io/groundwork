## Task 1: Make the check able to fail

Look at the method the check calls when `got != want`. Does it mark the test as failed, or does it only print?

### Closer

`t.Logf` only prints. `t.Errorf` prints and marks the test failed. Keep the same message and change only the method name.

### Answer

```go
if got != want {
	t.Errorf("ShortSHA(%q) = %q, want %q", sha, got, want)
}
```

## Task 2: Fix the expected value

Read the new failure message. `ShortSHA` keeps the first 7 characters. Count the characters in `want`.

### Closer

The message looks like `... = "9f86d08", want "9f86d081"`. The `want` has one character too many.

### Answer

```go
want := "9f86d08"
```

## Task 3: Check a short commit ID

An ID with 7 characters or fewer comes back unchanged. Add a second check with the got/want pattern you already have.

### Closer

You can call the function and compare in one line: `if got := ShortSHA("a1b2c"); got != "a1b2c" { ... }`.

### Answer

```go
if got := ShortSHA("a1b2c"); got != "a1b2c" {
	t.Errorf("ShortSHA(%q) = %q, want %q", "a1b2c", got, "a1b2c")
}
```
