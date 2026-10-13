# Port checks: wait for the database

Containers often start before the database they need is ready. A
*wait-for-it* step keeps knocking on the database port and only lets the
migrations run once the port answers. Here you fix and finish one.

## Learn

### Recap

A port check opens a TCP connection and closes it straight away. Code that
dials takes a `DialFunc` instead of calling `net.Dial` directly, so real code
passes the method value `(&net.Dialer{Timeout: 2 * time.Second}).DialContext`
and tests pass a fake. `context.Background()` is the empty starting context
that `DialContext` needs.

### Hosts and ports

An address is `host:port`, but an IPv6 address already contains colons. To
tell them apart, IPv6 hosts go inside square brackets: `[::1]:5432`. The `net`
package handles this for you in both directions:

```go
net.JoinHostPort("db.internal", "5432") // => "db.internal:5432"
net.JoinHostPort("::1", "5432")         // => "[::1]:5432"

host, port, err := net.SplitHostPort("[::1]:5432")
// host == "::1", port == "5432", err == nil
```

Both work with the port as a string. `strconv.Itoa` and `strconv.Atoi` convert
between `int` and `string`. `SplitHostPort` returns an error when the port is
missing or the IPv6 brackets are missing. It does not check that the port is a
number, or that the host is not empty.

A valid TCP port is a whole number from 1 to 65535.

### Injecting sleep

A retry loop waits between attempts. Calling `time.Sleep` directly would make
the tests slow, so `WaitForPort` takes a `sleep func(time.Duration)`
parameter. Real code passes `time.Sleep` itself (a function is a value). The
tests pass a fake that only writes down how long it was asked to sleep.

## Your task

Your deploy pipeline runs `WaitForPort` before database migrations. Its
helpers have a bug and two unfinished pieces.

### 1. Fix `Address` for IPv6

`Address(host, port)` glues the parts together with `":"`, which produces
`::1:5432` for IPv6. Fix it so IPv6 hosts get brackets.

```go
Address("db.internal", 5432) // => "db.internal:5432"
Address("::1", 5432)         // => "[::1]:5432"
```

### 2. Finish `ParseTarget`

`ParseTarget(s)` splits a target like `"db.internal:5432"` into its host and
an `int` port. Return an error (any message) when:

- `net.SplitHostPort` fails, for example `"db.internal"` (no port) or
  `"::1:5432"` (IPv6 without brackets)
- the host is empty, as in `":5432"`
- the port is not a number, as in `"db.internal:http"`
- the port is outside 1–65535, as in `"db.internal:0"` or `"db.internal:65536"`

```go
ParseTarget("db.internal:5432") // => "db.internal", 5432, nil
ParseTarget("[::1]:6379")       // => "::1", 6379, nil
ParseTarget("db.internal:0")    // => "", 0, error
```

### 3. Finish `WaitForPort`

`WaitForPort(dial, addr, attempts, sleep)` dials `addr` (network `"tcp"`) up
to `attempts` times. The loop is already there. Finish it:

- when a dial succeeds, close the connection (already done) and return `nil`
  right away
- after a failed dial, call `sleep(time.Second)`, except after the last
  attempt: there is no point waiting when you won't try again

If every attempt fails, the existing last line returns an error such as:

```text
db:5432 not ready after 3 attempts: connection refused
```

For example, with a port that answers on the third try and `attempts` of 5,
`WaitForPort` dials 3 times, sleeps 1 second twice, and returns `nil`. With a
port that never answers and `attempts` of 3, it dials 3 times and sleeps twice.

## Learn more

- [net.JoinHostPort](https://pkg.go.dev/net#JoinHostPort)
- [net.SplitHostPort](https://pkg.go.dev/net#SplitHostPort)
- [strconv.Atoi](https://pkg.go.dev/strconv#Atoi)
