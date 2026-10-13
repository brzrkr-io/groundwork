# Files: an append-only audit log

Every deploy, restart and config change should leave a trace. In this exercise
you fix an *append-only* audit log: a file that only ever grows, one line per
event, so nobody can quietly lose history.

## Learn

**Recap.** `os.ReadFile` and `os.WriteFile` handle small whole files.
`os.Open` plus `defer f.Close()` plus `bufio.Scanner` reads line by line.
`os.Create` opens for writing and *truncates* an existing file. A missing file
matches `errors.Is(err, fs.ErrNotExist)`, and for written files you return the
error from `Close`.

### os.OpenFile: choose exactly how to open

`os.Open` and `os.Create` are shortcuts for `os.OpenFile`, which takes *flags*
that you combine with `|`:

```go
f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
```

| Flag | Meaning |
|---|---|
| `os.O_RDONLY` | open for reading only |
| `os.O_WRONLY` | open for writing only |
| `os.O_RDWR` | open for reading and writing |
| `os.O_CREATE` | create the file if it doesn't exist |
| `os.O_APPEND` | every write goes to the end of the file |
| `os.O_TRUNC` | empty the file when opening it |
| `os.O_EXCL` | with `O_CREATE`: fail if the file already exists |

Use exactly one of the first three, plus any of the others. The last argument
is the permission used only when the file is created.

| Shortcut | Same as |
|---|---|
| `os.Open(p)` | `os.OpenFile(p, os.O_RDONLY, 0)` |
| `os.Create(p)` | `os.OpenFile(p, os.O_RDWR\|os.O_CREATE\|os.O_TRUNC, 0o666)` |

So `os.Create` on a log file wipes it every time. For a log you want
`O_APPEND|O_CREATE|O_WRONLY`: create it the first time, then always add to the
end.

### os.MkdirAll: make the directory first

Opening `/var/log/deployer/audit.log` fails if `/var/log/deployer` doesn't
exist. `os.MkdirAll` creates a directory and any missing parents, like
`mkdir -p`. If the directory already exists, it does nothing and returns `nil`:

```go
if err := os.MkdirAll(dir, 0o755); err != nil {
	return err
}
```

### "Nothing yet" is not an error

Before the first event, the log doesn't exist. For a reader, that simply means
"no entries". Return an empty slice, `[]string{}`, rather than an error. An
empty slice (not `nil`) also encodes as `[]` instead of `null` if you later
print it as JSON.

## Your task

Your deploy tool records each action in `<dir>/audit.log`. `LogName` holds the
file name. The current code loses history and crashes on fresh machines.

### 1. Stop wiping the log

`AppendAudit(dir, entry)` opens the log with `os.Create`, so each call erases
the earlier entries. Open it with `os.OpenFile` so new entries are added to the
end and the file is created (mode `0o644`) if needed. Each entry is written as
one line ending in `"\n"`. While you're there, return the error from `Close`
instead of ignoring it.

```go
AppendAudit(dir, "deploy api v1.4.2 by alice")
AppendAudit(dir, "rollback api v1.4.1 by bob")
// audit.log holds "deploy api v1.4.2 by alice\nrollback api v1.4.1 by bob\n"
```

### 2. Create the log directory

On a fresh host, `dir` may not exist yet, and may be several levels deep (for
example `<tmp>/var/log/deployer`). Make `AppendAudit` create it, and any
missing parents, with mode `0o755` before opening the log.

### 3. Read a log that doesn't exist yet

`ReadAudit(dir)` returns the log's lines, without their `"\n"`. When the log
file doesn't exist (even if `dir` doesn't exist either), return an empty,
non-nil slice `[]string{}` and a `nil` error. Return any other error as it is.
You will need to add `"errors"` and `"io/fs"` to the imports.

```go
ReadAudit(freshDir)
// => []string{}, nil
```

## Learn more

- [os.OpenFile](https://pkg.go.dev/os#OpenFile)
- [os.MkdirAll](https://pkg.go.dev/os#MkdirAll)
- [Go by Example: Directories](https://gobyexample.com/directories)
