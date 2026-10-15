# Reading input line by line: a hosts list

*Practice: release day* closed *Git and version control* and the tooling
detour. Back in Go, this module covers reading and writing the text and data
formats that ops tools rely on.

Ops work is full of line-based text: host lists, log files, inventory exports,
command output. Here you read a list of hosts one line at a time, skip the
noise, and report bad lines with their line number.

**New here:** `io.Reader`, `bufio.Scanner`, `sc.Scan`, `sc.Text`, `sc.Err`.

## Learn

### io.Reader: anything you can read bytes from

Go has one small interface for "a source of bytes":

```go
type Reader interface {
	Read(p []byte) (n int, err error)
}
```

Files, network connections, HTTP bodies and plain strings all satisfy it. A
function that takes an `io.Reader` works with all of them. The tests turn a
string into a reader:

```go
r := strings.NewReader("web01\nweb02\n")
```

You almost never call `Read` yourself. You hand the reader to a helper.

> **Why is it written like this?** Taking `io.Reader` instead of a file name
> is the author's choice, and a strong Go convention: it makes the function
> easy to test with a string.

### bufio.Scanner: one line per loop

`bufio.NewScanner` wraps a reader and splits it into lines. Here it prints
each line in upper case:

```go
sc := bufio.NewScanner(r)
for sc.Scan() {
	fmt.Println(strings.ToUpper(sc.Text()))
}
if err := sc.Err(); err != nil {
	return err
}
```

| Call | What it does |
|---|---|
| `sc.Scan()` | Moves to the next line. `false` when input ends **or** reading fails. |
| `sc.Text()` | The current line as a string, without the `\n`. |
| `sc.Err()` | After the loop: `nil` if input simply ended, else the error that stopped it. |

A disk error looks exactly like "end of input" unless you check `sc.Err()`.
Always check it after the loop.

> **Why is it written like this?** `for sc.Scan() { ... }` is Go's
> while-loop: a `for` with only a condition. The scanner needs
> `import "bufio"`. The name `sc` is the author's choice; `scanner` works too.

### Line numbers

The scanner does not count lines. Keep your own counter:

```go
n := 0
for sc.Scan() {
	n++
	// n is 1 on the first line, 2 on the second, ...
}
```

Increase it at the *top* of the loop body, before anything can `continue`
past it. Then every line counts, even skipped ones, and your numbers match
what someone sees in their editor.

### Cleaning up and skipping lines

`strings.TrimSpace` removes spaces and tabs from both ends.
`strings.HasPrefix(s, p)` reports whether `s` starts with `p`. `continue`
jumps straight to the next loop round. Inside the loop, this skips lines
that start with `//`:

```go
text := strings.TrimSpace(sc.Text())
if strings.HasPrefix(text, "//") {
	continue
}
```

| Raw line | After `TrimSpace` |
|---|---|
| `"  web02  "` | `"web02"` |
| `"   "` | `""` (blank) |
| `"  # old host"` | `"# old host"` |

Trim **first**, then check: an indented `#` still counts as a comment.

### Splitting host:port

As in *Strings*, `strings.Cut` splits around the first separator, and
`strconv.Atoi` (from *Type conversion*) turns text into an `int`:

```go
user, domain, found := strings.Cut("ops@example.com", "@")
// user = "ops", domain = "example.com", found = true
n, err := strconv.Atoi("42") // n = 42, err = nil
```

Without the separator, `found` is `false`. `strconv.Atoi("http")` returns an
error.

## Your task

Your monitoring setup reads a hand-edited list of hosts to probe. Write the
functions that read it. Add imports (`bufio`, `strings`, `strconv`, `fmt`) as
you need them.

### 1. Count the lines

**Goal:** `CountLines(r)` returns how many lines `r` has, blank ones
included, and the scanner's error.

```go
CountLines(strings.NewReader("web01\n\nweb02\n")) // => 3, nil
CountLines(strings.NewReader("web01\nweb02"))     // => 2, nil (no final \n still counts)
CountLines(strings.NewReader(""))                 // => 0, nil
```

If reading fails, return the error from `sc.Err()`, directly or wrapped with
`%w` (the test uses `errors.Is`). The count returned with an error isn't
checked.

The tests fake a broken disk with `io.MultiReader`: a normal string, then
`iotest.ErrReader(errDisk)`, a reader that always fails. Your scanner reads
the first line, then hits the error.

**Think it through:**
- What does the loop body need to do once per line?
- After the loop, where does the read error come from?

### 2. Read the hostnames

**Goal:** `ReadHosts(r)` returns the hostnames in order, trimmed, skipping
blank lines and comments.

- Trim spaces and tabs from both ends of each line.
- Skip a line that is empty after trimming.
- Skip a line that starts with `#` after trimming.

```go
ReadHosts(strings.NewReader("# web tier\n  web01  \n\nweb02\n"))
// => []string{"web01", "web02"}, nil
```

If reading fails, return `nil` and the error (directly or wrapped with `%w`).

**Think it through:**
- Trim before you check, or check before you trim?
- Two reasons to skip: which operator joins them in one `if`?
- On a read error, why return `nil` instead of the hosts collected so far?

### 3. Parse host:port targets

**Goal:** `ParseTargets(r)` returns a `Target{Host, Port}` for each kept
line, or stops at the first bad line with an error naming its line number.

Skip blank and comment lines exactly as in task 2. Each kept line has the
form `host:port`:

```go
ParseTargets(strings.NewReader("web01:80\n# db\ndb01:5432\n"))
// => []Target{{Host: "web01", Port: 80}, {Host: "db01", Port: 5432}}, nil
```

On a bad line, return `nil` and an error. The line number is 1-based and
counts **every** line, blank and comment lines included:

| Problem | Error message | Quoted part |
|---|---|---|
| No `:` in the line | `line 4: missing port in "db01.internal"` | the trimmed line |
| Port not a number, or outside 1–65535 | `line 6: invalid port "http"` | the text after `:` |

This input fails with `line 3: invalid port "70000"`:

```text
web01:80

web02:70000
```

Ports `1` and `65535` are valid; `0` is not. If reading fails, return `nil`
and the read error (directly or wrapped with `%w`).

**Think it through:**
- Where does `lineNo++` go so blank and comment lines still count?
- `fmt.Errorf` builds the message: `%d` prints a number, `%q` prints a string
  in double quotes. Which variable fills each spot?
- "Not a number" and "out of range" give the same message. Can one `if`
  cover both?

## Learn more

- [bufio.Scanner](https://pkg.go.dev/bufio#Scanner)
- [io.Reader](https://pkg.go.dev/io#Reader)
- [Go by Example: Line Filters](https://gobyexample.com/line-filters)
- [strings.Cut](https://pkg.go.dev/strings#Cut)
