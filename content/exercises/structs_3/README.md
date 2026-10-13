# Structs: parse a host inventory

Plenty of ops tooling starts from a plain text inventory: one host per line,
with its name, IP address and region. Here you turn that text into a slice of
structs, then summarise it, the way a deploy script would before a rollout.

**New here:** filling structs from parsed text, a struct with a map field,
adding context to an error.

## Learn

### Recap

A struct groups named fields, as in *Structs*. You build one with a literal
such as `Host{Name: "web-1", IP: "10.0.0.5"}` and reach fields with a dot.

When every field can be compared (strings, numbers, bools), two structs can
be compared with `==`. That's how the tests check your `Host` values.

### Building a slice of structs

Start with an empty slice and `append` one struct per item:

```go
var users []User
users = append(users, User{Login: "ana", Shell: "/bin/bash"})
```

> **Why is it written like this?** `var users []User` gives a `nil` slice,
> and `append` works on it fine. Go requires you to store `append`'s result
> (`users = append(...)`); calling it alone does nothing.

### A struct with a map inside

A map field starts as `nil`, like any map, and writing to a `nil` map
crashes. Make the map first, fill it, then put it in the struct:

```go
perShell := map[string]int{}
perShell["/bin/bash"]++
report := Report{Count: 1, PerShell: perShell}
```

`m[key]++` adds 1 to the value at `key`. A missing key starts at 0, so you
don't need to check first.

### Adding context to an error

When one function calls another that can fail, add *where* it failed to the
message. `fmt.Errorf` builds an error like `fmt.Sprintf` builds a string, and
`%v` prints another error's message:

```go
cfg, err := LoadConfig(path)
if err != nil {
	return fmt.Errorf("config %s: %v", path, err)
}
// => error "config /etc/app.yaml: file not found"
```

> **Why is it written like this?** Returning the error is Go convention: no
> exceptions, just a value you pass up. Prefixing it with context is
> convention too. The exact prefix text is the author's choice.

### Splitting text

| Function | Does |
|---|---|
| `strings.Split(s, "\n")` | cuts `s` at every `\n`, keeps empty pieces |
| `strings.Fields(s)` | cuts at runs of spaces/tabs, drops empty pieces |
| `strings.TrimSpace(s)` | removes spaces/tabs at both ends |
| `strings.HasPrefix(s, p)` | `true` if `s` starts with `p` |

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

**Goal:** make `ParseHost(line)` turn `"name ip region"` into a `Host` and a
`nil` error.

Split with `strings.Fields`, so extra spaces don't matter. A valid line has
exactly 3 fields: name, IP and region, in that order.

With any other number of fields (including an empty line), return the zero
`Host` (`Host{}`) and an error. The error text is up to you; the tests only
check that there is one.

```go
ParseHost("web-1 10.0.0.5 eu-west-1")
// => Host{Name: "web-1", IP: "10.0.0.5", Region: "eu-west-1"}, nil

ParseHost("  db-1   10.0.1.10  us-east-1 ")
// => Host{Name: "db-1", IP: "10.0.1.10", Region: "us-east-1"}, nil

ParseHost("web-2 10.0.0.6")
// => Host{}, error
```

**Think it through:**
- Check the number of fields *before* you read any of them.
- Which index holds the name, the IP, the region?

### 2. Parse the whole inventory

**Goal:** make `ParseInventory(text)` return one `Host` per host line, in
order.

Split `text` into lines with `strings.Split(text, "\n")`, then for each line:

1. Trim spaces with `strings.TrimSpace`.
2. Skip it if it's empty or starts with `#` (a comment).
3. Otherwise parse it with `ParseHost` and append the host to the result.

If the text has no host lines, return no hosts (length 0) and a `nil` error.

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

**Think it through:**
- A `range` loop over the lines gives you an index. Indexes start at 0, but
  line numbers start at 1.
- Which keyword moves on to the next line without parsing this one?
- The error message is the fixed text `line `, the number, `: `, then
  `ParseHost`'s error.

### 3. Summarise the hosts

**Goal:** make `Summarize(hosts)` return a `Summary` with the host count and
the count per region.

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

**Think it through:**
- Do you need a loop for `Total`, or is there a built-in that counts?
- Make the map *before* the loop, so it exists even when `hosts` is empty.

## Learn more

- [strings.Fields](https://pkg.go.dev/strings#Fields)
- [strings.Split](https://pkg.go.dev/strings#Split)
- [A Tour of Go: Structs](https://go.dev/tour/moretypes/2)
- [Ansible: How to build your inventory](https://docs.ansible.com/ansible/latest/inventory_guide/intro_inventory.html)
