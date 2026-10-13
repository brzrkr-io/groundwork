# Variadic functions

Some functions take "as many values as you like", such as `fmt.Println` and
`append`. Ops helpers often want the same: join any number of path parts, or
tag a resource with any number of labels.

**New here:** variadic parameters (`...T`), spreading a slice with `slice...`.

## Learn

### Declaring a variadic function

Put `...` in front of the type of the **last** parameter:

```go
// Total adds up any number of request counts.
func Total(counts ...int) int {
	sum := 0
	for _, c := range counts {
		sum += c
	}
	return sum
}
```

A function like this is *variadic*. Callers pass zero, one or many values,
separated by commas:

```go
Total()          // => 0
Total(5)         // => 5
Total(5, 10, 20) // => 35
```

> **Why is it written like this?** Go requires the `...` to sit *before* the
> type (`counts ...int`) and only on the last parameter. The name `counts` is
> your choice; a plural name is convention, because it holds many values.

### Inside, it's just a slice

Inside the function, `counts` is an ordinary `[]int`. Everything from
*Slices* and *Looping with range* works: `len(counts)`, `counts[0]`,
`counts[1:]`, `for _, c := range counts`.

If the caller passes nothing, the slice is empty: `len(counts)` is `0` and a
`range` loop runs zero times. `counts[0]` would *panic* (crash), just like on
any empty slice. So check the length before you index.

### Fixed parameters come first

A variadic function can also have normal parameters. They go first:

```go
func Deploy(service string, regions ...string) string {
	return service + " to " + strconv.Itoa(len(regions)) + " regions"
}

Deploy("api", "eu-west-1", "us-east-1") // => "api to 2 regions"
Deploy("api")                           // => "api to 0 regions"
```

The first argument fills `service`. Everything after it lands in `regions`.

> **Why is it written like this?** Go requires the variadic parameter to be
> last. Otherwise it couldn't tell where "the rest" ends.

### Passing a slice you already have

`Total(counts)` doesn't compile when `counts` is a `[]int`: `Total` wants
separate `int`s, not one slice. Put `...` *after* the slice to spread it out:

```go
counts := []int{5, 10, 20}
Total(counts...) // => 35
```

You did this in *append and copy* with `append(a, b...)`. Now you know why:
`append` is variadic too. Its first argument is a slice, and everything after
that gets added to the end of it.

```go
ports := []int{80}
more := []int{443, 8080}
ports = append(ports, more...) // => []int{80, 443, 8080}
```

> **Why is it written like this?** `...` *before* a type (in the `func` line)
> means "accept many". `...` *after* a value (in a call) means "spread this
> slice out". Both are required syntax; the position tells them apart.

## Your task

You're writing small helpers for your team's ops toolkit. Each one takes any
number of values.

### 1. Find the slowest response

**Goal:** make `Max` return the largest number in `values`, or `0` if there are none.

```go
Max(120, 340, 95) // => 340
Max(-5, -2, -9)   // => -2
Max()             // => 0
```

The values can be negative, so `0` is not a safe starting guess: in the
second example it would beat every value.

**Think it through:**
- What must happen *before* you touch `values[0]`?
- Which value is a safe "largest so far" to start from?
- In the loop, when should "largest so far" change?

### 2. Join a path

**Goal:** make `JoinPath` return the non-empty `parts` joined with `/` between them.

```go
JoinPath("var", "log", "nginx")  // => "var/log/nginx"
JoinPath("etc", "", "app.conf")  // => "etc/app.conf"
JoinPath("", "srv", "data")      // => "srv/data"
JoinPath("tmp")                  // => "tmp"
JoinPath()                       // => ""
```

- Skip parts that are the empty string `""`, wherever they are.
- No `/` at the start or the end.
- No parts, or only empty ones: return `""`.

Build the result with `+=`. (*The strings package* shows a shortcut later.)

**Think it through:**
- How can the loop skip one part and move on to the next?
- A `/` goes *between* two parts. How can you tell that a part is not the
  first one you've added?

### 3. Collect resource tags

**Goal:** make `Tags` return a new slice with `base` first, then every value in `extra`, in order.

```go
Tags("team=platform", "env=prod", "tier=web")
// => []string{"team=platform", "env=prod", "tier=web"}

Tags("team=platform")
// => []string{"team=platform"}

extra := []string{"env=dev", "owner=sam"}
Tags("team=data", extra...)
// => []string{"team=data", "env=dev", "owner=sam"}
```

The tests call `Tags` both ways: with separate values and with a slice
followed by `...`. Either way, `extra` is a `[]string` inside `Tags`.

**Think it through:**
- The starter already returns a slice holding just `base`. What's missing?
- Which built-in function adds the contents of one slice to another?

## Learn more

- [Go by Example: Variadic Functions](https://gobyexample.com/variadic-functions)
- [Go spec: Passing arguments to ... parameters](https://go.dev/ref/spec#Passing_arguments_to_..._parameters)
- [A Tour of Go: Appending to a slice](https://go.dev/tour/moretypes/15)
