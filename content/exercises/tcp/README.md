# TCP: a line-based control protocol

Databases, Redis, SSH, and HTTP itself all talk over TCP. Here you build both
ends of a tiny text protocol, so the next time a health check says
"connection refused" you know what happened underneath.

**New here:** `net.Listen`, `net.Conn`, `net.DialTimeout`, deadlines, `strings.CutPrefix`.

## Learn

### What TCP is

*TCP* gives two programs a reliable two-way stream of bytes: what one side
writes, the other reads, in order, nothing lost.

A program is reached at an *address* written `host:port`, such as
`127.0.0.1:6379` (Redis on this machine). The *port* (1–65535) picks one
program on that host.

The *server* listens on a port and waits. The *client* dials the server's
address. Once connected, both sides hold a `net.Conn`.

### The server side

```go
ln, err := net.Listen("tcp", "127.0.0.1:7000") // start listening
conn, err := ln.Accept()                       // wait for one client
```

The starter's `Serve` already does this in a loop. For each client it runs
`go HandleConn(conn)`. The `go` keyword runs `HandleConn` in the background,
so `Serve` can accept the next client at once (the Concurrency module covers
this properly).

### A connection is a reader and a writer

`net.Conn` has `Read` and `Write` methods, so tools you used on files work on
a connection too:

```go
scanner := bufio.NewScanner(conn)
for scanner.Scan() {
	word := scanner.Text()               // one line, without its "\n"
	fmt.Fprintf(conn, "%d\n", len(word)) // write a line back
}
```

`scanner.Scan()` waits for a whole line. It returns `false` when the other
side closes or the connection breaks. The starter's `HandleConn` already
reports broken connections on stderr.

### Closing with `defer`

Every connection must be closed, or the client keeps waiting. `defer` runs a
call when the surrounding function returns, *however* it returns:

```go
func readConfig(f *os.File) {
	defer f.Close() // runs at the end, even after an early return
	// ...
}
```

> **Why is it written like this?** Go requires `defer` to be followed by a
> function call. Putting it right after you get the thing to close is
> convention: readers see the open and the close together.

### The client side

```go
c, err := net.DialTimeout("tcp", "db.internal:5432", 5*time.Second)
if err != nil {
	return err // for example "connection refused"
}
```

`DialTimeout` gives up if connecting takes too long. A slow server could
still keep you waiting forever once connected, so set a *deadline*: a moment
after which reads and writes fail with an error.

```go
c.SetDeadline(time.Now().Add(30 * time.Second)) // 30 seconds from now
```

> **Why is it written like this?** `5*time.Second` is how Go writes a
> duration: a number times a unit. The unit is required in practice; a plain
> `5` means 5 nanoseconds. Whether there are spaces around `*` is decided by
> gofmt, so don't worry about it.

### Reading one reply line

Wrap the connection in a `bufio.Reader`. `ReadString('\n')` reads up to and
including the next newline. Note the single quotes: `'\n'` is one character
(a *rune*), not a string.

```go
r := bufio.NewReader(os.Stdin)
answer, err := r.ReadString('\n')         // "yes\n"
answer = strings.TrimSuffix(answer, "\n") // "yes"
```

### Splitting off a prefix

`strings.CutPrefix(s, prefix)` returns `s` without `prefix`, and `true` if
`s` started with it:

```go
rest, ok := strings.CutPrefix("SET color blue", "SET ") // "color blue", true
rest, ok = strings.CutPrefix("GET color", "SET ")       // "GET color", false
```

## Your task

Your team wants a tiny control protocol for an internal agent. The client
sends one command per line; the server answers each with one line.

| Client sends | Server replies |
|---|---|
| `PING` | `PONG` |
| `ECHO <text>` | `<text>` |
| `QUIT` | `BYE`, then closes the connection |
| anything else | `ERR unknown command` |

- Commands are case-sensitive: `ping` is unknown.
- `ECHO` needs a space after it, so a bare `ECHO` is unknown.
- Every line on the wire, both directions, ends with `"\n"`.

### 1. Answer one command

**Goal:** make `Respond(line)` return the reply from the table and whether to close afterwards.

- The reply has no `"\n"`; the second value is `true` only for `QUIT`.
- The starter handles `PING`. Add the other rows.
- Add the `strings` import.

```go
Respond("PING")           // => "PONG", false
Respond("ECHO deploy ok") // => "deploy ok", false
Respond("QUIT")           // => "BYE", true
Respond("RESTART")        // => "ERR unknown command", false
Respond("ECHO")           // => "ERR unknown command", false
```

**Think it through:**
- Copy the shape of the `PING` check. Which rows are exact matches, and which one starts with fixed text?
- What should the very last `return` give back when nothing matched?

### 2. Serve one connection

**Goal:** make `HandleConn(conn)` answer each line with `Respond` instead of echoing it.

- Write the reply followed by `"\n"`.
- Stop after replying to `QUIT`.
- Close `conn` when `HandleConn` returns, however that happens.

A session (`>` is what the client sends):

```text
> PING
PONG
> ECHO hello
hello
> QUIT
BYE
(connection closed)
```

**Think it through:**
- `Respond` returns two values. Where do you keep each one?
- Which of those two values tells you to leave the loop?

### 3. Send a command as a client

**Goal:** make `SendCommand(addr, cmd)` send `cmd` to the server and return the first reply line.

- Dial with `net.DialTimeout`, timeout 2 seconds.
- Close the connection when done.
- Before writing, set a deadline 2 seconds from now.
- Send `cmd` followed by `"\n"`.
- Return the reply line without its `"\n"`.
- On any error from dialing, the deadline, writing or reading, return `""` and the error.
- Add the `strings` and `time` imports.

```go
SendCommand("127.0.0.1:7000", "PING") // => "PONG", nil
SendCommand("127.0.0.1:1", "PING")    // => "", error (connection refused)
```

The test runs its own small server, so this works even before task 2 is done.
That server replies `got <command>`, so there `SendCommand(addr, "PING")`
returns `"got PING"`.

**Think it through:**
- Which parameter is the address and which is the text to send?
- The steps above are in order. Which ones return an error you must check?
- If the test reports "EOF", did the server ever see the end of your line?

## Learn more

- [net package](https://pkg.go.dev/net)
- [Go by Example: Line filters (bufio.Scanner)](https://gobyexample.com/line-filters)
- [bufio.Reader.ReadString](https://pkg.go.dev/bufio#Reader.ReadString)
- [Go by Example: Defer](https://gobyexample.com/defer)
