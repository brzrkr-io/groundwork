# Reading input line by line: a hosts list

Ops work is full of line-based text: host lists, log files, inventory exports,
command output. In this exercise you read a list of hosts one line at a time,
skip the noise, and report bad lines with their line number.

## Learn

### io.Reader: anything you can read bytes from

Go has one small interface for "a source of bytes":

```go
type Reader interface {
	Read(p []byte) (n int, err error)
}
```

Files, network connections, HTTP request bodies and plain strings all satisfy
it. Your functions take an `io.Reader`, so they work with all of them. You
will open files and network connections in later exercises. For now the tests
turn a string into a reader:

```go
r := strings.NewReader("web01\nweb02\n")
```

You almost never call `Read` yourself. You hand the reader to something that
does the work for you.

### bufio.Scanner: one line per loop

`bufio.NewScanner` wraps a reader and splits it into lines:

```go
sc := bufio.NewScanner(r)
for sc.Scan() {
	line := sc.Text() // the current line, without the trailing "\n"
	fmt.Println(line)
}
if err := sc.Err(); err != nil {
	return err
}
```

- `sc.Scan()` moves to the next line and returns `true`. It returns `false`
  when the input ends, *or when reading fails*.
- `sc.Text()` returns the current line as a string. The newline is removed.
- `sc.Err()` tells those two cases apart. After the loop it returns `nil` if
  the input simply ended, or the error that stopped the loop. A disk error or a
  dropped connection looks exactly like "end of input" unless you check
  `sc.Err()`. Always check it after the loop.

You need `import "bufio"` at the top of the file to use the scanner.

### Line numbers

The scanner does not count lines for you. Keep your own counter and add one at
the *top* of the loop body, so the first line is line 1:

```go
lineNo := 0
for sc.Scan() {
	lineNo++
	// lineNo is 1 on the first line, 2 on the second, ...
}
```

Count every line, even ones you skip, so your numbers match what someone sees
in their editor.

### Skipping blank lines and comments

Hand-edited files have blank lines, stray spaces and `#` comments.
`strings.TrimSpace` removes spaces and tabs from both ends, which makes the
checks easy:

```go
line := strings.TrimSpace(sc.Text())
if line == "" || strings.HasPrefix(line, "#") {
	continue // skip this line, go to the next one
}
```

| Raw line | After `TrimSpace` | Kept? |
|---|---|---|
| `"web01"` | `"web01"` | yes |
| `"  web02  "` | `"web02"` | yes |
| `""` | `""` | no, blank |
| `"   "` | `""` | no, blank |
| `"  # old host"` | `"# old host"` | no, comment |

### Splitting host:port

As in *Strings*, `strings.Cut` splits a string around the first separator, and
`strconv.Atoi` (from *Type conversion*) turns text into an `int`:

```go
host, port, found := strings.Cut("db01:5432", ":")
// host = "db01", port = "5432", found = true
n, err := strconv.Atoi(port) // n = 5432, err = nil
```

If there is no `:`, `found` is `false`. `strconv.Atoi("http")` returns an
error.

## Your task

Your monitoring setup reads a hand-edited list of hosts to probe. Write the
functions that read it. Add the imports you need (`bufio`, `strings`,
`strconv`, `fmt`) as you go.

### 1. Count the lines

Implement `CountLines`. It returns how many lines the input has, counting
every line, blank ones too.

```go
CountLines(strings.NewReader("web01\n\nweb02\n"))
// => 3, nil
CountLines(strings.NewReader(""))
// => 0, nil
```

A last line without a trailing `\n` still counts. If reading fails, return
the error from `sc.Err()`. Returning it directly or wrapping it with
`fmt.Errorf("...: %w", err)` both pass, because the test uses `errors.Is`.
The tests don't check the count you return alongside an error.

To fake a broken disk, the tests glue two readers together with
`io.MultiReader`: a normal string, then `iotest.ErrReader(errDisk)`, a reader
that always fails with `errDisk`. Your scanner reads the first line fine and
then hits the error.

### 2. Read the hostnames

Implement `ReadHosts`. It returns the hostnames in order, with spaces trimmed
from both ends. It skips blank lines and lines that start with `#` after
trimming.

```go
ReadHosts(strings.NewReader("# web tier\n  web01  \n\nweb02\n"))
// => []string{"web01", "web02"}, nil
```

If reading fails, return `nil` and the error (directly or wrapped with `%w`).

### 3. Parse host:port targets

Implement `ParseTargets`. Each kept line has the form `host:port`. Skip blank
and comment lines exactly as in task 2, and return a `Target` for each kept
line:

```go
ParseTargets(strings.NewReader("web01:80\n# db\ndb01:5432\n"))
// => []Target{{Host: "web01", Port: 80}, {Host: "db01", Port: 5432}}, nil
```

Stop at the first bad line and return `nil` and an error. The error names the
1-based line number, counting **every** line (blank and comment lines
included). There are two kinds of bad line:

| Problem | Error message |
|---|---|
| No `:` in the line | `line 4: missing port in "db01.internal"` |
| Port is not a number, or is outside 1–65535 | `line 6: invalid port "http"` |

The quoted part is the trimmed line for a missing port, and the text after the
`:` for an invalid port. `fmt.Errorf` with `%d` and `%q` produces exactly this
format:

```go
fmt.Errorf("line %d: invalid port %q", lineNo, port)
```

For example, this input fails on line 3 with `line 3: invalid port "70000"`:

```text
web01:80

web02:70000
```

If reading fails, return `nil` and the read error (directly or wrapped with
`%w`).

## Learn more

- [bufio.Scanner](https://pkg.go.dev/bufio#Scanner)
- [io.Reader](https://pkg.go.dev/io#Reader)
- [Go by Example: Line Filters](https://gobyexample.com/line-filters)
- [strings.Cut](https://pkg.go.dev/strings#Cut)
