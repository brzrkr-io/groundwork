## Task 1: Record the real outcome in Maintenance

`"maintenance " + host + ": " + status` is worked out at the `defer` line,
while `status` is still `"failed"`. You need the text to be built when the
function finishes instead.

### Closer

Defer a closure instead of a direct call: `defer func() { ... }()`. The
closure reads `status` when it runs.

### Answer

```go
defer func() {
	r.Record("maintenance " + host + ": " + status)
}()
```

## Task 2: Don't close what you never opened in RotateHost

Right now the `defer` is scheduled before anyone checks whether `Connect`
worked. Move it so it is only scheduled once the connection exists.

### Closer

Check the error first with `if err := r.Connect(host); err != nil { return
err }`, and put `defer r.Close(host)` below that check.

### Answer

```go
if err := r.Connect(host); err != nil {
	return err
}
defer r.Close(host)
return r.Rotate(host)
```

## Task 3: Close each host before the next in RotateAll

A `defer` in the loop waits for `RotateAll` to return. Your fixed `RotateHost`
already handles one host, and its `defer` runs at the end of each call.

### Closer

Replace the whole loop body with a call to `RotateHost(r, host)`, returning
its error if it is not `nil`.

### Answer

```go
for _, host := range hosts {
	if err := RotateHost(r, host); err != nil {
		return err
	}
}
return nil
```

Each `RotateHost` call closes its host before the loop moves on.
