# Stringers

Alerts, hosts and disks end up in log lines, chat messages and dashboards.
A *Stringer* is a type that knows how to print itself, so every
`fmt.Println` and `fmt.Sprintf` in your tools shows `ERROR` instead of `3`.

## Learn

### The fmt.Stringer interface

The `fmt` package declares a tiny interface:

```go
type Stringer interface {
	String() string
}
```

As in *Interfaces*, you never say "I implement this". Any type with a method
`String() string` satisfies `fmt.Stringer` automatically.

### fmt calls String for you

When you print a value, `fmt` checks whether it has a `String` method. If it
does, `fmt` prints whatever `String` returns:

```go
type Region string

func (r Region) String() string {
	return "region " + string(r)
}

r := Region("eu-west-1")
fmt.Println(r)
// prints: region eu-west-1

fmt.Sprintf("deploying to %v", r)
// => "deploying to region eu-west-1"

fmt.Sprint([]Region{"us-east-1", "eu-west-1"})
// => "[region us-east-1 region eu-west-1]"
```

The last line shows that `fmt` also calls `String` on every element of a
slice. You write the method once, and it is used everywhere.

Which formatting verbs use `String`?

| Verb or function | Calls `String`? |
|---|---|
| `%v`, `%s` | yes |
| `%q` | yes, and adds quotes |
| `fmt.Print`, `fmt.Println`, `fmt.Sprint` | yes |
| `%d` (on an integer type) | no, prints the raw number |

### Enums print their names

As in *Custom types and enums*, `iota` numbers a list of constants. Without a
`String` method they print as plain numbers:

```go
type Color int

const (
	Red Color = iota // 0
	Green            // 1
)

fmt.Println(Green) // 1, until Color gets a String method
```

A `switch` with one `case` per constant is the simplest `String` method. Add
a `default` for values that have no name. Note the `int(c)` below:

```go
func (c Color) String() string {
	switch c {
	case Red:
		return "RED"
	case Green:
		return "GREEN"
	default:
		return fmt.Sprintf("Color(%d)", int(c))
	}
}
```

### Trap 1: printing the receiver inside String

Inside `String`, never format the receiver itself with `%v` or `%s`:

```go
func (c Color) String() string {
	return fmt.Sprintf("Color(%v)", c) // calls c.String() again… forever
}
```

`%v` calls `String`, which calls `Sprintf`, which calls `String`, until the
program crashes with a *stack overflow*. Convert to the underlying type first
(`int(c)`), or print individual fields. `go vet` warns about this mistake.

### Trap 2: a pointer receiver hides String

Methods can have a pointer receiver, `func (d *Disk) String()`. You will
learn when that is useful in *Pointer receivers*. For `String` there is a
catch:

| Method declared on | `fmt.Print(d)` | `fmt.Print(&d)` |
|---|---|---|
| `func (d Disk) String()` | uses `String` | uses `String` |
| `func (d *Disk) String()` | **ignores** `String` | uses `String` |

With a pointer receiver, only `*Disk` has the method, so printing a plain
`Disk` value falls back to the default `{/var 91}` struct format. Give
`String` a value receiver and it works for both.

### Printing a percent sign

In a format string, `%` starts a verb. To print a literal `%`, write `%%`:

```go
fmt.Sprintf("%d%% full", 80) // "80% full"
```

## Your task

Your alerting tool prints severities, hosts and disks. Right now the output
is a mess of numbers and curly braces. Teach each type to print itself.

### 1. Name the severities

`Severity` is an enum with five values. Finish the `String` method in
`exercise.go` so it returns:

| Value | `String()` |
|---|---|
| `Debug` | `"DEBUG"` |
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

### 2. Print a host

Add a `String` method to `Host` (with a value receiver) that returns the name,
a space, and then the IP and port in brackets, separated by a colon:

```go
h := Host{Name: "web-1", IP: "10.0.0.5", Port: 8080}
fmt.Sprint(h)
// => "web-1 (10.0.0.5:8080)"

fmt.Sprintf("%v", []Host{h, {Name: "db-1", IP: "10.0.0.9", Port: 5432}})
// => "[web-1 (10.0.0.5:8080) db-1 (10.0.0.9:5432)]"
```

The tests print hosts with `fmt`, so they only pass once `fmt` can find your
method.

### 3. Fix the disk printout

`Disk` already has a correct `String` method, but `fmt.Sprint(disk)` still
prints `{/var 91}`. Fix the method so printing a plain `Disk` value gives:

```go
fmt.Sprint(Disk{Mount: "/var", UsedPercent: 91})
// => "/var 91% used"

fmt.Sprint([]Disk{{Mount: "/", UsedPercent: 40}, {Mount: "/data", UsedPercent: 100}})
// => "[/ 40% used /data 100% used]"
```

## Learn more

- [fmt.Stringer](https://pkg.go.dev/fmt#Stringer)
- [A Tour of Go: Stringers](https://go.dev/tour/methods/17)
- [Go by Example: String Formatting](https://gobyexample.com/string-formatting)
