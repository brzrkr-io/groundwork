# Signals: SIGHUP reloads, SIGTERM stops

Not every signal means "stop". Long-running services treat `SIGHUP` as
"re-read your config file", so you can change settings without dropping a
single connection. Here you fix a log-shipping agent that dies when asked to
reload.

## Learn

**Recap.** In *Signals: graceful stop for a rolling drain* you used `signal.Notify` to deliver signals on a
channel, and `for sig := range sigs` to handle each one until the channel
closes. `SIGINT` (Ctrl-C) and `SIGTERM` (`kill`, `docker stop`, Kubernetes)
ask a program to stop, and a program stopped by signal *n* exits with code
`128 + n`.

### SIGHUP means "reload"

`SIGHUP` ("hang up") originally told a program that its terminal had
disconnected. Services that run in the background have no terminal, so they
reuse it to mean **reload your configuration**:

```text
kill -HUP <pid>          # send SIGHUP by hand
nginx -s reload          # nginx sends itself SIGHUP
systemctl reload agent   # often configured to send SIGHUP
```

A good reload is safe. If the new config is broken, the service logs the
problem and **keeps running with the old config**. A typo in a config file
should never take down production.

### Telling signals apart

Signal values can be compared with `==`, so a `switch` on the signal works:

```go
switch sig {
case syscall.SIGHUP:
	// reload
case os.Interrupt, syscall.SIGTERM:
	// stop
}
```

`os.Interrupt` is the same signal as `syscall.SIGINT`. Each signal has a number,
and prints with `%v` as a short name:

| Signal | Number | Prints as | Exit code (128 + n) |
|---|---|---|---|
| `syscall.SIGHUP` | 1 | `hangup` | 129 |
| `os.Interrupt` (`SIGINT`) | 2 | `interrupt` | 130 |
| `syscall.SIGTERM` | 15 | `terminated` | 143 |

A service listens for all three. As in *Signals: graceful stop for a rolling drain*, `go serveLogs()` starts the
real work in the background; that is all you need to know about `go` here:

```go
func main() {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGHUP, os.Interrupt, syscall.SIGTERM)

	go serveLogs()
	os.Exit(reload.Serve(sigs, loadConfig, os.Stderr))
}
```

## Your task

Your log-shipping agent reads its routing rules from a config file. The team
sends it `SIGHUP` after every config change, and every time the agent shuts
down instead of reloading.

### 1. Exit with 128 + n

Make `ExitCodeFor` return the exit code for a stop signal: `130` for
`os.Interrupt` and `143` for `syscall.SIGTERM`. Any other signal returns `1`.

```go
ExitCodeFor(os.Interrupt)    // => 130
ExitCodeFor(syscall.SIGTERM) // => 143
ExitCodeFor(syscall.SIGHUP)  // => 1
```

### 2. Reload on SIGHUP instead of quitting

`Serve` has a bug: it returns on the **first** signal of any kind, so `SIGHUP`
kills the agent. Fix it so that each signal is handled like this:

| Signal | What `Serve` does | Log line |
|---|---|---|
| `syscall.SIGHUP`, reload succeeds | keeps going | `reloaded config` |
| `syscall.SIGHUP`, reload fails | keeps going (old config stays) | `reload failed: <err>; keeping old config` |
| anything else | returns `ExitCodeFor(sig)` | `received <sig>: shutting down` |

Call `reload()` once per `SIGHUP`. Every log line ends with a newline. If the
channel closes, return `0` without logging anything. Example:

```go
// sigs delivers: SIGHUP, SIGHUP, SIGTERM
// the first reload fails with "parse routes.yaml: line 3: bad indent"
code := Serve(sigs, reload, log) // => 143
```

```text
reload failed: parse routes.yaml: line 3: bad indent; keeping old config
reloaded config
received terminated: shutting down
```

## Learn more

- [os/signal package](https://pkg.go.dev/os/signal)
- [nginx: Controlling nginx](https://nginx.org/en/docs/control.html)
- [Go spec: Switch statements](https://go.dev/ref/spec#Switch_statements)
