# Signals: SIGHUP reloads, SIGTERM stops

Not every signal means "stop". Services treat `SIGHUP` as "re-read your
config", so you can change settings without dropping a connection. Here you
fix a log-shipping agent that dies when asked to reload.

**New here:** `SIGHUP`, `switch` on a signal.

## Learn

### Recap

In *Signals: graceful stop for a rolling drain* you received signals on a
channel with `for sig := range sigs`. `SIGINT` (Ctrl-C) and `SIGTERM` ask a
program to stop. A program stopped by signal *n* exits with `128 + n`.

### SIGHUP means "reload"

`SIGHUP` ("hang up") once meant "your terminal disconnected". Background
services have no terminal, so they reuse it to mean **reload your config**:

```text
kill -HUP <pid>          # send SIGHUP by hand
systemctl reload agent   # often configured to send SIGHUP
```

A good reload is safe: if the new config is broken, log it and **keep running
with the old config**. A typo should never take down production.

### Telling signals apart

Signals compare with `==`, so `switch` works. One `case` can list several
values, separated by commas:

```go
switch sig {
case syscall.SIGUSR1:
	fmt.Println("dump stats")
case syscall.SIGQUIT, syscall.SIGABRT:
	fmt.Println("crash with a stack trace")
default:
	fmt.Println("something else")
}
```

> **Why is it written like this?** Go requires `case` values to match the
> type after `switch`. Unlike C, Go doesn't fall through to the next case, so
> no `break` is needed. `default` is optional; it runs when no case matches.

| Signal | Number | Prints as (`%v`) | Exit code (128 + n) |
|---|---|---|---|
| `syscall.SIGHUP` | 1 | `hangup` | 129 |
| `os.Interrupt` (`SIGINT`) | 2 | `interrupt` | 130 |
| `syscall.SIGTERM` | 15 | `terminated` | 143 |

The real `main` listens for all three and passes the channel to `Serve`:

```go
signal.Notify(sigs, syscall.SIGHUP, os.Interrupt, syscall.SIGTERM)
go serveLogs()
os.Exit(reload.Serve(sigs, loadConfig, os.Stderr))
```

## Your task

Your agent reads routing rules from a config file. The team sends `SIGHUP`
after every change, and every time the agent shuts down instead.

### 1. Exit with 128 + n

**Goal:** make `ExitCodeFor(sig)` return `130` for `os.Interrupt`, `143` for
`syscall.SIGTERM`, and `1` for any other signal.

```go
ExitCodeFor(os.Interrupt)    // => 130
ExitCodeFor(syscall.SIGTERM) // => 143
ExitCodeFor(syscall.SIGHUP)  // => 1
```

Add `"syscall"` to the imports.

**Think it through:**
- Which value are you comparing: the parameter `sig`.
- Which `switch` part covers "any other signal"?

### 2. Reload on SIGHUP instead of quitting

**Goal:** fix `Serve` so `SIGHUP` reloads and keeps going, and only other
signals shut down.

Right now the loop returns on the first signal of any kind. Handle each
signal like this:

| Signal | What `Serve` does | Log line |
|---|---|---|
| `syscall.SIGHUP`, `reload()` returns `nil` | keeps going | `reloaded config` |
| `syscall.SIGHUP`, `reload()` returns an error | keeps going | `reload failed: <err>; keeping old config` |
| anything else | returns `ExitCodeFor(sig)` | `received <sig>: shutting down` |

Call `reload()` exactly once per `SIGHUP`. Every log line ends with a
newline. If the channel closes, return `0` without logging.

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

**Think it through:**
- The existing two lines in the loop are right for stop signals. Which signal should skip them?
- `reload()` returns an error. Where do you keep it so you can both test and print it?

## Learn more

- [os/signal package](https://pkg.go.dev/os/signal)
- [nginx: Controlling nginx](https://nginx.org/en/docs/control.html)
- [Go spec: Switch statements](https://go.dev/ref/spec#Switch_statements)
