# Reporting many errors

Before a deploy, a preflight check looks at every target host. If three hosts
each have two problems, you want to see all six at once, not fix one, rerun,
and find the next. Here you combine several errors into one.

**New here:** `errors.Join`, collecting errors in a `[]error`.

## Learn

### One error is not always enough

So far, functions returned at the first problem:

```go
if err := checkCert(s); err != nil {
	return err
}
```

That is right when later steps depend on earlier ones. For independent checks
it hides information: you want to run every check and report every failure.

### errors.Join

`errors.Join` combines any number of errors into one *joined error*:

```go
err := errors.Join(
	errors.New("cert expires in 3 days"),
	errors.New("no backup in 48 hours"),
)
fmt.Println(err)
```

```text
cert expires in 3 days
no backup in 48 hours
```

The message is each error's message, in the order you passed them, with a
newline (`"\n"`) between. There is no newline at the end.

| You pass | `errors.Join` returns |
|---|---|
| some errors and some `nil`s | a joined error; the `nil`s are skipped |
| only `nil`s | `nil` |
| nothing at all | `nil` |

Because `nil`s are skipped, you can pass check results straight in. If every
check passes, you get `nil`, which means success.

### Collecting errors in a slice

When the number of checks isn't fixed, collect failures in a `[]error` and
join them at the end:

```go
var errs []error
for _, user := range users {
	if user.Email == "" {
		errs = append(errs, fmt.Errorf("user %d: no email", user.ID))
	}
}
return errors.Join(errs...) // nil when errs is empty
```

> **Why is it written like this?** The `...` after `errs` is required: it
> spreads a slice into a function that takes any number of arguments, as in
> `append(a, b...)`. Naming the slice `errs` is convention. Wrapping each
> error with context is your choice; use `%w` so the original stays findable.

### errors.Is looks inside

`errors.Is` and `errors.AsType`, from *Wrapping errors*, search every error
inside a joined error, including ones wrapped with `%w`:

```go
err := errors.Join(
	fmt.Errorf("db-1: %w", ErrNoBackup),
	fmt.Errorf("edge: %w", ErrCertExpiring),
)
errors.Is(err, ErrNoBackup) // true
errors.Is(err, ErrDiskFull) // false
```

So callers that only check `err != nil` still work, and callers who care about
one problem can ask for it.

## Your task

The deploy pipeline runs a preflight on every host before a rollout. The
three checks already exist in `exercise.go`. Each returns `nil` when the host
is fine, or an error built on one of the sentinels `ErrUnreachable`,
`ErrLowDisk` or `ErrLowMemory`:

| Check | Fails when | Message |
|---|---|---|
| `checkReachable` | `Reachable` is false | `host unreachable` |
| `checkDisk` | `DiskFreeGB` is below 10 | `low disk space: 3 GB free, need 10` |
| `checkMemory` | `MemoryFreeGB` is below 2 | `low memory: 1 GB free, need 2` |

The `checks` slice lists the three checks in that order.

### 1. Run all checks on one host

**Goal:** make `Preflight(h)` run all three checks and return their errors joined.

Run them in the table's order. Return `nil` when every check passes.

```go
Preflight(Host{Name: "web-1", Reachable: true, DiskFreeGB: 3, MemoryFreeGB: 1})
// => error with the message:
// low disk space: 3 GB free, need 10
// low memory: 1 GB free, need 2

Preflight(Host{Name: "web-2", Reachable: true, DiskFreeGB: 40, MemoryFreeGB: 8})
// => nil
```

**Think it through:**
- What does `errors.Join` do with the `nil` from a check that passed?
- Do you need any `if` statements at all?

### 2. Check every host

**Goal:** make `PreflightAll(hosts)` report every failed check on every host, joined, instead of stopping at the first.

The starter already loops over every host and every check. Change it so:

- each failure becomes one error: the host's `Name`, then `": "`, then the
  check's error, wrapped with `%w` so the sentinel stays visible;
- the order is hosts as given, and for each host, checks in `checks` order;
- it returns `nil` when nothing failed, including when `hosts` is empty.

```go
PreflightAll([]Host{
	{Name: "web-1", Reachable: true, DiskFreeGB: 40, MemoryFreeGB: 8},
	{Name: "web-2", Reachable: true, DiskFreeGB: 3, MemoryFreeGB: 1},
	{Name: "web-3", Reachable: false, DiskFreeGB: 40, MemoryFreeGB: 8},
})
// => error with the message:
// web-2: low disk space: 3 GB free, need 10
// web-2: low memory: 1 GB free, need 2
// web-3: host unreachable
```

**Think it through:**
- Where do you keep failures until the loops finish?
- Which line returns early today, and what should it do instead?

### 3. Decide what to do

**Goal:** make `Decide(err)` turn a preflight result into `"deploy"`, `"abort"` or `"warn"`.

| `err` | `Decide` returns |
|---|---|
| `nil` | `"deploy"` |
| contains `ErrUnreachable` anywhere | `"abort"` |
| any other error | `"warn"` |

```go
Decide(nil)
// => "deploy"

Decide(errors.Join(
	fmt.Errorf("web-2: %w", ErrLowDisk),
	fmt.Errorf("web-3: %w", ErrUnreachable),
))
// => "abort"

Decide(fmt.Errorf("web-2: %w", ErrLowMemory))
// => "warn"
```

**Think it through:**
- Check the rows top to bottom: does the order matter?
- Which function from *Wrapping errors* looks inside joined errors?

## Learn more

- [errors.Join](https://pkg.go.dev/errors#Join)
- [Go 1.20 release notes: wrapping multiple errors](https://go.dev/doc/go1.20#errors)
- [Go blog: Working with errors in Go 1.13](https://go.dev/blog/go1.13-errors)
