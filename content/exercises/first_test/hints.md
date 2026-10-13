## Got and want

Every check has the same three steps: call `ParsePort`, compare the result
with what you expect, and call `t.Errorf` or `t.Fatalf` if they differ. Write
them inside `TestParsePort`, under the TODO comments.

## Read TestGrader's message

If `TestGrader` says "your tests didn't notice when ParsePort accepted port
"0"", your test has no check that would fail when `ParsePort("0")` returns no
error. Add that check.

## Task 1, nearly done

```go
port, err := ParsePort("8080")
if err != nil {
	t.Fatalf("ParsePort(%q) returned error: %v", "8080", err)
}
if port != 8080 {
	t.Errorf("ParsePort(%q) = %d, want %d", "8080", port, 8080)
}
```

For `"443"`, reuse the variables with `=` instead of `:=`.

## Task 2, nearly done

```go
if _, err := ParsePort("0"); err == nil {
	t.Errorf("ParsePort(%q) returned no error, want an error", "0")
}
```

Do the same for `"65536"`.
