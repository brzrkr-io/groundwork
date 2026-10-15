# Errors are values: parse health checks

A monitoring agent reads its health checks from a small text file, one check
per line. A typo in that file should give an error that points straight at the
bad line, not a silently wrong check.

**New here:** zero-value structs as error results, layered context. (`strings.Fields` is from *The strings package*; `HasPrefix` and `TrimSpace` from *Control flow and functions*.)

## Learn

### Recap

From *Errors are values*: return `error` last; `nil` means success; on failure
return early with zero values; add context with `fmt.Errorf("context: %v", err)`.

### Zero values of structs and slices

The zero value of a struct is the struct with every field at its zero value.
You write it as the type name with empty braces. The zero value of a slice is `nil`.

```go
type Disk struct {
	Name string
	SizeGB int
}

return Disk{}, fmt.Errorf("disk %q not found", name) // Disk{Name: "", SizeGB: 0}
return nil, err                                       // a nil []Disk
```

### Splitting on spaces: strings.Fields

`strings.Fields(s)` splits `s` around runs of spaces and drops the spaces at
the ends. It returns a `[]string`. Use `len` to count the pieces and `[0]`,
`[1]`... to read them.

```go
parts := strings.Fields("  disk1   500  ")
// parts = []string{"disk1", "500"}, len(parts) = 2
```

### Two more string helpers

| Call | Returns | Example |
|---|---|---|
| `strings.HasPrefix(s, p)` | `true` if `s` starts with `p` | `strings.HasPrefix("/var", "/")` → `true` |
| `strings.TrimSpace(s)` | `s` without spaces at both ends | `strings.TrimSpace("  ")` → `""` |

`!` in front of a bool flips it: `!strings.HasPrefix(...)` means "does *not* start with".

### Context stacks up

As an error travels up through your functions, each layer adds what it knows:

```text
line 2: check "api x /healthz": port "x" is not a number
│       │                       └ the problem
│       └ added by ParseCheck
└ added by ParseAll
```

### Skipping and counting in a loop

`for i, line := range lines` gives you the index `i`, starting at **0**.
`continue` jumps straight to the next round of the loop.

```go
for i, name := range hosts {
	if name == "skip-me" {
		continue
	}
	fmt.Printf("host #%d: %s\n", i+1, name) // human counting starts at 1
}
```

> **Why is it written like this?** Go requires slice indexes to start at 0.
> Showing numbers from 1 to people is convention. The names `i` and `line`
> are the author's choice.

## Your task

Each line of the file holds a service name, a port and an HTTP path,
separated by spaces:

```text
api 8080 /healthz
```

Add the imports you need: `fmt`, `strconv` and `strings`.

### 1. Parse one line

**Goal:** `ParseCheck` turns one line into a `Check`, or returns `Check{}` and an error.

The three fields go into `Name`, `Port` (as a number) and `Path`. Check these
in order. In every message, `<line>` is the **whole** input line in quotes:

| Problem | Error message |
|---|---|
| not exactly 3 fields | `check <line>: want 3 fields, got 2` (the real count) |
| port is not a number | `check <line>: port "x" is not a number` |
| port outside 1-65535 | `check <line>: port 0 is out of range 1-65535` |
| path doesn't start with `/` | `check <line>: path "healthz" must start with "/"` |

```go
ParseCheck("api 8080 /healthz")
// => Check{Name: "api", Port: 8080, Path: "/healthz"}, nil
ParseCheck("  web   80  /  ")
// => Check{Name: "web", Port: 80, Path: "/"}, nil
ParseCheck("api 8080 /healthz extra")
// => Check{}, error `check "api 8080 /healthz extra": want 3 fields, got 4`
ParseCheck("api 8080 healthz")
// => Check{}, error `check "api 8080 healthz": path "healthz" must start with "/"`
```

**Think it through:**
- Every message starts with the same `check <line>: ` part. Which parameter fills it?
- In the port messages, is the bad value the text field or the converted number?
- The `"/"` at the end of the last message is fixed text, quotes included.

### 2. Parse the whole file

**Goal:** `ParseAll` parses every non-blank line and returns the checks in order, or `nil` and the first error.

- Skip blank lines: empty, or only spaces.
- Line numbers start at 1 and count **every** line, blank ones too.
- At the first bad line, stop. Return `nil` and `ParseCheck`'s error with `line <n>: ` in front.
- If every line is blank, return no checks and `nil`.

```go
ParseAll([]string{"api 8080 /healthz", "", "web 80 /"})
// => [{api 8080 /healthz} {web 80 /}], nil
ParseAll([]string{"", "api 8080 /healthz", "", "db 5432"})
// => nil, error `line 4: check "db 5432": want 3 fields, got 2`
```

**Think it through:**
- Which helper turns `"   "` into `""` so one comparison catches both kinds of blank line?
- The loop index starts at 0, but line numbers start at 1. What do you add?
- Where do good checks go while you loop? (`append`, as in earlier exercises.)

## Learn more

- [pkg.go.dev: strings.Fields](https://pkg.go.dev/strings#Fields)
- [pkg.go.dev: fmt.Errorf](https://pkg.go.dev/fmt#Errorf)
- [Kubernetes: Configure liveness and readiness probes](https://kubernetes.io/docs/tasks/configure-pod-container/configure-liveness-readiness-startup-probes/)
