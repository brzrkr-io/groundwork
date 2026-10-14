## Task 1: Exit with 128 + n

Compare the signal with each stop signal you know. `SIGINT` (`os.Interrupt`)
is signal 2 and `SIGTERM` is signal 15.

### Closer

`switch sig { case os.Interrupt: ... }` compares the signal with each case.
The codes are `128 + 2` and `128 + 15`. Add `"syscall"` to the imports.

### Answer

```go
switch sig {
case os.Interrupt:
	return 128 + 2
case syscall.SIGTERM:
	return 128 + 15
default:
	return 1
}
```

## Task 2: Reload on SIGHUP instead of quitting

The bug is that the loop body always reaches `return`. Only some signals
should return; `SIGHUP` should reload and let the loop carry on.

### Closer

Wrap the body in `switch sig`: a `case syscall.SIGHUP:` that reloads, and a
`default:` that keeps the existing log line and `return`. For the reload, use
`if err := reload(); err != nil { ... } else { ... }`.

### Answer

```go
for sig := range sigs {
	switch sig {
	case syscall.SIGHUP:
		if err := reload(); err != nil {
			fmt.Fprintf(log, "reload failed: %v; keeping old config\n", err)
		} else {
			fmt.Fprintln(log, "reloaded config")
		}
	default:
		fmt.Fprintf(log, "received %v: shutting down\n", sig)
		return ExitCodeFor(sig)
	}
}
return 0
```
