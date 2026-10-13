# Reporting many errors

Before a deploy, a preflight check looks at every target host. If three hosts
each have two problems, you want to see all six at once, not fix one, rerun,
and discover the next. In this exercise you combine several errors into one.

## Learn

### One error is not always enough

So far every function returned at the first problem:

```go
if err := checkDisk(h); err != nil {
	return err
}
```

That is right when later steps depend on earlier ones. For checks that are
independent of each other, it hides information. You want to run every check
and report everything that failed.

### errors.Join

`errors.Join` takes any number of errors and combines them into one *joined
error*:

```go
err := errors.Join(
	errors.New("low disk space"),
	errors.New("low memory"),
)
fmt.Println(err)
```

```text
low disk space
low memory
```

Three rules to remember:

| You pass | `errors.Join` returns |
|---|---|
| some errors and some `nil`s | a joined error; the `nil`s are skipped |
| only `nil`s | `nil` |
| nothing at all | `nil` |

The message of a joined error is the messages of its errors, in the order you
passed them, with a newline (`"\n"`) between each one. There is no newline at
the end.

Because `nil` arguments are skipped, you can pass the results of several
checks straight in. If they all succeed, the result is `nil`, which means
success, just like any other error:

```go
return errors.Join(checkA(), checkB())
```

### Collecting errors in a slice

When the number of checks is not fixed, collect the failures in a `[]error`
and join them at the end. The `...` spreads a slice into a variadic call, as
with `append(a, b...)`:

```go
var errs []error
for _, name := range names {
	if name == "" {
		errs = append(errs, errors.New("empty name"))
	}
}
return errors.Join(errs...) // nil when errs is empty
```

You can wrap each error with `%w` before you append it, to add context. The
wrapped error keeps working with `errors.Is`.

### errors.Is looks inside

`errors.Is` and `errors.AsType`, from *Wrapping errors*, search every error
inside a joined error, including errors wrapped with `%w`:

```go
err := errors.Join(
	fmt.Errorf("web-1: %w", ErrLowDisk),
	fmt.Errorf("web-2: %w", ErrUnreachable),
)
errors.Is(err, ErrUnreachable) // true
errors.Is(err, ErrLowMemory)   // false
```

So a joined error is safe to return from any function that returns `error`.
Callers who only check `err != nil` still work, and callers who care about a
particular problem can ask for it.

## Your task

The deploy pipeline runs a preflight on every host before it rolls out a
release. The three checks already exist in `exercise.go`. Each returns `nil`
when the host is fine, or an error wrapping one of the sentinels
`ErrUnreachable`, `ErrLowDisk` and `ErrLowMemory`:

| Check | Fails when | Message |
|---|---|---|
| `checkReachable` | `Reachable` is false | `host unreachable` |
| `checkDisk` | `DiskFreeGB` is below 10 | `low disk space: 3 GB free, need 10` |
| `checkMemory` | `MemoryFreeGB` is below 2 | `low memory: 1 GB free, need 2` |

The `checks` slice lists the three checks in that order.

### 1. Run all checks on one host

Change `Preflight` so it runs all three checks, in the order above, and joins
their errors. It returns `nil` when every check passes.

```go
Preflight(Host{Name: "web-1", Reachable: true, DiskFreeGB: 3, MemoryFreeGB: 1})
// => error with the message:
// low disk space: 3 GB free, need 10
// low memory: 1 GB free, need 2

Preflight(Host{Name: "web-2", Reachable: true, DiskFreeGB: 40, MemoryFreeGB: 8})
// => nil
```

### 2. Check every host

The starter `PreflightAll` already loops over every host and every check in
`checks`, but it returns the first problem it finds. Change it to keep going.
Each failure becomes one error, prefixed with the host name and `": "`. Wrap
with `%w` so the sentinel is still visible. Join all of them: hosts in the
order given, and checks in the order of `checks` for each host. Return `nil`
when nothing failed, including when `hosts` is empty.

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

### 3. Decide what to do

Change `Decide` so it turns a preflight result into a decision:

| `err` | `Decide` returns |
|---|---|
| `nil` | `"deploy"` |
| contains `ErrUnreachable` anywhere (use `errors.Is`) | `"abort"` |
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

## Learn more

- [errors.Join](https://pkg.go.dev/errors#Join)
- [Go 1.20 release notes: wrapping multiple errors](https://go.dev/doc/go1.20#errors)
- [Go blog: Working with errors in Go 1.13](https://go.dev/blog/go1.13-errors)
