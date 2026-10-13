# Variadic functions

Some functions take "as many values as you like", such as `fmt.Println` and
`append`. Ops helpers want the same thing. Join any number of path parts, tag a
resource with any number of labels, find the slowest of any number of
response times.

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

### Inside, it's just a slice

Inside the function, `counts` is an ordinary `[]int`. Everything from
*Slices* and *Looping with range* works on it: `len(counts)`, `counts[0]`,
`for _, c := range counts`.

When the caller passes no values at all, the slice is empty: `len(counts)` is
`0` and the loop body never runs. Watch out for `counts[0]` in that case. It
panics, just like indexing any empty slice.

### Fixed parameters come first

A variadic function can have normal parameters too, but the variadic one must
be the last:

```go
// Deploy describes a deploy of service to one or more regions.
func Deploy(service string, regions ...string) string {
	return service + " to " + strconv.Itoa(len(regions)) + " regions"
}

Deploy("api", "eu-west-1", "us-east-1") // => "api to 2 regions"
Deploy("api")                           // => "api to 0 regions"
```

The first argument fills `service`. Everything after it lands in `regions`.

### Passing a slice you already have

If your values are already in a slice, you can't pass the slice as-is:
`Total(counts)` doesn't compile, because `Total` wants separate `int`s, not
one `[]int`. Put `...` *after* the slice to spread it out:

```go
counts := []int{5, 10, 20}
Total(counts...) // => 35
```

You did the same thing in *append and copy* with `append(a, b...)`. Now you
know why: `append` is variadic too.

Building a slice inside a variadic function works the same way. This makes a
new slice that starts with one value and then holds all the others:

```go
all := append([]string{"first"}, rest...)
```

## Your task

You're writing small helpers for your team's ops toolkit. Each one takes any
number of values.

### 1. Find the slowest response

Change `Max` so it returns the largest of `values`. With no values at all,
return `0`. The values can be negative, so don't start your search from `0`:
start from the first value.

```go
Max(120, 340, 95) // => 340
Max(-5, -2, -9)   // => -2
Max()             // => 0
```

### 2. Join a path

Change `JoinPath` so it joins `parts` with `/` between them. Skip parts that
are the empty string `""`. With no parts (or only empty ones), return `""`.
Don't add a `/` at the start or the end.

```go
JoinPath("var", "log", "nginx")  // => "var/log/nginx"
JoinPath("etc", "", "app.conf")  // => "etc/app.conf"
JoinPath("tmp")                  // => "tmp"
JoinPath()                       // => ""
```

Build the result with `+=`. (*The strings package* shows a shortcut for
this later.)

### 3. Collect resource tags

Change `Tags` so it returns a new slice holding `base` first, followed by
every value in `extra`, in order. With no extras, the slice holds just
`base`.

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
followed by `...`.

## Learn more

- [Go by Example: Variadic Functions](https://gobyexample.com/variadic-functions)
- [Go spec: Passing arguments to ... parameters](https://go.dev/ref/spec#Passing_arguments_to_..._parameters)
- [A Tour of Go: Appending to a slice](https://go.dev/tour/moretypes/15)
