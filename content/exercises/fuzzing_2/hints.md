## One if per property

Call `got := Truncate(msg, limit)` once, then write one `if` with
`t.Errorf` for each property in the README's table.

## Task 1, nearly done

```go
got := Truncate(msg, limit)
if limit >= 0 && len(got) > limit {
	t.Errorf("Truncate(%q, %d) = %q, longer than %d bytes", msg, limit, got, limit)
}
if !strings.HasPrefix(msg, got) {
	t.Errorf("Truncate(%q, %d) = %q, not the start of the message", msg, limit, got)
}
```

## Task 2: guard the UTF-8 check

```go
if utf8.ValidString(msg) && !utf8.ValidString(got) {
	t.Errorf("Truncate(%q, %d) = %q, cuts a character in half", msg, limit, got)
}
```

## Task 3: step back to a character start

```go
cut := limit
for cut > 0 && !utf8.RuneStart(msg[cut]) {
	cut--
}
return msg[:cut]
```

`msg[limit]` exists because this code only runs when `len(msg) > limit`.
