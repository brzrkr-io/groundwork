# Project: ops inventory CLI

Every ops team keeps a list of its machines somewhere: which hosts exist,
which environment and region they live in, and what they run. In this project
you build `inventory`, a small command-line tool that loads that list from a
JSON file, checks it for mistakes, filters it, and prints it as a table for
people or as JSON for scripts.

It brings the whole module together: flags, stdin, files, exit codes, errors
that point at the exact bad record, and a thin `main`.

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

The inventory file is a JSON array. There is a sample `inventory.json` in the
exercise folder, so the commands above work once you are done:

```json
[
  {"name": "web-1", "env": "prod", "region": "eu-west-1", "address": "10.1.1.11", "tags": ["web", "nginx"]},
  {"name": "build-1", "env": "dev", "region": "us-east-1", "address": "10.9.0.5"}
]
```

### How the code is organised

The data flows through three small packages, and `main` only connects them
to the real operating system:

```text
os.Args, stdin ──► cli.Run ──► inventory.Parse ──► inventory.Validate
                       │                                 │
                       │       inventory.Filter.Apply ◄──┘
                       │                │
                       └──► report.WriteTable / report.WriteJSON ──► stdout
```

| File | What it holds | Tasks |
|---|---|---|
| `cmd/inventory/main.go` | The thin `main`: `os.Exit(cli.Run(os.Args[1:], os.Stdin, os.Stdout, os.Stderr))`. Done. | – |
| `internal/inventory/inventory.go` | `Host`, `Envs`, `Parse`, `Validate`, and the `label` helper. | 1, 2 |
| `internal/inventory/filter.go` | `Filter`, `Match`, and `Apply` (done). | 3 |
| `internal/report/report.go` | `WriteTable` and `WriteJSON`. | 4, 5 |
| `internal/cli/cli.go` | `options`, `parseFlags`, `Run`, plus the finished helpers `usageError` and `load`. | 6, 7 |
| `*_test.go`, `testdata/` | The tests and their fixture files. Read them; don't change them. | – |

Each package has its own tests, so you can finish them in order and watch the
checklist fill up. The suggested order is the task order: the inner packages
first, and the command line last.

### New tools in this project

**`json.Decoder.DisallowUnknownFields`.** By default, JSON keys that match no
struct field are silently ignored. A typo such as `"regoin"` would then leave
`Region` empty with no warning. Turn that into an error instead:

```go
dec := json.NewDecoder(r) // reads JSON straight from any io.Reader
dec.DisallowUnknownFields()
var hosts []Host
err := dec.Decode(&hosts)
```

**`netip.ParseAddr`** (package `net/netip`) checks that a string is an IPv4 or
IPv6 address. You only need its error:

```go
_, err := netip.ParseAddr("10.0.0.300") // err != nil: 300 is too big
```

**`text/tabwriter`** lines up columns. Write cells separated by `\t`, end
each row with `\n`, and call `Flush` at the end. Nothing is printed until
`Flush`, because the tabwriter must see every row before it knows how wide
each column is.

```go
tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0) // pad with 2 spaces
fmt.Fprintln(tw, "NAME\tENV")
fmt.Fprintln(tw, "web-1\tprod")
return tw.Flush()
```

`kubectl get pods` prints its tables this way.

**`json.Encoder.SetIndent("", "  ")`** makes `Encode` pretty-print with two
spaces. `Encode` also adds a final newline.

**`fs.StringVar` and `fs.Func`.** `fs.String` gives you a new pointer.
`fs.StringVar(&opts.file, ...)` stores the value in a variable you already
have, such as a struct field. `fs.Func` calls your function once for each
time the flag appears, which makes a repeatable flag:

```go
fs.Func("tag", "only hosts with this `tag`", func(tag string) error {
	opts.filter.Tags = append(opts.filter.Tags, tag)
	return nil
})
// -tag web -tag nginx  =>  Tags: ["web", "nginx"]
```

`fs.NArg()` is the number of arguments left after the flags, and `fs.Arg(0)`
is the first of them.

### Exit codes

| Code | Meaning | Examples |
|---|---|---|
| `0` | Success, even if no host matched | a table, `-h` |
| `1` | The inventory is the problem | missing file, broken JSON, invalid hosts |
| `2` | The command line is the problem | unknown flag, bad `-env`, extra argument |

A deploy script can then say `inventory -env prod -json > hosts.json && deploy`
and stop safely when anything is wrong.

## Your task

Your team's host list lives in a JSON file that people edit by hand. Build the
tool that everyone (and every script) uses to read it.

### 1. Parse the inventory

In `internal/inventory/inventory.go`, write `Parse(r io.Reader) ([]Host, error)`.
Decode a JSON array of hosts from `r` with a `json.Decoder` and reject unknown
fields. Wrap every error so it starts with `parse inventory: `:

```go
Parse(strings.NewReader(`[{"name": "web-1", "env": "prod"}]`))
// => []Host{{Name: "web-1", Env: "prod"}}, nil

Parse(strings.NewReader(`[{"name": "web-1", "regoin": "eu-west-1"}]`))
// => nil, error "parse inventory: json: unknown field \"regoin\""
```

Broken JSON, a JSON object instead of an array, and empty input are all
errors too.

### 2. Validate every host

Write `Validate(hosts []Host) error`. Check each host against these rules,
in this order, and make one error for each problem:

| Rule | Error text after the label |
|---|---|
| `Name` is not empty | `name is required` |
| `Env` is one of `Envs` (`dev`, `staging`, `prod`) | `env "qa" must be one of dev, staging, prod` |
| `Region` is not empty | `region is required` |
| `Address` is a valid IP (`netip.ParseAddr`) | `address "10.0.0.300" is not a valid IP address` |
| `Name` was not used by an earlier host | `duplicate name, first seen at host 1` |

Each error is `label(n, h) + ": " + text`, where `n` counts hosts from 1. The
`label` helper is already written. It gives `host 3 (web-2)`, or `host 3`
when the name is empty. The starter already loops over the hosts and checks
the region rule; add the other four around it, in the order of the table. Skip the duplicate check for hosts with no name.
Keep going after a problem, so the user sees **every** mistake at once, and
return `errors.Join(errs...)`. That is `nil` when there are no errors, and
otherwise one error whose text has one problem per line:

```text
host 2 (web-2): env "qa" must be one of dev, staging, prod
host 3: name is required
host 3: address "10.1.1.300" is not a valid IP address
```

`strings.Join(Envs, ", ")` builds the `dev, staging, prod` list.

### 3. Match hosts against a filter

In `internal/inventory/filter.go`, write `Match(h Host) bool`. A host matches
when all of these are true:

- `f.Env` is empty, or equals `h.Env`.
- `f.Region` is empty, or equals `h.Region`.
- The host has **every** tag in `f.Tags`.

```go
Filter{Env: "prod", Tags: []string{"web"}}.Match(Host{Env: "prod", Tags: []string{"web", "nginx"}}) // => true
Filter{Tags: []string{"web", "redis"}}.Match(Host{Tags: []string{"web", "nginx"}})                 // => false
Filter{}.Match(anyHost)                                                                             // => true
```

`Apply` already uses `Match` to keep the matching hosts.

### 4. Print a table

In `internal/report/report.go`, write `WriteTable(w io.Writer, hosts []inventory.Host) error`.
Use `tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)`, print the header row
`NAME ENV REGION ADDRESS TAGS`, then one row per host in the order given. Join
tags with `,`, and print `-` for a host with no tags. Return the error from
`Flush`. The expected output is in `internal/report/testdata/table.golden`:

```text
NAME     ENV      REGION     ADDRESS    TAGS
cache-1  staging  eu-west-1  10.1.3.10  cache,redis
db-1     prod     us-east-1  10.0.2.10  db
build-1  dev      us-east-1  10.9.0.5   -
```

With no hosts, print only the header: `NAME  ENV  REGION  ADDRESS  TAGS`.

### 5. Print JSON

Write `WriteJSON(w io.Writer, hosts []inventory.Host) error`. Use a
`json.Encoder` with `SetIndent("", "  ")`. The `Host` struct tags already
give the lower-case key names, and leave out `tags` when a host has none.
A `nil` slice encodes as `null`, but scripts expect a list, so print `[]` for
`nil` or empty input:

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

### 6. Parse the flags

In `internal/cli/cli.go`, finish `parseFlags`. `-file` and `-json` are
already defined. Add:

| Flag | Fills | Default |
|---|---|---|
| `-env` | `opts.filter.Env` | `""` |
| `-region` | `opts.filter.Region` | `""` |
| `-tag` (repeatable, with `fs.Func`) | `opts.filter.Tags`, in the order given | none |

After `fs.Parse`, mistakes are rejected with the `usageError` helper. It
prints `inventory: <message>` and the usage text, then returns the error. The
starter already rejects any leftover argument with
`unexpected argument "web-1"`. Add the second check the same way:

- An `-env` that is set but not in `inventory.Envs`:
  `-env must be one of dev, staging, prod, got "qa"`.

```go
parseFlags([]string{"-env", "prod", "-tag", "web", "-tag", "nginx"}, stderr)
// => options{file: "inventory.json", filter: Filter{Env: "prod", Tags: ["web", "nginx"]}}, nil

parseFlags([]string{"-env", "qa"}, stderr)
// stderr: inventory: -env must be one of dev, staging, prod, got "qa"
//         Usage of inventory: ...
```

Errors from `fs.Parse` itself (an unknown flag, or `-h`) are already printed
by `flag`. Just return them.

### 7. Wire up Run

Finish `Run(args, stdin, stdout, stderr) int`. Steps 1 and 2 are already
written; replace the placeholder after them with steps 3 to 5. In full, `Run`:

1. Calls `parseFlags`. Return `0` if `errors.Is(err, flag.ErrHelp)`, and `2`
   for any other error. Don't print anything more, because `parseFlags`
   already did.
2. Calls `load(opts.file, stdin)`, a finished helper: `-file -` reads
   stdin. On error, print `inventory: <err>` and return `1`.
3. Calls `inventory.Validate`. On error print `inventory: <file> is invalid:`,
   then the error on the next line, and return `1`:

   ```text
   inventory: testdata/invalid.json is invalid:
   host 2 (web-2): env "qa" must be one of dev, staging, prod
   host 3: name is required
   ```

4. Applies `opts.filter`, then sorts the hosts by name with `slices.SortFunc`
   and `cmp.Compare(a.Name, b.Name)`.
5. Prints them with `report.WriteJSON` when `-json` is set, and otherwise with
   `report.WriteTable`. On a write error, print `inventory: <err>` and return
   `1`. Otherwise return `0`. Finding no matches is still a success.

On success nothing goes to stderr, and on failure nothing goes to stdout.

When all seven tests pass, try the real thing from the exercise folder:

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
