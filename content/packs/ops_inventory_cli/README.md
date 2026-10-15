# Project: ops inventory CLI

Every ops team keeps a list of its machines: which hosts exist, which
environment and region they're in, and what they run. You'll build
`inventory`, a command-line tool that loads that list from JSON, checks it
for mistakes, filters it, and prints a table for people or JSON for scripts.

This project closes the *Practical Go* phase. It combines the command line
from *CLI tools and the OS*, the JSON from *Text and data*, and the
`cmd/` + `internal/` layout from *Practice: lay out a tool* into one
multi-package tool.

**New here:** `text/tabwriter`, `netip.ParseAddr`. Everything else is practice from earlier modules.

## Learn

### What you are building

```console
$ go run ./cmd/inventory -env prod -tag web
NAME   ENV   REGION     ADDRESS    TAGS
web-1  prod  eu-west-1  10.1.1.11  web,nginx
web-2  prod  eu-west-1  10.1.1.12  web,nginx
web-3  prod  us-east-1  10.0.1.13  web,nginx

$ go run ./cmd/inventory -region eu-west-1 -json | jq -r '.[].address'
10.1.1.11
...

$ go run ./cmd/inventory -env qa
inventory: -env must be one of dev, staging, prod, got "qa"
Usage of inventory:
...
$ echo $?
2
```

The inventory file is a JSON array. A sample `inventory.json` sits in the
exercise folder:

```json
[
  {"name": "web-1", "env": "prod", "region": "eu-west-1", "address": "10.1.1.11", "tags": ["web", "nginx"]},
  {"name": "build-1", "env": "dev", "region": "us-east-1", "address": "10.9.0.5"}
]
```

### How the code is organised

Data flows through three small packages. `main` only connects them to the
real operating system:

```text
os.Args, stdin ──► cli.Run ──► inventory.Parse ──► inventory.Validate
                       │                                 │
                       │       inventory.Filter.Apply ◄──┘
                       │                │
                       └──► report.WriteTable / report.WriteJSON ──► stdout
```

| File | What it holds | Tasks |
|---|---|---|
| `cmd/inventory/main.go` | The thin `main`, which calls `cli.Run`. Done. | – |
| `internal/inventory/inventory.go` | `Host`, `Envs`, `Parse`, `Validate`, the `label` helper. | 1, 2 |
| `internal/inventory/filter.go` | `Filter`, `Match`, and `Apply` (done). | 3 |
| `internal/report/report.go` | `WriteTable` and `WriteJSON`. | 4, 5 |
| `internal/cli/cli.go` | `options`, `parseFlags`, `Run`, and the finished helpers `usageError` and `load`. | 6, 7 |
| `*_test.go`, `testdata/` | Tests and fixture files. Read them; don't change them. | – |

Each package has its own tests. Work in task order: inner packages first,
the command line last.

> **Why is it written like this?** Go requires the `internal/` directory
> name to mean "only this module may import these packages". `cmd/<name>/`
> for a program's `main` is community convention. Splitting into
> `inventory`, `report` and `cli` is the author's choice.

### Rejecting unknown JSON keys

As in *Decoding JSON: load a service config*, JSON keys that match no struct
field are silently ignored by default. A typo like `"regoin"` would leave
`Region` empty with no warning. `DisallowUnknownFields` makes it an error:

```go
dec := json.NewDecoder(os.Stdin) // reads JSON from any io.Reader
dec.DisallowUnknownFields()      // a setting: call it before Decode
var cfg Config
err := dec.Decode(&cfg)          // & so Decode can fill cfg in
```

### Checking an IP address

`netip.ParseAddr` (package `net/netip`) checks that a string is an IPv4 or
IPv6 address. Here you only need its error, so throw the value away with `_`.
*IP addresses: classify, allowlist, overlap* in the next module does much more
with this package.

```go
_, err := netip.ParseAddr("192.168.1.256") // err != nil: 256 is too big
```

### Lining up columns with `tabwriter`

