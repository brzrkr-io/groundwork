# Reading input line by line: scanning logs

A service crashed overnight and you have its log. Here you fix a buggy log
scanner, finish a small `grep`, and make it survive the giant one-line JSON
records real services love to write.

**New here:** the scanner's line-size limit, `sc.Buffer`, `bufio.ErrTooLong`.

## Learn

### Recap

As in *Reading input line by line: a hosts list*:

```go
sc := bufio.NewScanner(r)
n := 0
for sc.Scan() {
	n++              // your own line counter, first thing in the loop
	line := sc.Text()
	_ = line
}
err := sc.Err()      // nil, or the error that stopped the loop
```

### Lines have a size limit

The scanner keeps the current line in a buffer. By default that buffer can
grow to 64 KiB (65,536 bytes). A longer line isn't cut in half. Instead,
`sc.Scan()` returns `false` and `sc.Err()` returns `bufio.ErrTooLong`:

```go
for sc.Scan() {
	// never reached for a 100 KiB line
}
fmt.Println(sc.Err()) // bufio.Scanner: token too long
```

Without a `sc.Err()` check, the scanner silently stops at the first huge
line and reports everything as fine. Lines over 64 KiB are common: a service
that logs a whole request body as one JSON record makes them easily.

### Raising the limit with sc.Buffer

`sc.Buffer(buf, max)` sets the starting buffer and the maximum line size.
This scanner allows lines up to 4 MiB:

```go
sc := bufio.NewScanner(r)
sc.Buffer(make([]byte, 0, 4096), 4*1024*1024)
```

| Argument | Meaning |
|---|---|
| `make([]byte, 0, 4096)` | Starting buffer: empty, with room for 4,096 bytes. |
| `4*1024*1024` | The *maximum* line size, in bytes. 1 KiB = 1024 bytes, 1 MiB = 1024 KiB. |

The scanner grows only when a line needs it. A line over the maximum still
fails with `bufio.ErrTooLong`, so a runaway file can't eat all your memory.

> **Why is it written like this?** Go requires `sc.Buffer` to be called
> **before** the first `sc.Scan()`; calling it later panics. Writing sizes as
> `4*1024*1024` instead of `4194304` is convention: it reads as "4 MiB". The
> starting size is the author's choice.

## Your task

The log for `payments-api` is in front of you. Each line is one log entry.

### 1. Fix FindPanics

**Goal:** `FindPanics(r)` returns the 1-based line numbers of every line
containing `panic:`, in order. It has two bugs to fix.

- Bug 1: the numbers are off by one; the first line is reported as `0`.
- Bug 2: it ignores read errors. If reading fails, return `nil` and the error
  (directly or wrapped with `%w`; the test uses `errors.Is`).

```go
log := "INFO starting\npanic: nil map\nINFO restarting\npanic: index out of range\n"
FindPanics(strings.NewReader(log))
// => []int{2, 4}, nil
```

No panics gives a `nil` slice and a `nil` error.

**Think it through:**
- When the first line is checked, what is `lineNo`? Where should `lineNo++`
  sit?
- What does the function return after the loop? Does it ever look at
  `sc.Err()`?

### 2. Finish Grep

**Goal:** `Grep(r, word)` returns every line containing `word`, formatted as
`"<line number>: <line>"`.

- `<line number>` is 1-based; `<line>` is the line exactly as it is, no
  trimming.
- The match is case-sensitive: `"error"` does not match `ERROR`.
- Fixed text between them: a colon and one space.

```go
log := "INFO ok\nERROR db timeout\nWARN slow\nERROR db timeout\n"
Grep(strings.NewReader(log), "ERROR")
// => []string{"2: ERROR db timeout", "4: ERROR db timeout"}, nil
```

The starter already returns `nil` and the error when reading fails. Add
`import "fmt"` for `fmt.Sprintf`.

**Think it through:**
- `Grep` needs its own counter, like task 1. Where is it declared?
- Which parameter holds the text to look for? (Use the parameter, not a
  fixed string.)
- In `fmt.Sprintf`, which verb prints a number and which prints a string?

### 3. Handle huge lines

**Goal:** make `Grep` accept lines up to 1 MiB with `sc.Buffer`.

The maximum must be exactly 1 MiB (1,048,576 bytes). The tests feed in:

- a log with a 100 KiB JSON line in the middle. `Grep` must read past it and
  find matches after it, and inside it if the word appears there.
- a log with a 2 MiB line. `Grep` must return `nil` and an error that
  `errors.Is(err, bufio.ErrTooLong)` matches.

```go
// log: "INFO start\n" + <100 KiB line> + "\nERROR disk full\n"
Grep(strings.NewReader(log), "ERROR")
// => []string{"3: ERROR disk full"}, nil
```

**Think it through:**
- Where in `Grep` does the call have to go: before or after the loop?
- How do you write 1 MiB the way the Learn example writes 4 MiB?
- For the 2 MiB line, does the `sc.Err()` check already do the work?

## Learn more

- [bufio.Scanner.Buffer](https://pkg.go.dev/bufio#Scanner.Buffer)
- [bufio.ErrTooLong](https://pkg.go.dev/bufio#pkg-variables)
- [Go by Example: Line Filters](https://gobyexample.com/line-filters)
