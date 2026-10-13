## Task 1: look at the quotes

The `alertWord` pattern is in double quotes, so Go turns `\b` into a
backspace character before the regexp package ever sees it. Use backticks.

## Task 2: dots and anchors

Two things are wrong: the `.` before `internal` matches any character, and
nothing pins the pattern to the start and end of the host. Labels can repeat,
so put "label then dot" in a group and repeat the group with `+`.

```go
var internalHost = regexp.MustCompile(`^([a-z0-9-]+\.)+internal$`)
```

## Task 3: match the value

`\S+` matches one or more non-space characters. Add it after the `=`. Then
add `REDACTED` to the end of the replacement string in `Redact`.

## Task 3: almost the answer

```go
var secretPair = regexp.MustCompile(`(password|token|secret)=\S+`)

func Redact(line string) string {
	return secretPair.ReplaceAllString(line, "${1}=REDACTED")
}
```
