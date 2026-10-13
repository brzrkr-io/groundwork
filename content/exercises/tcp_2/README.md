# TCP: banner grabbing with deadlines

When you connect to an SSH or mail server, it greets you with a one-line
*banner* that names the software and version. Security scanners collect
these banners to find outdated servers. Here you finish a banner grabber that
never hangs, and fix a stats server that leaves its clients hanging.

## Learn

### Recap

A server gets a `net.Conn` from `ln.Accept()`; a client gets one from
`net.DialTimeout("tcp", addr, timeout)`. A `net.Conn` is a reader and a
writer, so `fmt.Fprintf(conn, ...)` writes to it and
`bufio.NewReader(conn).ReadString('\n')` reads one line from it. Always
`defer conn.Close()`.

### Reading until the end

`io.ReadAll(conn)` reads everything the other side sends and only stops at
*EOF* ("end of file"), which on a network connection means "the other side
closed the connection". If the server writes its data but never closes,
`ReadAll` waits forever, unless a deadline stops it.

### Read deadlines

`conn.SetReadDeadline(t)` makes any read that is still waiting at time `t`
give up with an error. Unlike `SetDeadline`, it leaves writes alone.

```go
conn.SetReadDeadline(time.Now().Add(500 * time.Millisecond))
```

### Is it a timeout?

Network errors implement the `net.Error` interface, which has a `Timeout()`
method. Errors often arrive wrapped (`read banner: i/o timeout`), so use
`errors.As` to find a `net.Error` inside, as you did with your own error
types. Here the target is an interface variable:

```go
var netErr net.Error
if errors.As(err, &netErr) && netErr.Timeout() {
	// the operation ran out of time
}
```

`errors.As` returns `false` for a `nil` error, so you don't need a separate
`nil` check.

### Trimming line endings

Many network protocols, including SSH and SMTP, end lines with `"\r\n"`
instead of just `"\n"`. `strings.TrimRight(s, "\r\n")` removes any mix of
`\r` and `\n` characters from the end of `s`:

```go
strings.TrimRight("SSH-2.0-OpenSSH_9.6\r\n", "\r\n") // => "SSH-2.0-OpenSSH_9.6"
```

## Your task

Your security team wants a banner grabber, and the on-call team has a stats
server whose clients keep hanging.

### 1. Fix the stats server

`HandleStats(conn, stats)` writes one line per counter, `name value`, sorted
by name, and is then done. Clients read the whole reply with `io.ReadAll`,
but the reply never ends: the function never closes the connection. Fix it so
the connection is always closed when `HandleStats` returns.

```text
errors 3
requests 1200
(connection closed)
```

With an empty `stats` map, the client gets no lines, then the connection
closes.

### 2. Recognise a timeout

`IsTimeout(err)` reports whether `err` is, or wraps, a `net.Error` whose
`Timeout()` method returns `true`. You need to add the `errors` import.

```go
IsTimeout(os.ErrDeadlineExceeded)          // => true
IsTimeout(errors.New("disk full"))         // => false
IsTimeout(nil)                             // => false

wrapped := fmt.Errorf("read banner: %w", os.ErrDeadlineExceeded)
IsTimeout(wrapped)                         // => true
```

(`os.ErrDeadlineExceeded` is the error a read returns when its deadline
passes.)

### 3. Finish the banner grabber

`ReadBanner(addr, timeout)` connects to `addr` and returns the first line the
server sends. The dialing and reading are written. Finish it:

- Before reading, set a read deadline `timeout` from now, so a server that
  never says anything can't hang the grabber.
- Remove the trailing `"\r\n"` (or `"\n"`) from the banner. You need to add
  the `strings` import.

Return `""` and the error if anything fails.

```go
ReadBanner("10.0.0.5:22", time.Second) // => "SSH-2.0-OpenSSH_9.6", nil
ReadBanner("10.0.0.9:25", time.Second) // => "220 mail.example.com ESMTP Postfix", nil
// a server that accepts but never writes, with a 100ms timeout:
// => "", an error for which IsTimeout returns true
```

## Learn more

- [net.Conn deadlines](https://pkg.go.dev/net#Conn)
- [net.Error](https://pkg.go.dev/net#Error)
- [io.ReadAll](https://pkg.go.dev/io#ReadAll)
- [errors.As](https://pkg.go.dev/errors#As)
