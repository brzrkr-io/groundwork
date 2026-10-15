# Lookup tables

Some functions are a long `switch` where every case does the same thing
with different values: "for this alert, page that team". The logic is one
line; the rest is data dressed up as code. Move the data into a table (a
map), and the function shrinks to a lookup, a new case becomes one line,
and other code can read the table too.

**New here:** replace conditional with a table, maps of structs, a default
for missing keys.

## Learn

### Spot the table in the switch

```go
switch region {
case "us-east-1":
	return "Virginia"
case "eu-west-1":
	return "Ireland"
default:
	return "unknown"
}
```

Every case has the same shape: a key goes in, a value comes out. That's a
map:

```go
var regionNames = map[string]string{
	"us-east-1": "Virginia",
	"eu-west-1": "Ireland",
}
```

The function becomes a lookup, with the `default` as the fallback:

```go
func RegionName(region string) string {
	if name, ok := regionNames[region]; ok {
		return name
	}
	return "unknown"
}
```

As in *Maps*, `value, ok := m[key]` sets `ok` to `false` when the key
isn't there.

### When each row has several values

If each case returns two things, make the map's values a small struct:

```go
type limits struct {
	cpu    string
	memory string
}

var sizes = map[string]limits{
	"small": {"250m", "256Mi"},
	"large": {"2", "4Gi"},
}
```

Inside a map literal, Go already knows the value type, so each row is just
`{...}` without the type name.

> **Why is it written like this?** **Go requires** `var` for a
> package-level map: maps can't be constants. Leaving out the type inside
> the literal is allowed by **Go** and preferred by **gofmt -s**, the
> simplifier. Aligning the values in a column is **gofmt**'s doing. A
> lower-case type and map keep them private to the package: **your
> choice**.

### Why bother?

| | `switch` | table |
|---|---|---|
| Add a case | 2 lines of code | 1 line of data |
| See every case at once | scroll through `return`s | one aligned block |
| Loop over all cases | impossible | `for key, value := range table` |

The last row is the big one. Code can't ask a `switch` "which teams do you
know?", but it can loop over a map. A table can also move out of the code
later, into a config file, without changing the function.

Not every `switch` should become a table. If the cases do **different
work** (call different functions, check different conditions), keep the
`switch`.

### Maps don't keep an order

Looping over a map visits the keys in an order that can change from run
to run. If the
result must be in order, collect it into a slice and sort it with
`slices.Sort`, as in *Sorting*.

### How the tests check

`structure_test.go` reads `exercise.go` with Go's `go/ast` package and
checks that `Route` and `ExitReason` have no `switch` left and that each
map is declared at package level. The behaviour checks run first.

## Your task

The alert router decides who gets paged. Its two lookups are long
`switch` statements. Turn them into tables.

### 1. Turn the routing switch into a table

**Goal:** declare a `route` struct (a team and a severity) and a
`routes` map from alert name to `route`, and make `Route` look the alert up
in it.

Keep every route exactly as it is now. Alerts that aren't in the map still
go to `"platform"` at `"info"`:

```go
Route("DiskFull")     // => "infra", "critical"
Route("SomethingNew") // => "platform", "info"
```

**Think it through:**
- What goes in as the key, and what two values come out?
- Where does the `default` case go once there is no `switch`?

### 2. Turn the exit code switch into a table

**Goal:** declare an `exitReasons` map from exit code to reason, and make
`ExitReason` look the code up in it.

Unknown codes still return `exit code` and the number:

```go
ExitReason(137) // => "killed (SIGKILL, often out of memory)"
ExitReason(3)   // => "exit code 3"
```

**Think it through:**
- Each case returns one value. What are the map's key and value types?

### 3. List the teams

**Goal:** add a route for `"PodCrashLooping"` (team `"apps"`, severity
`"critical"`), then make `Teams` return every team from the `routes` map
plus `"platform"`, sorted, each once.

```go
Teams() // => ["apps", "data", "infra", "payments", "platform", "security"]
```

Build the list by looping over `routes`, so a new team shows up as soon as
its route is added.

**Think it through:**
- How many lines did the new route take?
- Several routes share a team. How do you add each team only once?
- The loop visits routes in a random order. What makes the result the same every time?

## Learn more

- [Control table (Wikipedia)](https://en.wikipedia.org/wiki/Control_table)
- [Go blog: Go maps in action](https://go.dev/blog/maps)
- [Effective Go: Composite literals](https://go.dev/doc/effective_go#composite_literals)