`text/tabwriter` lines up columns. Write cells separated by `\t`, end each
row with `\n`, and call `Flush` at the end:

```go
tw := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
fmt.Fprintln(tw, "POD\tSTATUS")
fmt.Fprintln(tw, "api-7d9f\tRunning")
tw.Flush()
```

Nothing is printed until `Flush`: the tabwriter must see every row before it
knows how wide each column is. `kubectl get pods` prints tables this way.

> **Why is it written like this?** Go requires all six arguments to
> `NewWriter`. Here they mean: write to `os.Stdout`, no minimum width, no
> tab width, at least `2` spaces between columns, pad with `' '`, no flags.
> `' '` is in single quotes because it's one character (a *rune*), not a string.

### Pretty-printed JSON

As in *Project: access log report*, a `json.Encoder` writes JSON to an
`io.Writer`. `SetIndent("", "  ")` makes `Encode` indent with two spaces;
`Encode` also adds a final newline.

A `nil` slice encodes as `null`, but an empty (non-nil) slice encodes as `[]`.

### Flags that fill a variable you already have

`fs.String` gives you a *new* pointer. `fs.StringVar`, which the starter of
*Config precedence: defaults, file, env and flags* used, stores the value in a
variable you already have, such as a struct field:

```go
fs.StringVar(&cfg.Owner, "owner", "", "team that owns the `service`")
```

`fs.Func`, from *Flags: repeatable and validated flags*, calls your function
once each time the flag appears, which makes a *repeatable* flag:

```go
var ports []string
fs.Func("port", "a `port` to open (repeatable)", func(p string) error {
	ports = append(ports, p)
	return nil
})
// -port 80 -port 443  =>  ports: ["80", "443"]
```

> **Why is it written like this?** The `&` is required: `StringVar` must
> write into your variable, so it needs its address. `fs.Func` requires a
> function with exactly the shape `func(string) error`; returning `nil`
> means "this value is fine". In the help text, backquotes mark the word that
> `flag` shows as the value's name; that's a `flag` package convention.

`fs.NArg()` is the number of arguments left after the flags; `fs.Arg(0)` is
the first of them.

### Exit codes

The same split as in *Standard streams: stdin, stdout, stderr and exit codes*:

| Code | Meaning | Examples |
|---|---|---|
| `0` | Success, even if no host matched | a table, `-h` |
| `1` | The inventory is the problem | missing file, broken JSON, invalid hosts |
| `2` | The command line is the problem | unknown flag, bad `-env`, extra argument |

A deploy script can then run `inventory -env prod -json > hosts.json && deploy`
and stop safely when anything is wrong.

## Your task

Your team's host list lives in a JSON file people edit by hand. Build the
tool everyone (and every script) uses to read it.

### 1. Parse the inventory

**Goal:** in `internal/inventory/inventory.go`, make `Parse(r)` decode a JSON array of hosts from `r` and reject unknown keys.

- Use a `json.Decoder` reading from `r`, and decode into a `[]Host`.
- Every error must start with `parse inventory: `, wrapping the original with `%w`.
- Broken JSON, an object instead of an array, and empty input are all errors.
- Add the `encoding/json` import.

```go
Parse(strings.NewReader(`[{"name": "web-1", "env": "prod"}]`))
// => []Host{{Name: "web-1", Env: "prod"}}, nil

Parse(strings.NewReader(`[{"name": "web-1", "regoin": "eu-west-1"}]`))
// => nil, error "parse inventory: json: unknown field \"regoin\""
```

**Think it through:**
- Which value does the decoder read from: a parameter, or something like `os.Stdin`?
- Where does the decoded list end up, and why does `Decode` need `&` in front of it?

### 2. Validate every host

**Goal:** make `Validate(hosts)` return one combined error listing *every* problem in every host, or `nil` if there are none.

Check each host against these rules, in this order:

