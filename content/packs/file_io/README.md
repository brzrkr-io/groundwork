# Files: read, scan and write

Ops tools live on files: host lists, login banners, inventories, generated
configs. Here you read a whole file, read a file line by line, and write a
file safely, using the `os` package.

**New here:** `os.Open`, `os.Create`, `fs.ErrNotExist`, permission bits.

## Learn

### Small files: read it all at once

You used `os.ReadFile` in *Subtests and helpers: files and t.TempDir*. It
returns the whole file as a `[]byte` (raw bytes) and an error:

```go
data, err := os.ReadFile("/etc/hostname")
if err != nil {
	return err
}
name := string(data) // bytes -> string
```

> **Why is it written like this?** Go requires you to handle (or explicitly
> ignore) the second return value. `string(data)` is a *conversion*: Go
> requires it because `[]byte` and `string` are different types.

### A missing file is often not a real error

When a file doesn't exist, the error matches `fs.ErrNotExist` (package
`io/fs`). Check it with `errors.Is`, as in *Wrapping errors*:

```go
cfg, err := os.ReadFile(path)
if errors.Is(err, fs.ErrNotExist) {
	// no file: fall back to something sensible
}
```

Other errors, such as "permission denied" or "is a directory", are real
problems. Return them.

> **Why is it written like this?** Order matters (Go runs `if`s top to
> bottom): check the special case *before* the general `err != nil`, or the
> general one catches it first.

### Big files: open, defer Close, scan

`os.Open` opens a file for reading and returns an `*os.File`. A file is an
`io.Reader`, so the `bufio.Scanner` from *Reading input line by line* works
on it:

```go
f, err := os.Open("/var/log/app.log")
if err != nil {
	return 0, err
}
defer f.Close()

count := 0
sc := bufio.NewScanner(f)
for sc.Scan() {
	count++
}
return count, sc.Err()
```

As in *defer*, `defer f.Close()` runs `f.Close()` when the function returns, on every path.
Each open file uses an OS *file descriptor*, and a process only gets so many.

> **Why is it written like this?** `defer` is Go syntax. Putting it right
> after the error check is convention: you can't forget it, and you never
> close a file that failed to open.

### Skipping lines in a loop

`continue` jumps straight to the next loop round:

```go
if strings.HasPrefix(line, "//") {
	continue // ignore this line
}
```

### Writing: os.Create truncates

`os.Create(path)` opens a file for writing. If it exists, it is
**truncated** (emptied) first; if not, it is created. A file is also an
`io.Writer`, so `fmt.Fprintln(f, x)` writes `x` and a newline to it.

When you *write*, `Close` matters: the OS may only report a failure (like a
full disk) when you close. So return Close's error instead of deferring it:

```go
f, err := os.Create("/tmp/report.txt")
if err != nil {
	return err
}
if _, err := fmt.Fprintln(f, "all good"); err != nil {
	f.Close() // already failing; ignore this second error
	return err
}
return f.Close()
```

### Permission bits

`os.WriteFile(path, data, 0o644)` also writes a whole file. `0o644` is an
octal (base 8) number: one digit each for owner, group and others, adding
read = 4, write = 2, execute = 1.

| Mode | Owner | Group | Others | Typical use |
|---|---|---|---|---|
| `0o644` | read, write | read | read | ordinary files and configs |
| `0o600` | read, write | none | none | secrets, keys, tokens |
| `0o755` | read, write, enter | read, enter | read, enter | directories, scripts |

The process *umask* removes some bits, so `os.Create`'s `0o666` usually ends
up as `0o644`.

### Test files: t.TempDir

The tests use `t.TempDir()`: a fresh empty directory per test, deleted
afterwards. Your code never touches real system files.

## Your task

You're building a small host-maintenance tool: it shows a login banner,
reads the hosts to work on, and writes the list back out. The starter has no
imports; add the ones you need (`bufio`, `errors`, `fmt`, `io/fs`, `os`,
`strings`) as you go.

### 1. Load the login banner

**Goal:** `LoadBanner(path)` returns the whole file as a string, unchanged.

- File missing (or its directory missing): return `DefaultBanner` and a `nil`
  error.
- Any other error (for example, `path` is a directory): return `""` and the
  error.

```go
LoadBanner("banner.txt") // file holds "Maintenance tonight 22:00 UTC\n"
// => "Maintenance tonight 22:00 UTC\n", nil
LoadBanner("missing.txt")
// => DefaultBanner, nil
```

**Think it through:**
- Which `os` function reads a whole file in one call?
- Which check must come first: "missing" or "any error"?

### 2. Read the host list

**Goal:** `ReadHosts(path)` returns one host per line of the file.

- Trim spaces from each line with `strings.TrimSpace`.
- Skip blank lines and lines starting with `#`.
- A last line without `"\n"` still counts.
- If the file can't be opened, return `nil` and the error unchanged, so
  `errors.Is(err, fs.ErrNotExist)` still works for a missing file.
- After the loop, return the scanner's error with the hosts, as the *Learn*
  example does with `sc.Err()`, so a failed read isn't mistaken for a short
  file.

```text
# web tier
web01

  web02
```

```go
ReadHosts("hosts.txt")
// => []string{"web01", "web02"}, nil
```

**Think it through:**
- Open, make sure it gets closed, then scan: which three steps from *Learn*?
- Trim first, then decide: keep the line or `continue`?

### 3. Write the host list

**Goal:** `WriteHosts(path, hosts)` writes each host plus `"\n"` to the file.

- Use `os.Create`, so an existing file is replaced, not appended to.
- An empty list gives an empty file.
- Return any error: from `os.Create` (for example, the parent directory
  doesn't exist), from a write, or from `Close`.

```go
WriteHosts("hosts.txt", []string{"db01", "db02"})
// file now holds "db01\ndb02\n"
```

**Think it through:**
- Which `fmt` function writes a value plus a newline to a writer?
- Why should the last line return `Close`'s error instead of using `defer`?

## Learn more

- [os package](https://pkg.go.dev/os)
- [Go by Example: Reading Files](https://gobyexample.com/reading-files)
- [Go by Example: Writing Files](https://gobyexample.com/writing-files)
- [testing.T.TempDir](https://pkg.go.dev/testing#T.TempDir)
