# Port checks: wait for the database

Containers often start before their database is ready. A *wait-for-it* step
keeps knocking on the database port and only lets migrations run once it
answers. Here you fix and finish one.

**New here:** `net.JoinHostPort`, `net.SplitHostPort`, `strconv.Atoi`, passing a sleep function.

## Learn

### Recap: dialing through a `DialFunc`

A port check opens a TCP connection and closes it straight away. The code
takes a `DialFunc` parameter instead of calling `net.Dial` itself, so the
tests can pass a fake. `context.Background()` is the empty starting context
the dial needs.

### Hosts and ports

An address is `host:port`. An IPv6 host already contains colons, so it goes
inside square brackets: `[fd00::9]:80`. The `net` package handles this both ways:

```go
net.JoinHostPort("web.local", "80") // => "web.local:80"
net.JoinHostPort("fd00::9", "80")   // => "[fd00::9]:80"

host, port, err := net.SplitHostPort("[fd00::9]:80")
// host == "fd00::9", port == "80", err == nil
```

Both use the port as a `string`.

- `SplitHostPort` returns an error when the port is missing, or an IPv6 host
  has no brackets.
- It does **not** check that the port is a number, or that the host is not empty.
  That is your job.

A valid TCP port is a whole number from 1 to 65535.

### Converting between `int` and `string`

```go
s := strconv.Itoa(8080)       // "8080"
n, err := strconv.Atoi("8080") // 8080, nil
_, err = strconv.Atoi("http")  // err != nil
```

> **Why is it written like this?** Go requires you to receive both values
> from `Atoi`; it returns an error because text might not be a number.
> Checking `if err != nil` right after is convention. The names `n` and `s`
> are the author's choice.

### Passing `sleep` in

A retry loop waits between attempts. Calling `time.Sleep` directly would make
the tests slow, so `WaitForPort` takes a `sleep func(time.Duration)`
parameter. Real code passes `time.Sleep` itself; the tests pass a fake that
only writes down how long it was asked to sleep.

Inside the function you call the **parameter**, `sleep(...)`, not `time.Sleep`.

> **Why is it written like this?** Go allows functions as values, so a
> function can be a parameter. Using this to swap in fakes for tests is a
> convention, not a rule.

## Your task

Your deploy pipeline runs `WaitForPort` before database migrations. Its
helpers have one bug and two unfinished pieces.

### 1. Fix `Address` for IPv6

**Goal:** make `Address(host, port)` put square brackets around IPv6 hosts.

Right now it glues the parts with `":"`, which gives `::1:5432` for IPv6.

```go
Address("db.internal", 5432) // => "db.internal:5432"
Address("::1", 5432)         // => "[::1]:5432"
```

**Think it through:**
- Which `net` function already knows about brackets?
- `port` is an `int`, but that function wants a `string`. What converts it?

### 2. Finish `ParseTarget`

**Goal:** split a target like `"db.internal:5432"` into its host and an `int` port, or return an error.

Return an error (any message) when:

- `net.SplitHostPort` fails: `"db.internal"` (no port), `"::1:5432"` (IPv6 without brackets)
- the host is empty: `":5432"`
- the port is not a number: `"db.internal:http"`
- the port is outside 1–65535: `"db.internal:0"`, `"db.internal:65536"`, `"db.internal:-1"`

On an error, return `""`, `0` and the error.

```go
ParseTarget("db.internal:5432") // => "db.internal", 5432, nil
ParseTarget("[::1]:6379")       // => "::1", 6379, nil
ParseTarget("db.internal:0")    // => "", 0, error
```

**Think it through:**
- Do the checks in the order of the list: split, host, number, range.
- After each step, what do you return if it went wrong?

### 3. Finish `WaitForPort`

**Goal:** return `nil` as soon as a dial works, and sleep one second between failed tries (not after the last one).

`WaitForPort(dial, addr, attempts, sleep)` dials `addr` with network `"tcp"`
up to `attempts` times. The loop and the connection close are already there.
Fill in the two `TODO(3)` lines:

- dial succeeded: return `nil` right away
- dial failed: call `sleep(time.Second)`, unless this was the last attempt

If every attempt fails, the existing last line returns an error such as:

```text
db:5432 not ready after 3 attempts: connection refused
```

Examples:
- Port answers on try 3, `attempts` is 5: 3 dials, 2 sleeps of 1s, returns `nil`.
- Port never answers, `attempts` is 3: 3 dials, 2 sleeps, returns the error.

**Think it through:**
- Which loop variable tells you this is the last attempt?
- Call the `sleep` parameter, not `time.Sleep`.

## Learn more

- [net.JoinHostPort](https://pkg.go.dev/net#JoinHostPort)
- [net.SplitHostPort](https://pkg.go.dev/net#SplitHostPort)
- [strconv.Atoi](https://pkg.go.dev/strconv#Atoi)
