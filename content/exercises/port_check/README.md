# Port checks: open, closed or filtered

"Is the database port reachable from this box?" is one of the first questions
you ask when something breaks. Here you build a small port checker that tells
an open port from a closed one and from one hidden behind a firewall.

## Learn

### Three answers to "is the port open?"

To check a port you try to open a TCP connection to it, as in *TCP servers and
clients*. What comes back tells you what is on the other side:

| Status | What happened | How fast |
|---|---|---|
| `open` | A program accepted the connection. | Fast |
| `closed` | The host answered "nobody listens here". Go reports *connection refused*. | Fast |
| `timeout` | Nothing answered at all. Usually a firewall silently drops the packets: the port is *filtered*. | As slow as your timeout |
| `error` | Anything else, for example a hostname that doesn't exist. | Varies |

The `timeout` row is why a port checker needs a time limit. Without one, a
single filtered port can make your check hang for minutes.

### Dialing with a timeout

A `net.Dialer` holds settings for opening connections. Its `Timeout` field
caps how long one dial may take:

```go
d := &net.Dialer{Timeout: 2 * time.Second}
conn, err := d.DialContext(context.Background(), "tcp", "10.0.0.5:5432")
```

`DialContext` also takes a *context*. A `context.Context` carries a deadline
and a cancel signal, and `context.Background()` is the empty starting context.
The Concurrency module covers contexts properly; here you always pass
`context.Background()`.

### Methods are values too

You can store a function in a variable or a struct field. A method works the
same way: `d.DialContext`, written without the call parentheses, is a
*method value*. It is a function that remembers `d`, timeout included:

```go
dial := d.DialContext // no () here
conn, err := dial(context.Background(), "tcp", "10.0.0.5:5432")
```

This pack names that function shape `DialFunc`, and `Checker` keeps one in its
`Dial` field. Real code stores `d.DialContext` there. Tests store a fake, so
they can pretend a firewall dropped the packets without needing a firewall.
This is dependency injection, the same trick you used for fakes in the Testing
module.

### Telling the errors apart

A dial error wraps the low-level reason, so you dig with `errors.Is` and
`errors.AsType`:

```go
if errors.Is(err, syscall.ECONNREFUSED) {
	// closed: the host refused the connection
}

if netErr, ok := errors.AsType[net.Error](err); ok && netErr.Timeout() {
	// timeout: nothing answered in time
}
```

`syscall.ECONNREFUSED` is the operating system's "connection refused" error
code. `net.Error` is an interface for network errors with a `Timeout()` method
that reports whether the error was a timeout.

### Lining up columns

In a format string, `%-21s` pads a string with spaces to 21 characters, and
the `-` puts the padding on the right. 21 characters fit the longest IPv4
address with a port, `255.255.255.255:65535`.

```go
fmt.Sprintf("%-6s|", "db")
// => "db    |"
```

## Your task

Before a migration your team wants a quick table of which database and cache
ports answer.

### 1. Classify the dial error

`Classify(err)` returns:

- `Open` when `err` is `nil`
- `Closed` when `err` is, or wraps, `syscall.ECONNREFUSED`
- `Timeout` when `err` is, or wraps, a `net.Error` whose `Timeout()` is `true`
- `Failed` for every other error

```go
Classify(nil)                                    // => Open
Classify(fmt.Errorf("check db: %w", refusedErr)) // => Closed
Classify(errors.New("network is unreachable"))   // => Failed
```

Add the `errors` and `syscall` imports.

### 2. Dial and check

`NewChecker(timeout)` returns a `*Checker` whose `Dial` field is the
`DialContext` method value of a `net.Dialer` with that `Timeout`.

`(c *Checker) Check(addr)` calls `c.Dial` with `context.Background()`,
network `"tcp"` and `addr`. If the connection opened, close it straight away:
you only wanted to know that it opens. Return a `Result` holding `addr` and the
status from `Classify`.

```go
NewChecker(time.Second).Check("127.0.0.1:5432")
// => Result{Addr: "127.0.0.1:5432", Status: Open}   (something listens there)
// => Result{Addr: "127.0.0.1:5432", Status: Closed} (nothing listens there)
```

The tests check a real open port and a real closed port on your own machine.
For the filtered case they put a fake in `Dial` that returns a timeout error.

### 3. Print a report

`Report(results)` returns a header row, one row per result in the order given,
and a summary line. Every row, header included, is the address (or `ADDRESS`)
padded with `%-21s`, one space, then the status (or `STATUS`). Every line ends
with `"\n"`.

```go
Report([]Result{
	{"127.0.0.1:5432", Open},
	{"db.internal:6379", Closed},
	{"10.0.0.9:443", Timeout},
})
```

returns

```text
ADDRESS               STATUS
127.0.0.1:5432        open
db.internal:6379      closed
10.0.0.9:443          timeout
1 of 3 ports open
```

With no results, the report is the header followed by `0 of 0 ports open`.
Build it in a `strings.Builder` with `fmt.Fprintf(&b, ...)`, as in
*Stringers: a kubectl-style deployment report*, rather than with `+=` in a
loop. Add the `fmt` and `strings` imports.

## Learn more

- [net.Dialer](https://pkg.go.dev/net#Dialer)
- [net.Error](https://pkg.go.dev/net#Error)
- [Go spec: Method values](https://go.dev/ref/spec#Method_values)
- [fmt: Printing (widths and flags)](https://pkg.go.dev/fmt#hdr-Printing)