| Rule | Error text after the label |
|---|---|
| `Name` is not empty | `name is required` |
| `Env` is one of `Envs` (`dev`, `staging`, `prod`) | `env "qa" must be one of dev, staging, prod` |
| `Region` is not empty (already written) | `region is required` |
| `Address` is a valid IP (`netip.ParseAddr`) | `address "10.0.0.300" is not a valid IP address` |
| `Name` wasn't used by an earlier host | `duplicate name, first seen at host 1` |

- Each error is the label, `": "`, then the text. The starter's `where` variable already holds the label: `host 3 (web-2)`, or `host 3` when the name is empty. Hosts count from 1.
- An empty env is invalid too: `env "" must be one of dev, staging, prod`.
- The number in `first seen at host N` is the host number where that name first appeared.
- Skip the duplicate check for hosts with no name.
- Keep going after a problem. The starter collects errors in `errs` and returns `errors.Join(errs...)`: `nil` when empty, otherwise one problem per line.
- `strings.Join(Envs, ", ")` builds the `dev, staging, prod` list.
- Add the `net/netip`, `slices` and `strings` imports.

```text
host 2 (web-2): env "qa" must be one of dev, staging, prod
host 3: name is required
host 3: address "10.1.1.300" is not a valid IP address
```

**Think it through:**
- Copy the shape of the existing region check. Which two `TODO(2)` spots get which rules?
- In the error text, which parts are fixed words and which come from the host (and need `%q` for quotes)?
- To spot duplicates you must remember each name and where you saw it. Which Go type maps a name to a number, and where must it be declared so it survives the whole loop?

### 3. Match hosts against a filter

**Goal:** in `internal/inventory/filter.go`, make `f.Match(h)` return `true` only when the host passes every part of the filter.

A host matches when all of these are true:

- `f.Env` is empty, or equals `h.Env`.
- `f.Region` is empty, or equals `h.Region`.
- The host has **every** tag in `f.Tags`.

```go
Filter{Env: "prod", Tags: []string{"web"}}.Match(Host{Env: "prod", Tags: []string{"web", "nginx"}}) // => true
Filter{Tags: []string{"web", "redis"}}.Match(Host{Tags: []string{"web", "nginx"}})                 // => false
Filter{}.Match(anyHost)                                                                             // => true
```

`Apply` already uses `Match` to keep the matching hosts.

**Think it through:**
- `f` is the filter and `h` is the host. For each rule, which field of each do you compare?
- Is it easier to look for one reason to say `false`, and say `true` if none was found?

### 4. Print a table

**Goal:** in `internal/report/report.go`, make `WriteTable(w, hosts)` print a header row and one aligned row per host.

- Use `tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)` (writing to the `w` parameter).
- Header: `NAME`, `ENV`, `REGION`, `ADDRESS`, `TAGS`.
- One row per host, in the order given: name, env, region, address, tags.
- Join tags with `,` (no space). A host with no tags shows `-`.
- Return the error from `Flush`.
- Add the `fmt`, `strings` and `text/tabwriter` imports.

The expected output is in `internal/report/testdata/table.golden`:

```text
NAME     ENV      REGION     ADDRESS    TAGS
cache-1  staging  eu-west-1  10.1.3.10  cache,redis
db-1     prod     us-east-1  10.0.2.10  db
build-1  dev      us-east-1  10.9.0.5   -
```

With no hosts, print only the header: `NAME  ENV  REGION  ADDRESS  TAGS`.

**Think it through:**
- Between cells you write `\t`, not spaces. The tabwriter turns them into spaces.
- How do you know a host has no tags, and which value do you print then?

### 5. Print JSON

**Goal:** make `WriteJSON(w, hosts)` print the hosts as a JSON array indented with two spaces.

- Use a `json.Encoder` writing to `w`, with `SetIndent("", "  ")`.
- The `Host` struct tags already give the lower-case keys and leave out `tags` when a host has none.
- For `nil` or empty input, print `[]` (not `null`), followed by a newline.
- Add the `encoding/json` import.

