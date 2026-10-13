## Task 1: a switch on the signal

`switch sig { case os.Interrupt: ... }` compares the signal with each case.
`SIGINT` is signal 2 and `SIGTERM` is signal 15, so the codes are `128 + 2` and
`128 + 15`. Add `"syscall"` to the imports.

## Task 2: only some signals return

The bug is that the loop body always reaches `return`. Wrap it in a `switch`:
the `syscall.SIGHUP` case reloads and lets the loop carry on, and the
`default` case keeps the existing log line and `return`.

## Task 2: two outcomes of a reload

`if err := reload(); err != nil { ... } else { ... }` lets you log
`reload failed: %v; keeping old config` or `reloaded config`.

## Almost the answer

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
