## Task 1: when is the text built?

`"maintenance " + host + ": " + status` is worked out at the `defer` line,
while `status` is still `"failed"`. Wrap the call in a closure so the text is
built when the function finishes instead.

## Task 2: check, then defer

Move `defer r.Close(host)` below the `if err != nil { return err }` check, so
it only runs once `Connect` has succeeded. You can write the check as
`if err := r.Connect(host); err != nil {`.

## Task 3: let RotateHost do the work

Once task 2 is fixed, `RotateHost` already does one host correctly, and its
`defer` runs at the end of each call. Replace the whole loop body with a call
to it.

## Almost the answer

```go
defer func() {
	r.Record("maintenance " + host + ": " + status)
}()
```

```go
for _, host := range hosts {
	if err := RotateHost(r, host); err != nil {
		return err
	}
}
return nil
```
