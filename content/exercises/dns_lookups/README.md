# DNS: lookups, timeouts and failure types

Before a service can connect anywhere, it turns a name into an address. "The
name doesn't exist", "nobody answered" and "the DNS server is broken" need
different fixes. Here you build a lookup that never hangs and says which one
happened.

**New here:** `net.Resolver` and `LookupHost`, `context.WithTimeout` and `cancel`, `*net.DNSError`.

## Learn

### Three ways a lookup fails

Your program asks a *resolver* for a name's records: an *A* record holds an
IPv4 address, an *AAAA* record an IPv6 one. Go reports every failure as a
`*net.DNSError` and sets a field to say what went wrong:

| Status here | What happened | Field Go sets |
|---|---|---|
| `nxdomain` | The name has no such record. | `IsNotFound` |
| `timeout` | No resolver answered in time. | `IsTimeout` |
| `servfail` | The resolver couldn't get an answer (*SERVFAIL*). | `IsTemporary` |
| `error` | Anything else. | none of them |

Go also sets `IsTemporary` on most timeouts, so check `IsTimeout` first.

### Asking through a `Resolver`

`net.DefaultResolver` is a ready-made `*net.Resolver` that uses your system's
settings. Its `LookupHost` method returns every address as a string:

```go
addrs, err := net.DefaultResolver.LookupHost(ctx, "api.example.com")
// addrs == []string{"203.0.113.10", "2001:db8::10"}, err == nil
```

This pack's `Resolver` interface has just that method, so the tests can pass
a fake, as in *Fakes and dependency injection*.

### A time limit with `context.WithTimeout`

In *Port checks: open, closed or filtered* you passed `context.Background()`.
`context.WithTimeout` makes a context from it that expires after a duration;
anything you pass it to gives up then:

```go
ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
defer cancel()
conn, err := dialer.DialContext(ctx, "tcp", "db.internal:5432")
```

Calling `cancel` ends the context and frees its timer, so always call it.

> **Why is it written like this?** Go requires you to receive both results.
> `defer cancel()` straight after is convention, and `go vet` warns if you
> drop `cancel`. *Context and cancellation* in the *Concurrency* module goes
> deeper.

### Reading a `*net.DNSError`

Find it inside a wrapped error with `errors.AsType`, as in *Wrapping errors*:

```go
if dnsErr, ok := errors.AsType[*net.DNSError](err); ok {
	fmt.Println("asked", dnsErr.Server, "about", dnsErr.Name)
}
```

To tell A from AAAA answers, parse each address with `netip.ParseAddr` and ask
`Is4()`, as in *IP addresses: classify, allowlist, overlap*.

### Checking a status by hand

| Command | What you see | Status |
|---|---|---|
| `dig +short NAME` | the addresses | `ok` |
| `dig NAME` | `status: NXDOMAIN` in the `->>HEADER<<-` line | `nxdomain` |
| `dig NAME` | `status: SERVFAIL` in the header | `servfail` |
| `dig NAME` | no header, then `;; no servers could be reached` | `timeout` |
| `getent hosts NAME` | `203.0.113.10  NAME`, or nothing and exit code 2 | `ok`, or a failure: dig tells you which |

`dig @SERVER NAME` asks one particular resolver; `nameserver` in
`/etc/resolv.conf` says which one you use. Step **1. DNS** of the
*Networking from the ground up* guide covers search domains, `ndots` and more.

## Your task

Your team's monitor resolves each important name every minute and prints a
status a human can act on.

### 1. Classify the lookup error

**Goal:** make `Classify(err)` turn a lookup error into one of the `Status` constants.

Check in this order:

- `OK` when `err` is `nil`
- `Failed` when `err` is not, and doesn't wrap, a `*net.DNSError`
- `NotFound` when its `IsNotFound` is `true`
- `Timeout` when its `IsTimeout` is `true`
- `ServFail` when its `IsTemporary` is `true`
- `Failed` otherwise

Add the `errors` and `net` imports.

```go
Classify(nil)                                       // => OK
Classify(&net.DNSError{IsNotFound: true})           // => NotFound
Classify(fmt.Errorf("check api: %w", timeoutErr))   // => Timeout  (IsTimeout and IsTemporary both true)
Classify(&net.DNSError{IsTemporary: true})          // => ServFail
Classify(errors.New("no route to host"))            // => Failed
```

**Think it through:**
- The constants are declared at the top of the file. Return them; don't redeclare them.
- Once you have the `*net.DNSError`, why must `IsTimeout` come before `IsTemporary`?

### 2. Split A and AAAA answers

**Goal:** make `SplitFamilies(addrs)` return the IPv4 addresses and the IPv6 addresses as two slices.

- Keep each address exactly as given, in its original order.
- Skip anything `netip.ParseAddr` rejects.
- Add the `net/netip` import.

```go
SplitFamilies([]string{"203.0.113.10", "2001:db8::10", "203.0.113.11"})
// => []string{"203.0.113.10", "203.0.113.11"}, []string{"2001:db8::10"}

SplitFamilies([]string{"localhost", "10.0.0.5"})
// => []string{"10.0.0.5"}, nil
```

**Think it through:**
- `v4` and `v6` are named results that start as `nil`. Which one does each address go into?
- What do you do with an address whose parse returned an error?

### 3. Look up with a timeout

**Goal:** make `Lookup(r, host, timeout)` resolve `host` through `r`, giving up after `timeout`.

- Make a context from `context.Background()` that times out after `timeout`, and cancel it when `Lookup` returns.
- Call `r.LookupHost` with that context and `host`.
- Return a `Result` with `host`, the two slices from `SplitFamilies`, and the status from `Classify`.

```go
Lookup(net.DefaultResolver, "api.example.com", 2*time.Second)
// => Result{Host: "api.example.com", IPv4: []string{"203.0.113.10"},
//           IPv6: []string{"2001:db8::10"}, Status: OK}

Lookup(net.DefaultResolver, "nosuch.example.com", 2*time.Second)
// => Result{Host: "nosuch.example.com", Status: NotFound}
```

The tests pass a fake resolver. It checks that its context has a deadline
`timeout` from now, and was cancelled once `Lookup` returned.

**Think it through:**
- Call `r.LookupHost`, the interface method, not `net.DefaultResolver`.
- On an error, the addresses are `nil`. Does `SplitFamilies` need a special case for that?

## Learn more

- [net.Resolver.LookupHost](https://pkg.go.dev/net#Resolver.LookupHost)
- [net.DNSError](https://pkg.go.dev/net#DNSError)
- [context.WithTimeout](https://pkg.go.dev/context#WithTimeout)
- [Kubernetes: Debugging DNS Resolution](https://kubernetes.io/docs/tasks/administer-cluster/dns-debugging-resolution/)
