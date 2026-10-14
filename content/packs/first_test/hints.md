## Task 1: Check valid ports

Every check has three steps: call `ParsePort`, compare what you got with what you want, and report a failure if they differ. Write them inside `TestParsePort`.

### Closer

`ParsePort` returns two values, so catch both with `port, err := ParsePort("8080")`. Stop with `t.Fatalf` if `err != nil`, then use `t.Errorf` if `port != 8080`.

### Answer

```go
port, err := ParsePort("8080")
if err != nil {
	t.Fatalf("ParsePort(%q) returned error: %v", "8080", err)
}
if port != 8080 {
	t.Errorf("ParsePort(%q) = %d, want %d", "8080", port, 8080)
}

port, err = ParsePort("443")
if err != nil {
	t.Fatalf("ParsePort(%q) returned error: %v", "443", err)
}
if port != 443 {
	t.Errorf("ParsePort(%q) = %d, want %d", "443", port, 443)
}
```

The second check uses `=` instead of `:=` because `port` and `err` already exist.

## Task 2: Check out-of-range ports

Here the only thing that matters is the error. A bad port must give one. What should your test do if `err` is `nil`?

### Closer

Throw the port away with `_` and check for a missing error: `if _, err := ParsePort("0"); err == nil { ... }`.

### Answer

```go
if _, err := ParsePort("0"); err == nil {
	t.Errorf("ParsePort(%q) returned no error, want an error", "0")
}
if _, err := ParsePort("65536"); err == nil {
	t.Errorf("ParsePort(%q) returned no error, want an error", "65536")
}
```
