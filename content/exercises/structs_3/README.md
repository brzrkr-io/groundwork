# Structs: parse a host inventory

Plenty of ops tooling starts from a plain text inventory: one host per line,
with its name, IP address and region. In this exercise you turn that text into
a slice of structs, then summarise it, the way a deploy script would before it
decides where to roll out.

## Learn

### Recap

A struct groups named fields, as in *Structs*. You build one with a literal
such as `Host{Name: "web-1", IP: "10.0.0.5"}` and reach fields with a dot.
Structs are copied when you assign them or pass them to a function, as you
saw when fixing the health checker. When every field can be compared
(strings, numbers, bools), two structs can be compared with `==`. That is
how the tests check your `Host` values.

### Building a slice of structs

Start with an empty slice and `append` one struct per item:

```go
var hosts []Host
hosts = append(hosts, Host{Name: "web-1", IP: "10.0.0.5", Region: "eu-west-1"})
```

### A struct with a map inside

A map field starts as `nil`, like any map. Make the map before you write to
it, then put it in the struct:

```go
counts := map[string]int{}
counts["eu-west-1"]++
s := Summary{Total: 1, ByRegion: counts}
```

### Adding context to an error

When one function calls another that can fail, add where it failed to the
message. `%v` prints an error's message:

```go
h, err := ParseHost(line)
if err != nil {
	return nil, fmt.Errorf("line %d: %v", n, err)
}
```

## Your task

The inventory arrives as one string, with `\n` between lines:

```text
# production hosts
web-1 10.0.0.5 eu-west-1
web-2 10.0.0.6 eu-west-1

db-1 10.0.1.10 us-east-1
```

`exercise.go` declares the `Host` and `Summary` types. You need the `fmt` and
`strings` packages, so add them to an `import` block at the top of the file.

### 1. Parse one line

`ParseHost(line)` splits `line` into fields separated by spaces (use
`strings.Fields`, so extra spaces don't matter). A valid line has exactly 3
fields: name, IP and region, in that order. Return the `Host` and a `nil`
error.

With any other number of fields, return the zero `Host` (`Host{}`) and an
error. The error text is up to you; the tests only check that there is one.

```go
ParseHost("web-1 10.0.0.5 eu-west-1")
// => Host{Name: "web-1", IP: "10.0.0.5", Region: "eu-west-1"}, nil

ParseHost("  db-1   10.0.1.10  us-east-1 ")
// => Host{Name: "db-1", IP: "10.0.1.10", Region: "us-east-1"}, nil

ParseHost("web-2 10.0.0.6")
// => Host{}, error
```

### 2. Parse the whole inventory

`ParseInventory(text)` splits `text` into lines with `strings.Split(text,
"\n")` and handles each line in order:

- Trim spaces from the line with `strings.TrimSpace`.
- Skip a line that is empty, or that starts with `#` (a comment).
- Otherwise parse it with `ParseHost` and append the host to the result.

Hosts come back in the same order as in the text. If the text has no host
lines at all, return no hosts (a slice of length 0) and a `nil` error.

If `ParseHost` fails, stop and return `nil` and an error whose message starts
with `line N: `, where `N` is the line's number in the text. The first line is
number 1, and blank lines and comments count too.

```go
ParseInventory("# hosts\nweb-1 10.0.0.5 eu-west-1\n\ndb-1 10.0.1.10 us-east-1\n")
// => []Host{
//     {Name: "web-1", IP: "10.0.0.5", Region: "eu-west-1"},
//     {Name: "db-1", IP: "10.0.1.10", Region: "us-east-1"},
//    }, nil

ParseInventory("web-1 10.0.0.5 eu-west-1\n# db\ndb-1 10.0.1.10\n")
// => nil, error "line 3: ..."
```

### 3. Summarise the hosts

`Summarize(hosts)` returns a `Summary`:

- `Total` is the number of hosts.
- `ByRegion` maps each region to how many hosts are in it.

`ByRegion` is never `nil`. With no hosts it is an empty map.

```go
Summarize([]Host{
	{Name: "web-1", IP: "10.0.0.5", Region: "eu-west-1"},
	{Name: "web-2", IP: "10.0.0.6", Region: "eu-west-1"},
	{Name: "db-1", IP: "10.0.1.10", Region: "us-east-1"},
})
// => Summary{Total: 3, ByRegion: map[string]int{"eu-west-1": 2, "us-east-1": 1}}

Summarize(nil)
// => Summary{Total: 0, ByRegion: map[string]int{}}
```

## Learn more

- [strings.Fields](https://pkg.go.dev/strings#Fields)
- [strings.Split](https://pkg.go.dev/strings#Split)
- [A Tour of Go: Structs](https://go.dev/tour/moretypes/2)
- [Ansible: How to build your inventory](https://docs.ansible.com/ansible/latest/inventory_guide/intro_inventory.html)