```text
[
  {
    "name": "db-1",
    "env": "prod",
    "region": "us-east-1",
    "address": "10.0.2.10",
    "tags": [
      "db"
    ]
  }
]
```

```go
WriteJSON(w, nil) // writes "[]\n"
```

**Think it through:**
- What does a `nil` slice encode as? What could you swap it for before encoding?

### 6. Parse the flags

**Goal:** in `internal/cli/cli.go`, make `parseFlags` define `-env`, `-region` and a repeatable `-tag`, and reject an unknown env.

`-file` and `-json` are already defined. Add:

| Flag | Fills | Default |
|---|---|---|
| `-env` | `opts.filter.Env` | `""` |
| `-region` | `opts.filter.Region` | `""` |
| `-tag` (repeatable, with `fs.Func`) | `opts.filter.Tags`, in the order given | none |

After `fs.Parse`, the starter rejects a leftover argument with the
`usageError` helper (it prints `inventory: <message>` and the usage text,
then returns the error). Add a second check the same way:

- An `-env` that is set but not in `inventory.Envs` gives
  `-env must be one of dev, staging, prod, got "qa"`.
- Add the `slices` and `strings` imports.

```go
parseFlags([]string{"-env", "prod", "-tag", "web", "-tag", "nginx"}, stderr)
// => options{file: "inventory.json", filter: Filter{Env: "prod", Tags: ["web", "nginx"]}}, nil

parseFlags([]string{"-env", "qa"}, stderr)
// stderr: inventory: -env must be one of dev, staging, prod, got "qa"
//         Usage of inventory: ...
```

Errors from `fs.Parse` itself (an unknown flag, or `-h`) are already printed
by `flag` and returned.

**Think it through:**
- Which flag function stores into a field you already have, and which one makes a flag repeatable?
- Look at the leftover-argument check. What changes for the env check: the condition, the message, the values?
- An empty `-env` is allowed here (it means "any env"). How does your condition let it through?

### 7. Wire up Run

**Goal:** make `Run` validate, filter, sort and print the hosts, returning the right exit code.

Steps 1 and 2 are already written; replace the placeholder after them. In full:

1. `parseFlags`: return `0` for `flag.ErrHelp`, `2` for any other error. (Done.)
2. `load(opts.file, stdin)`: `-file -` reads stdin. On error print `inventory: <err>` and return `1`. (Done.)
3. `inventory.Validate`: on error print `inventory: <file> is invalid:`, then the error on the next line, and return `1`.
4. Keep only hosts that match `opts.filter`, then sort them by name with `slices.SortFunc` and `cmp.Compare`.
5. Print with `report.WriteJSON` when `-json` is set, otherwise `report.WriteTable`. On a write error print `inventory: <err>` and return `1`. Otherwise return `0`; no matches is still success.

- On success nothing goes to stderr; on failure nothing goes to stdout.
- Add the `cmp`, `slices` and `groundwork.local/ops_inventory_cli/internal/report` imports.

```text
inventory: testdata/invalid.json is invalid:
host 2 (web-2): env "qa" must be one of dev, staging, prod
host 3: name is required
```

**Think it through:**
- `<file>` is the path the user gave. Which field of `opts` holds it?
- `slices.SortFunc` needs a function that compares two hosts. Which field of each do you hand to `cmp.Compare`?
- Which variable in `opts` tells you to print JSON?

When all the tests pass, try the real thing from the exercise folder:

```console
go run ./cmd/inventory -env prod
go run ./cmd/inventory -tag db -json
cat inventory.json | go run ./cmd/inventory -file - -region us-east-1
```

## Learn more

- [text/tabwriter](https://pkg.go.dev/text/tabwriter)
- [flag.FlagSet.Func](https://pkg.go.dev/flag#FlagSet.Func)
- [encoding/json: Decoder.DisallowUnknownFields](https://pkg.go.dev/encoding/json#Decoder.DisallowUnknownFields)
- [errors.Join](https://pkg.go.dev/errors#Join)
