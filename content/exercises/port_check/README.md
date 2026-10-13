# Port checks: open, closed or filtered

"Is the database port reachable from this box?" is one of the first questions
you ask when something breaks. Here you build a port checker that tells an
open port from a closed one and from one hidden behind a firewall.

**New here:** `net.Dialer`, method values, `syscall.ECONNREFUSED`, `%-21s` padding.

## Learn

### Three answers to "is the port open?"

To check a port, you try to open a TCP connection to it, as in *TCP: a
line-based control protocol*. What comes back tells you what's there:

| Status | What happened | How fast |
|---|---|---|
| `open` | A program accepted the connection. | Fast |
| `closed` | The host said "nobody listens here". Go reports *connection refused*. | Fast |
| `timeout` | Nothing answered. Usually a firewall drops the packets: the port is *filtered*. | As slow as your timeout |
| `error` | Anything else, such as a hostname that doesn't exist. | Varies |

The `timeout` row is why a checker needs a time limit: without one, a single
filtered port can hang your check for minutes.

### Dialing with a `net.Dialer`

A `net.Dialer` is a struct of settings for opening connections. Its `Timeout`
field caps how long one dial may take; `KeepAlive` is another field:

```go
d := &net.Dialer{KeepAlive: 30 * time.Second}
conn, err := d.DialContext(context.Background(), "tcp", "10.0.0.5:5432")
```

`DialContext` also takes a *context*, which can carry a deadline or a cancel
signal. `context.Background()` is the empty starting context. The Concurrency
module covers contexts; here you always pass `context.Background()`.

> **Why is it written like this?** `&net.Dialer{Field: value}` is a struct
> literal; `&` gives a pointer to it. Naming the fields is convention (and
> lets you skip the ones you don't need). The name `d` is the author's choice.

### Methods are values too

You can store a function in a variable or struct field. A method works the
same way: write it *without* the call brackets and you get a *method value*,
a function that remembers the value before the dot.

```go
var b strings.Builder
add := b.WriteString // no () here: nothing is written yet
add("hello ")        // same as b.WriteString("hello ")
```

> **Why is it written like this?** Go requires the brackets to be missing:
> with `()` you'd call the method and store its *result* instead. The
> function's shape (its parameters and results) must match the variable's type.

This pack names the dial function shape `DialFunc`, and `Checker` keeps one
in its `Dial` field. Real code stores a dialer's `DialContext` there. Tests
store a fake, so they can pretend a firewall dropped the packets. This is
dependency injection, as with fakes in the Testing module.

### Telling the errors apart

A dial error wraps the low-level reason, so dig with `errors.Is` and
`errors.AsType`, as you did with file errors:

```go
if errors.Is(err, fs.ErrNotExist) {
	// the file isn't there
}
if pathErr, ok := errors.AsType[*fs.PathError](err); ok && pathErr.Op == "open" {
	// opening a file failed
}
```

Two values for this task:

- `syscall.ECONNREFUSED` is the operating system's "connection refused" error.
- `net.Error` is an interface for network errors. Its `Timeout()` method reports whether the error was a timeout.

### Lining up columns

In a format string, `%-21s` pads a string with spaces to 21 characters. The
`-` puts the padding on the right. 21 fits the longest IPv4 address with a
port, `255.255.255.255:65535`.

```go
fmt.Sprintf("%-6s|", "db") // => "db    |"
```

## Your task

Before a migration your team wants a quick table of which database and cache
ports answer.

### 1. Classify the dial error

**Goal:** make `Classify(err)` turn a dial error into one of the `Status` constants.

Check in this order:

- `Open` when `err` is `nil`
- `Closed` when `err` is, or wraps, `syscall.ECONNREFUSED`
- `Timeout` when `err` is, or wraps, a `net.Error` whose `Timeout()` is `true`
- `Failed` for everything else (including a "no such host" DNS error)

Add the `errors` and `syscall` imports.

```go
Classify(nil)                                    // => Open
Classify(fmt.Errorf("check db: %w", refusedErr)) // => Closed
Classify(fmt.Errorf("check web: %w", timeoutErr)) // => Timeout
Classify(errors.New("network is unreachable"))   // => Failed
```

**Think it through:**
- The constants `Open`, `Closed`, `Timeout`, `Failed` are already declared at the top of the file. Return them; don't redeclare them.
- Which check is for one specific error value, and which is for an error *type* you then ask a question?

### 2. Dial and check

**Goal:** make `NewChecker(timeout)` store a real dial function, and `Check(addr)` use it.

`NewChecker(timeout)`:
- Make a `net.Dialer` whose `Timeout` is the `timeout` parameter.
- Return a `*Checker` whose `Dial` field is that dialer's `DialContext` method value.

`(c *Checker) Check(addr)`:
- Call `c.Dial` with `context.Background()`, network `"tcp"` and `addr`.
- If a connection opened, close it straight away: you only wanted to know it opens.
- Return a `Result` with `addr` and the status from `Classify`.

```go
NewChecker(time.Second).Check("127.0.0.1:5432")
// => Result{Addr: "127.0.0.1:5432", Status: Open}   (something listens there)
// => Result{Addr: "127.0.0.1:5432", Status: Closed} (nothing listens there)
```

The tests use a real open port and a real closed port on your machine. For
the filtered case they put a fake in `Dial` that returns a timeout error.

**Think it through:**
- In `NewChecker`, what's the difference between `DialContext` with and without `()`?
- In `Check`, call `c.Dial` (the field), not `net.Dial`. When is the returned connection safe to close?

### 3. Print a report

**Goal:** make `Report(results)` return a table plus a summary line, as one string.

- First line: the header, `ADDRESS` padded with `%-21s`, one space, `STATUS`.
- Then one row per result, in the given order: the address padded with `%-21s`, one space, the status.
- Last line: `N of M ports open`, where N counts `Open` results and M is how many results there are.
- Every line ends with `"\n"`.
- Build it in a `strings.Builder` with `fmt.Fprintf(&b, ...)`, as in *Stringers: a kubectl-style deployment report*. Add the `fmt` and `strings` imports.

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

With no results, it's the header followed by `0 of 0 ports open`. A `Failed`
result prints as `error`.

**Think it through:**
- The header and the rows use the same format. What differs: fixed text or values from `r`?
- Which variable do you need to count open ports as you loop?

## Learn more

- [net.Dialer](https://pkg.go.dev/net#Dialer)
- [net.Error](https://pkg.go.dev/net#Error)
- [Go spec: Method values](https://go.dev/ref/spec#Method_values)
- [fmt: Printing (widths and flags)](https://pkg.go.dev/fmt#hdr-Printing)
