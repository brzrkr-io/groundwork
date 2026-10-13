# TCP: a line-based control protocol

Databases, Redis, SSH, and HTTP itself all talk over TCP. In this exercise you
build both ends of a tiny text protocol, so the next time a health check says
"connection refused" you know what happened underneath.

## Learn

### What TCP is

*TCP* gives two programs a reliable two-way stream of bytes. Whatever one side
writes, the other side reads, in the same order, with nothing lost. A program
is reached at an *address* written `host:port`, such as `127.0.0.1:6379`
(Redis on this machine) or `db.internal:5432` (Postgres). The *port* is a
number from 1 to 65535 that picks one program on that host.

The *server* listens on a port and waits. The *client* dials the server's
address. Once connected, both sides hold a `net.Conn`.

### The server side

```go
ln, err := net.Listen("tcp", "127.0.0.1:7000") // start listening
conn, err := ln.Accept()                       // wait for one client
```

The starter's `Serve` function already does this in a loop. For each client
it runs `go HandleConn(conn)`. The `go` keyword runs `HandleConn` in the
background so `Serve` can accept the next client straight away; the
Concurrency module covers this properly.

### A connection is a reader and a writer

`net.Conn` has a `Read` and a `Write` method, so it works anywhere an
`io.Reader` or `io.Writer` does. That means tools you already know work on a
network connection:

```go
scanner := bufio.NewScanner(conn) // read the client's lines, as with a file
for scanner.Scan() {
	line := scanner.Text() // the line without its "\n"
	fmt.Fprintf(conn, "you said %s\n", line) // write a line back
}
if err := scanner.Err(); err != nil {
	fmt.Fprintf(os.Stderr, "reading from %s: %v\n", conn.RemoteAddr(), err)
}
```

`scanner.Scan()` waits until a whole line arrives. It returns `false` when
the other side closes the connection, and also when the connection breaks.
`scanner.Err()` tells the two apart, just as with a file. The starter's
`HandleConn` already reports such errors on stderr.

Every connection must be closed when you're done with it, or it stays open
and the client keeps waiting. `defer conn.Close()` at the top of a function
closes it however the function returns.

### The client side

```go
conn, err := net.DialTimeout("tcp", "127.0.0.1:7000", 2*time.Second)
if err != nil {
	return "", err // for example "connection refused"
}
defer conn.Close()
```

`DialTimeout` gives up if connecting takes longer than the timeout. Once
connected, a slow server could still keep you waiting forever, so set a
*deadline*: a point in time after which reads and writes fail with an error.

```go
conn.SetDeadline(time.Now().Add(2 * time.Second))
```

To read one reply line, wrap the connection in a `bufio.Reader`.
`ReadString('\n')` reads up to and including the next `"\n"`:

```go
line, err := bufio.NewReader(conn).ReadString('\n') // "PONG\n"
line = strings.TrimSuffix(line, "\n")                // "PONG"
```

### Splitting off a prefix

`strings.CutPrefix(s, prefix)` returns `s` without `prefix`, and `true` if
`s` started with it:

```go
rest, ok := strings.CutPrefix("ECHO hi there", "ECHO ") // "hi there", true
rest, ok = strings.CutPrefix("PING", "ECHO ")           // "PING", false
```

## Your task

Your team wants a tiny line-based control protocol for an internal agent.
The client sends one command per line; the server answers each one with one
line.

| Client sends | Server replies |
|---|---|
| `PING` | `PONG` |
| `ECHO <text>` | `<text>` |
| `QUIT` | `BYE`, then closes the connection |
| anything else | `ERR unknown command` |

Commands are case-sensitive: `ping` is an unknown command. `ECHO` needs a
space after it, so a bare `ECHO` is unknown too. Every line on the wire, in
both directions, ends with `"\n"`.

### 1. Answer one command

`Respond(line)` returns the reply for one command line (without the `"\n"`)
and whether the connection should close afterwards. Only `QUIT` closes. The
starter handles `PING`; add the other rows of the table. You need to add the
`strings` import.

```go
Respond("PING")           // => "PONG", false
Respond("ECHO deploy ok") // => "deploy ok", false
Respond("QUIT")           // => "BYE", true
Respond("RESTART")        // => "ERR unknown command", false
Respond("ECHO")           // => "ERR unknown command", false
```

### 2. Serve one connection

`HandleConn(conn)` reads the client's lines one by one. Right now it writes
each line straight back. Change it to:

- write the reply from `Respond`, followed by `"\n"`
- stop after replying to `QUIT`
- close `conn` when `HandleConn` returns, however that happens

A session looks like this (`>` is what the client sends):

```text
> PING
PONG
> ECHO hello
hello
> QUIT
BYE
(connection closed)
```

### 3. Send a command as a client

`SendCommand(addr, cmd)` connects to the server at `addr`, sends `cmd`
followed by `"\n"`, and returns the first line of the reply without its
`"\n"`.

- Dial with `net.DialTimeout` and a 2-second timeout.
- Close the connection when you're done.
- Set a deadline 2 seconds from now before you write.
- Return any error from dialing, writing or reading, with `""`.

```go
SendCommand("127.0.0.1:7000", "PING") // => "PONG", nil
SendCommand("127.0.0.1:1", "PING")    // => "", error (connection refused)
```

The test for this task runs its own small server, so it works even if task 2
is not done yet. That server answers every command with `got <command>`, so
`SendCommand(addr, "PING")` returns `"got PING"` there. You need to add the
`strings` and `time` imports.

## Learn more

- [net package](https://pkg.go.dev/net)
- [Go by Example: Line filters (bufio.Scanner)](https://gobyexample.com/line-filters)
- [bufio.Reader.ReadString](https://pkg.go.dev/bufio#Reader.ReadString)
