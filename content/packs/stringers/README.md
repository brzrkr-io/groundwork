# Stringers

Alerts, hosts and disks end up in log lines, chat messages and dashboards.
A *Stringer* is a type that knows how to print itself, so `fmt.Println` shows
`ERROR` instead of `3`.

**New here:** the `String() string` method `fmt` looks for (`fmt.Stringer`),
which verbs call `String`, value vs pointer receivers for `String`.

## Learn

### A method fmt looks for

`fmt` has one rule for printing your own types: if a value has a method
named `String` that takes nothing and returns a `string`, `fmt` prints what
that method returns. A type with that method is called a *Stringer*; the
`fmt` docs call it `fmt.Stringer`.

```go
func (r Region) String() string
```

You don't register the method anywhere: having it is enough. *Interfaces*,
in the next module, shows the Go feature behind this.

> **Why is it written like this?** Go requires the exact name `String`, no
> parameters and one `string` result; `ToString()` or `String() []byte`
> won't be found by `fmt`.

### fmt calls String for you

When you print a value, `fmt` checks for a `String` method and prints
whatever it returns:

```go
type Region string

func (r Region) String() string {
	return "region " + string(r)
}

r := Region("eu-west-1")
fmt.Println(r)                     // prints: region eu-west-1
fmt.Sprintf("deploying to %v", r)  // "deploying to region eu-west-1"
fmt.Sprint([]Region{"us-east-1"})  // "[region us-east-1]"
```

`fmt` even calls `String` on each element of a slice. Write the method once;
it's used everywhere.

| Verb or function | Calls `String`? |
|---|---|
| `%v`, `%s` | yes |
| `%q` | yes, and adds quotes |
| `fmt.Print`, `fmt.Println`, `fmt.Sprint` | yes |
| `%d` (on an integer type) | no, prints the raw number |

### Enums print their names

As in *Custom types and enums*, `iota` numbers a list of constants. Without
a `String` method they print as plain numbers:

```go
type Color int

const (
	Red Color = iota // 0
	Green            // 1
)

fmt.Println(Green) // 1, until Color gets a String method
```

A `switch` with one `case` per constant is the simplest `String` method. A
`default` case handles numbers that have no name.

> **Why is it written like this?** Using a `switch` is the author's choice
> (a slice of names also works, as you'll see later). Writing names in
> CAPITALS here is also a choice. The `default` is not required by Go, but
> without it unknown values print as `""`.

### Trap 1: printing the receiver inside String

Inside `String`, never format the receiver itself with `%v` or `%s`:

```go
func (c Color) String() string {
	return fmt.Sprintf("Color(%v)", c) // calls c.String() again… forever
}
```

`%v` calls `String`, which calls `Sprintf`, which calls `String`… until the
program crashes with a *stack overflow*. Fix: convert to the underlying type
first. `int(c)` is a plain `int`, which has no `String` method. `go vet`
warns about this mistake.

### Trap 2: a pointer receiver hides String

As in *Pointer receivers*, a method can have a pointer receiver,
`func (d *Disk) String()`. For `String` there's a catch:

| Method declared on | `fmt.Print(d)` | `fmt.Print(&d)` |
|---|---|---|
| `func (d Disk) String()` | uses `String` | uses `String` |
| `func (d *Disk) String()` | **ignores** `String` | uses `String` |

With `*Disk`, only a pointer has the method. Printing a plain `Disk` falls
back to the default struct format, `{/var 91}`.

### Printing a percent sign

In a format string, `%` starts a verb. To print a real `%`, write `%%`:

```go
fmt.Sprintf("%d%% CPU", 80) // "80% CPU"
```

## Your task

Your alerting tool prints severities, hosts and disks, but the output is a
mess of numbers and curly braces. Teach each type to print itself.

### 1. Name the severities

**Goal:** finish `Severity`'s `String` method so each constant returns its
name in capitals.

| Value | `String()` |
|---|---|
| `Debug` | `"DEBUG"` (already done) |
| `Info` | `"INFO"` |
| `Warning` | `"WARNING"` |
| `Error` | `"ERROR"` |
| `Critical` | `"CRITICAL"` |
| anything else | `"Severity(N)"`, with the number in place of `N` |

```go
Error.String()        // => "ERROR"
Severity(7).String()  // => "Severity(7)"
Severity(-1).String() // => "Severity(-1)"
```

**Think it through:**
- The `case Debug:` lines show the pattern. What do the other four look like?
- In `default`, which verb prints a number? Pass `int(s)` rather than `s`
  (Trap 1): a plain `int` can never call `String` again, whichever verb
  you pick.

### 2. Print a host

**Goal:** add a `String` method to `Host`, with a value receiver, that
returns `name (ip:port)`.

- `Name`, `IP` and `Port` come from the struct's fields.
- Fixed text: one space after the name, `(` and `)` around the rest, and
  `:` between IP and port.

```go
h := Host{Name: "web-1", IP: "10.0.0.5", Port: 8080}
fmt.Sprint(h)
// => "web-1 (10.0.0.5:8080)"

fmt.Sprintf("%v", []Host{h, {Name: "db-1", IP: "10.0.0.9", Port: 5432}})
// => "[web-1 (10.0.0.5:8080) db-1 (10.0.0.9:5432)]"
```

The tests print hosts with `fmt`, so they pass only once `fmt` finds your
method.

**Think it through:**
- Copy the method line shape from `Severity.String`. What changes: the
  receiver, the name, or the result?
- `Port` is an `int` and the others are `string`s. Which verb fits each?

### 3. Fix the disk printout

**Goal:** make `fmt.Sprint(disk)` print `"/var 91% used"` for a plain
`Disk` value.

`Disk` already has a `String` method with the right body, but `fmt` prints
`{/var 91}` instead.

```go
fmt.Sprint(Disk{Mount: "/var", UsedPercent: 91})
// => "/var 91% used"

fmt.Sprint([]Disk{{Mount: "/", UsedPercent: 40}, {Mount: "/data", UsedPercent: 100}})
// => "[/ 40% used /data 100% used]"
```

**Think it through:**
- The body is fine. Look at the receiver in the method's first line.
- Check Trap 2's table: which receiver works for a plain `Disk`?

## Learn more

- [fmt.Stringer](https://pkg.go.dev/fmt#Stringer)
- [A Tour of Go: Stringers](https://go.dev/tour/methods/17)
- [Go by Example: String Formatting](https://gobyexample.com/string-formatting)
