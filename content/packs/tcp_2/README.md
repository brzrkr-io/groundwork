# TCP: banner grabbing with deadlines

When you connect to an SSH or mail server, it greets you with a one-line
*banner* naming its software and version. Security scanners collect banners
to find outdated servers. You'll finish a banner grabber that never hangs,
and fix a stats server that leaves clients hanging.

**New here:** `io.ReadAll` and EOF, `SetReadDeadline`, `net.Error`, `strings.TrimRight`.

## Learn

### Recap

A server gets a `net.Conn` from `ln.Accept()`; a client gets one from
`net.DialTimeout`. A `net.Conn` is a reader and a writer, so `fmt.Fprintf`
writes to it and a `bufio.Reader` reads lines from it. As in *TCP: a
line-based control protocol*, every connection must be closed when you're done.

### Reading until the end

`io.ReadAll(conn)` reads everything the other side sends. It only stops at
*EOF* ("end of file"). On a network connection, EOF means "the other side
closed the connection". If the server never closes, `ReadAll` waits forever
(unless a deadline stops it).

### Read deadlines

`conn.SetReadDeadline(t)` makes any read still waiting at time `t` give up
with an error. Unlike `SetDeadline`, it leaves writes alone. It returns an
error, which you should check.

```go
err := c.SetReadDeadline(time.Now().Add(500 * time.Millisecond))
```

The deadline only helps reads that happen *after* you set it.

### Is it a timeout?

Network errors implement the `net.Error` *interface*, which has a
`Timeout() bool` method. Errors often arrive wrapped (`read banner: i/o
timeout`), so dig one out with `errors.As`: the older form of `errors.AsType`
from *Wrapping errors*, also used in *CSV files: fix the TSV DNS export*.
`errors.As` takes a *pointer* to a variable of the type you want:

```go
var pathErr *fs.PathError
if errors.As(err, &pathErr) && pathErr.Path == "/etc/app.conf" {
	// err is (or wraps) a PathError about that file
}
```

The target type can also be an interface type like `net.Error`; then
`errors.As` finds any error that has the interface's methods.

> **Why is it written like this?** Go requires the `&`: `errors.As` must
> *fill in* your variable, and it can only do that through a pointer. `&&`
> stops early, so the right side only runs when `errors.As` found a match;
> that's what makes it safe to use `pathErr` there. `errors.As` returns
> `false` for a `nil` error, so no separate `nil` check is needed.

### Trimming line endings

Many protocols, including SSH and SMTP, end lines with `"\r\n"` instead of
just `"\n"`. `strings.TrimRight(s, cutset)` removes any mix of the characters
in `cutset` from the end of `s`:

```go
strings.TrimRight("disk full!!?!", "!?") // => "disk full"
```

> **Why is it written like this?** The second argument of `TrimRight` is a
> *set of characters*, not a whole suffix (that's `TrimSuffix`). The order of
> characters inside the quotes doesn't matter.

## Your task

Your security team wants a banner grabber, and the on-call team has a stats
server whose clients keep hanging.

### 1. Fix the stats server

**Goal:** make sure `HandleStats(conn, stats)` always closes `conn` when it returns.

`HandleStats` writes one `name value` line per counter, sorted by name. That
part works. Clients read the reply with `io.ReadAll`, so the reply never ends
until the connection closes, and right now nothing closes it.

```text
errors 3
requests 1200
restarts 1
(connection closed)
```

With an empty `stats` map, the client gets no lines, then the connection closes.

**Think it through:**
- When does `io.ReadAll` stop reading?
- Which Go keyword runs a call when the function returns, however it returns?

### 2. Recognise a timeout

**Goal:** make `IsTimeout(err)` return `true` only when `err` is, or wraps, a `net.Error` whose `Timeout()` returns `true`.

Add the `errors` import.

```go
IsTimeout(os.ErrDeadlineExceeded)  // => true
IsTimeout(fmt.Errorf("read banner: %w", os.ErrDeadlineExceeded)) // => true
IsTimeout(errors.New("disk full")) // => false
IsTimeout(nil)                     // => false
```

(`os.ErrDeadlineExceeded` is what a read returns when its deadline passes.)
A `net.Error` that is *not* a timeout, such as a "no such host" DNS error,
must also give `false`.

**Think it through:**
- What type should the variable you hand to `errors.As` have?
- Finding a `net.Error` isn't enough. What do you still need to ask it?

### 3. Finish the banner grabber

**Goal:** make `ReadBanner(addr, timeout)` return the server's first line, without its line ending, and never hang.

Dialing and reading are already written. Finish it:

- *Before* reading, set a read deadline `timeout` from now (use the `timeout` parameter, not a fixed number).
- Remove the trailing `"\r\n"` or `"\n"` from the banner. Add the `strings` import.
- If anything fails, return `""` and the error.

```go
ReadBanner("10.0.0.5:22", time.Second) // => "SSH-2.0-OpenSSH_9.6", nil  (sent "SSH-2.0-OpenSSH_9.6\r\n")
ReadBanner("10.0.0.7:6379", time.Second) // => "+OK redis ready", nil  (sent "+OK redis ready\n")
// a server that accepts but never writes, with a 100ms timeout:
// => "", an error for which IsTimeout returns true
```

**Think it through:**
- There are two `TODO(3)` markers. Which one goes before the read, and which changes the `return`?
- Which characters must come off the end of the line?

## Learn more

- [net.Conn deadlines](https://pkg.go.dev/net#Conn)
- [net.Error](https://pkg.go.dev/net#Error)
- [errors.As](https://pkg.go.dev/errors#As)
- [strings.TrimRight](https://pkg.go.dev/strings#TrimRight)
