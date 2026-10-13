# Files: read, scan and write

Ops tools live on files: host lists, login banners, inventories, generated
configs. In this exercise you read a whole file, read a file line by line, and
write a file safely, using the `os` package.

## Learn

### Small files: os.ReadFile and os.WriteFile

For a small file, read or write it all in one call:

```go
data, err := os.ReadFile("/etc/motd") // data is a []byte
if err != nil {
	return err
}
text := string(data)

err = os.WriteFile("/tmp/notes.txt", []byte("hello\n"), 0o644)
```

`os.WriteFile` creates the file if it is missing and replaces its contents if
it exists. The last argument is the permission for a new file (see below).

### A missing file is often not a real error

When a file doesn't exist, the error matches `fs.ErrNotExist` (from the
`io/fs` package). Check it with `errors.Is`, as in *Wrapping errors*:

```go
data, err := os.ReadFile(path)
if errors.Is(err, fs.ErrNotExist) {
	// no file: use a default
}
```

Any other error, such as "permission denied" or "is a directory", is a real
problem. Return it.

### Big files: open, defer Close, scan

`os.Open` opens a file for reading and returns an `*os.File`. A file is an
`io.Reader`, so the `bufio.Scanner` from *Reading input line by line* works on
it directly:

```go
f, err := os.Open(path)
if err != nil {
	return nil, err
}
defer f.Close() // runs when the function returns, on every path

sc := bufio.NewScanner(f)
for sc.Scan() {
	line := sc.Text()
	// ...
}
return lines, sc.Err()
```

Every file you open uses an operating system *file descriptor*, and a process
gets a limited number. `defer f.Close()` right after the error check means you
can't forget to give it back.

### Writing: os.Create truncates

`os.Create(path)` opens a file for writing. If the file exists, it is
**truncated** (emptied) first. If it doesn't exist, it is created. A file is
also an `io.Writer`, so `fmt.Fprintln(f, "web01")` writes a line to it.

When you *write*, `Close` matters. The OS may hold your data in a buffer and
report a failure (such as a full disk) only when you close. So for written
files, return the error from `Close` instead of deferring it:

```go
f, err := os.Create(path)
if err != nil {
	return err
}
if _, err := fmt.Fprintln(f, "web01"); err != nil {
	f.Close() // already failing; ignore this second error
	return err
}
return f.Close()
```

### Permission bits

`0o644` is an octal (base 8) number. Each digit is the permissions for one
group of users: read = 4, write = 2, execute = 1, added together.

| Mode | Owner | Group | Others | Typical use |
|---|---|---|---|---|
| `0o644` | read, write | read | read | ordinary files and configs |
| `0o600` | read, write | none | none | secrets, keys, tokens |
| `0o755` | read, write, enter | read, enter | read, enter | directories, scripts |

The process's *umask* then removes some bits, so a requested `0o666` usually
ends up as `0o644`. `os.Create` uses `0o666` before the umask.

### Test files: t.TempDir

The tests call `t.TempDir()`, which gives each test its own empty directory
and deletes it when the test ends. Your code never touches real system files.

## Your task

You're building a small host-maintenance tool. It shows a login banner, reads
the list of hosts to work on, and writes the list back out. Add the imports you
need (`bufio`, `errors`, `fmt`, `io/fs`, `os`, `strings`) as you go.

### 1. Load the login banner

`LoadBanner(path)` returns the whole file as a string, unchanged. If the file
does not exist, return `DefaultBanner` and a `nil` error. Return any other
error as it is (with `""`), for example when `path` is a directory.

```go
LoadBanner("banner.txt") // file holds "Maintenance tonight 22:00 UTC\n"
// => "Maintenance tonight 22:00 UTC\n", nil
LoadBanner("missing.txt")
// => DefaultBanner, nil
```

### 2. Read the host list

`ReadHosts(path)` opens the file and returns one host per line. Trim spaces
from each line with `strings.TrimSpace`, and skip blank lines and lines that
start with `#`. If the file can't be opened, return `nil` and the error
unchanged, so `errors.Is(err, fs.ErrNotExist)` still works for a missing file.

```text
# web tier
web01

  web02
```

```go
ReadHosts("hosts.txt")
// => []string{"web01", "web02"}, nil
```

### 3. Write the host list

`WriteHosts(path, hosts)` creates (or truncates) the file with `os.Create` and
writes each host followed by `"\n"`. An empty list gives an empty file. Return
any error, including the one from `Close`.

```go
WriteHosts("hosts.txt", []string{"db01", "db02"})
// file now holds "db01\ndb02\n"
```

If the parent directory doesn't exist, `os.Create` fails. Return that error.

## Learn more

- [os package](https://pkg.go.dev/os)
- [Go by Example: Reading Files](https://gobyexample.com/reading-files)
- [Go by Example: Writing Files](https://gobyexample.com/writing-files)
- [testing.T.TempDir](https://pkg.go.dev/testing#T.TempDir)
