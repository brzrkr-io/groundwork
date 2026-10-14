# Files: an append-only audit log

Every deploy, restart and config change should leave a trace. Here you fix an
*append-only* audit log: a file that only ever grows, one line per event, so
nobody can quietly lose history.

**New here:** `os.OpenFile` and its flags, `os.MkdirAll`.

## Learn

### Recap

- `os.Open` + `defer f.Close()` + `bufio.Scanner` reads line by line.
- `os.Create` opens for writing and *truncates* (empties) an existing file.
- A missing file matches `errors.Is(err, fs.ErrNotExist)`.
- For files you write, return the error from `Close`.

### os.OpenFile: choose exactly how to open

`os.Open` and `os.Create` are shortcuts for `os.OpenFile`. It takes *flags*,
combined with `|`, and a permission:

```go
// a lock file: read+write, create it, and fail if it already exists
f, err := os.OpenFile("/run/deploy.lock", os.O_RDWR|os.O_CREATE|os.O_EXCL, 0o600)
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

Use exactly **one** of the first three, plus any of the others.

| Shortcut | Same as |
|---|---|
| `os.Open(p)` | `os.OpenFile(p, os.O_RDONLY, 0)` |
| `os.Create(p)` | `os.OpenFile(p, os.O_RDWR\|os.O_CREATE\|os.O_TRUNC, 0o666)` |

So `os.Create` on a log wipes it every time.

> **Why is it written like this?** Go requires all three arguments. The flags
> are bits, and `|` (bitwise OR) is how you switch several on at once; their
> order doesn't matter. The permission is only used when the file is created.

### os.MkdirAll: make the directory first

Opening `/var/log/deployer/audit.log` fails if `/var/log/deployer` doesn't
exist. `os.MkdirAll(path, perm)` creates a directory and any missing parents,
like `mkdir -p`. If it already exists, it does nothing and returns `nil`.

```go
if err := os.MkdirAll("/srv/backups/daily", 0o700); err != nil {
	return err
}
```

> **Why is it written like this?** Directories need the *execute* bit to be
> entered, so they usually get `0o755` (or `0o700` when private). That's
> convention, not a Go rule.

### "Nothing yet" is not an error

Before the first event, the log doesn't exist. For a reader that just means
"no entries", so return an empty slice `[]string{}` and no error. An empty
slice (not `nil`) also prints as `[]` instead of `null` in JSON.

## Your task

Your deploy tool records each action in `<dir>/audit.log`; the constant
`LogName` holds the file name. The current code loses history and fails on
fresh machines.

### 1. Stop wiping the log

**Goal:** in `AppendAudit`, open the log with `os.OpenFile` so new entries go at the end.

- The file is opened write-only, created if missing, with mode `0o644`.
- Each entry is written as one line ending in `"\n"` (the existing
  `Fprintln` already does this).
- Return the error from `Close` instead of ignoring it.

```go
AppendAudit(dir, "deploy api v1.4.2 by alice")
AppendAudit(dir, "rollback api v1.4.1 by bob")
// audit.log holds "deploy api v1.4.2 by alice\nrollback api v1.4.1 by bob\n"
```

An existing log keeps its earlier lines.

**Think it through:**
- Which three flags from the table describe "write-only, create, add to the end"?
- Keep the path argument as it is; only the function and flags change.
- Look at the last two lines: where is Close's error lost?

### 2. Create the log directory

**Goal:** `AppendAudit` creates `dir`, and any missing parents, with mode `0o755` before opening the log.

`dir` may be several levels deep and not exist yet:

```go
AppendAudit(tmp+"/var/log/deployer", "deploy api v2.0.0 by erin")
// creates var/log/deployer, then audit.log holding "deploy api v2.0.0 by erin\n"
```

**Think it through:**
- Which `os` function behaves like `mkdir -p`?
- Which parameter holds the directory? (Not the log path.)

### 3. Read a log that doesn't exist yet

**Goal:** `ReadAudit(dir)` returns `[]string{}, nil` when the log file doesn't exist.

- This holds even when `dir` itself doesn't exist.
- The slice must be empty and **not** `nil`.
- Any other error is returned as it is.
- Add `"errors"` and `"io/fs"` to the imports.

Lines come back without their `"\n"`; that part already works.

```go
ReadAudit(freshDir)
// => []string{}, nil
```

**Think it through:**
- Which error value means "no such file"? How do you test for it?
- That check must come before the general `err != nil` one.

## Learn more

- [os.OpenFile](https://pkg.go.dev/os#OpenFile)
- [os.MkdirAll](https://pkg.go.dev/os#MkdirAll)
- [Go by Example: Directories](https://gobyexample.com/directories)
