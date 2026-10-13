# Reading input line by line: scanning logs

A service crashed overnight and you have its log. In this exercise you fix a
buggy log scanner, finish a small `grep`, and make it survive the giant
one-line JSON records that real services love to write.

## Learn

### Recap

`bufio.NewScanner(r)` reads any `io.Reader` one line at a time:
`for sc.Scan() { line := sc.Text() }`. When the loop ends, `sc.Err()` says
whether the input really ended (`nil`) or reading failed. Line numbers are your
own counter, increased at the top of the loop so the first line is line 1.

### Lines have a size limit

A scanner keeps the current line in a buffer. By default that buffer can grow
to 64 KiB (65,536 bytes). A longer line doesn't get cut in half. Instead,
`sc.Scan()` returns `false` and `sc.Err()` returns `bufio.ErrTooLong`:

```go
sc := bufio.NewScanner(r)
for sc.Scan() {
	// never reached for a 100 KiB line
}
fmt.Println(sc.Err()) // bufio.Scanner: token too long
```

That is one more reason to always check `sc.Err()`. Without the check, your
scanner silently stops at the first huge line and reports everything as fine.

Log lines over 64 KiB are more common than you'd think: a service that logs a
whole request body or stack trace as one JSON record can easily produce them.

To allow longer lines, call `sc.Buffer` **before** the first `sc.Scan()`:

```go
sc := bufio.NewScanner(r)
sc.Buffer(make([]byte, 0, 64*1024), 1024*1024)
```

| Argument | Meaning |
|---|---|
| `make([]byte, 0, 64*1024)` | The starting buffer: empty, with room for 64 KiB. |
| `1024*1024` | The *maximum* line size: 1 MiB (1,048,576 bytes). |

The scanner starts small and grows only when a line needs it. A line longer
than the maximum still fails with `bufio.ErrTooLong`, so you keep a safety
limit against a runaway file eating all your memory.

## Your task

The log for `payments-api` is in front of you. Each line is one log entry.

### 1. Fix FindPanics

`FindPanics` should return the 1-based line numbers of every line that
contains the text `panic:`, in order. It has two bugs:

- The line numbers are off by one: the first line is reported as `0`.
- It ignores read errors. If reading fails, it must return `nil` and the
  error (directly or wrapped with `%w`; the test uses `errors.Is`).

```go
log := "INFO starting\npanic: nil map\nINFO restarting\npanic: index out of range\n"
FindPanics(strings.NewReader(log))
// => []int{2, 4}, nil
```

No panics gives a `nil` (or empty) slice and a `nil` error.

### 2. Finish Grep

`Grep` returns every line that contains `word`, formatted as
`"<line number>: <line>"`. The match is case-sensitive and the line is kept
exactly as it is (no trimming). Add `import "fmt"` for `fmt.Sprintf`.

```go
log := "INFO ok\nERROR db timeout\nWARN slow\nERROR db timeout\n"
Grep(strings.NewReader(log), "ERROR")
// => []string{"2: ERROR db timeout", "4: ERROR db timeout"}, nil
```

If reading fails, return `nil` and the error, just like task 1. The starter
already does this part.

### 3. Handle huge lines

Make `Grep` accept lines up to 1 MiB by calling `sc.Buffer` exactly as shown
in the Learn section, with a maximum of `1024*1024`. The tests feed in:

- a log with a 100 KiB JSON line in the middle. `Grep` must read past it and
  still find matches after it (and in it, if the word appears there).
- a log with a 2 MiB line. `Grep` must return `nil` and an error that
  `errors.Is(err, bufio.ErrTooLong)` matches. If you returned `sc.Err()` in
  task 2, this works with no extra code.

## Learn more

- [bufio.Scanner.Buffer](https://pkg.go.dev/bufio#Scanner.Buffer)
- [bufio.ErrTooLong](https://pkg.go.dev/bufio#pkg-variables)
- [Go by Example: Line Filters](https://gobyexample.com/line-filters)
