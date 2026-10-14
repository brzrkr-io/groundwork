## Task 1: Fix the compile error in BackupAgeLabel

Read the compiler error: it says `days` is undefined on the last line. A
variable made in an `if` short statement disappears after the `if` ends.

### Closer

Move `days := ageHours / 24` onto its own line above the `if`, and leave only
the condition `days < 1` in the `if`.

### Answer

```go
days := ageHours / 24
if days < 1 {
	return "fresh"
}
return fmt.Sprintf("%dd old", days)
```

## Task 2: Fix the order of the certificate checks

Go runs only the first branch that matches. Which check is the most urgent?
Put that one first.

### Closer

Check `daysLeft < 0` first, then `< 7`, then `< 30`. Everything left over is
`"valid"`.

### Answer

```go
if daysLeft < 0 {
	return "expired"
} else if daysLeft < 7 {
	return "renew now"
} else if daysLeft < 30 {
	return "renew soon"
}
return "valid"
```

Each later check only sees the values the earlier ones let through.

## Task 3: Finish LogLevelFor

The `debug` rule and the final `"info"` are already there. The new rule goes
between them.

### Closer

Add an `if` that compares `env` with `==`. `"production"` is fixed text, so
it goes in quotes; `env` is a parameter, so it doesn't.

### Answer

```go
if env == "production" {
	return "warn"
}
```

Put this after the `debug` check and before `return "info"`.
