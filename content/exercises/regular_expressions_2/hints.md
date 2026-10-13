## Task 1: Fix the alert matcher

Look at how the `alertWord` pattern is quoted. In double quotes, Go handles
backslashes before the regexp package ever sees the pattern.

### Closer

In `"..."`, `\b` becomes a backspace character, so the pattern never matches.
Switch to backticks so `\b` reaches the regexp package as a word boundary,
and make sure the pattern is `\b(ERROR|FATAL)\b`.

### Answer

```go
var alertWord = regexp.MustCompile(`\b(ERROR|FATAL)\b`)
```

The word boundaries on both sides reject "ERRORS" and "NOERROR".

## Task 2: Fix the hostname check

Two things are wrong: the `.` before `internal` matches any character, and
nothing pins the pattern to the start and end of the host. Labels can also
repeat.

### Closer

Use `\.` for a real dot and wrap the pattern in `^...$`. Put "label then dot"
in a group and repeat the group with `+`: `([a-z0-9-]+\.)+`.

### Answer

```go
var internalHost = regexp.MustCompile(`^([a-z0-9-]+\.)+internal$`)
```

Each label needs at least one character, so empty labels like `.internal` and
`db..internal` fail.

## Task 3: Redact secrets

The pattern stops at the `=`, so only the key gets replaced. Extend it to also
match the value, then put `REDACTED` into the replacement.

### Closer

`\S+` matches one or more non-space characters. Add it after the `=` in
`secretPair`, and change the replacement to `"${1}=REDACTED"`.

### Answer

```go
var secretPair = regexp.MustCompile(`(password|token|secret)=\S+`)

func Redact(line string) string {
	return secretPair.ReplaceAllString(line, "${1}=REDACTED")
}
```

`${1}` puts the matched key back, so only the value changes.
