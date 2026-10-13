## Task 1: move days out of the if

A variable from an `if` short statement disappears after the `if`. Declare it
first, then check it:

```go
days := ageHours / 24
if days < 1 {
```

## Task 2: most urgent first

Check `daysLeft < 0` first, then `< 7`, then `< 30`. Everything left over is
`"valid"`.

## Task 3: one more if

Between the `debug` check and the final `return "info"`, add:

```go
if env == "production" {
	return "warn"
}
```

## Almost the answer for task 2

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
