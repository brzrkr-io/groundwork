# Atomic writes: safely change one setting

Ops tools rarely write a config from scratch. Usually they change one setting
and leave the rest alone: bump `REPLICAS`, flip `LOG_LEVEL`. In this exercise
you build that *read-modify-write* update for an env file, and fix the atomic
writer it relies on.

## Learn

**Recap.** An atomic write creates a temp file in the target's directory with
`os.CreateTemp`, writes it, `Chmod`s it, `Sync`s and `Close`s it, then
`os.Rename`s it over the target. Readers see the old file or the new one, never
half of each. On any failure the temp file must be removed.

### Read, modify, write

Updating one setting takes three steps:

1. **Read** the whole file with `os.ReadFile`. A missing file is not an error
   here: treat it as empty, so the first `SetKey` creates the file.
2. **Modify** the lines in memory. Change only the line you mean to; keep
   comments, blank lines and other keys exactly as they were.
3. **Write** the result back atomically, so a crash halfway never leaves a
   config with half its keys missing.

Doing the change on a `[]string` in a separate function keeps it easy to test
without any files.

### The env file format

Each setting is one line, `KEY=value`. Lines starting with `#` are comments.

```text
# api service
LOG_LEVEL=info
PORT=8080
```

A line belongs to `KEY` when it starts with `KEY=`. So `PORT=8080` belongs to
`PORT`, but `PORT_ADMIN=9000` and `# PORT=80` do not. `strings.HasPrefix(line,
key+"=")` checks exactly that.

The file has a provided `splitLines(s)` helper. It splits text into lines with
the trailing `"\n"` removed, and returns no lines for an empty string.

## Your task

Your deploy tool changes settings in `/etc/default/*`-style env files with
`SetKey`. The file-writing helper it uses has a bug.

### 1. Update the lines in memory

`setLine(lines, key, value)` returns the lines with the line belonging to `key`
replaced by `key + "=" + value`. If no line belongs to `key`, append
`key=value` at the end. All other lines stay as they are, in order. Right now
it always appends, so a file ends up with two `LOG_LEVEL=` lines.

```go
setLine([]string{"# api", "LOG_LEVEL=info", "PORT=8080"}, "LOG_LEVEL", "debug")
// => []string{"# api", "LOG_LEVEL=debug", "PORT=8080"}
setLine([]string{"PORT=8080"}, "REPLICAS", "3")
// => []string{"PORT=8080", "REPLICAS=3"}
```

### 2. Read, modify, write

`SetKey(path, key, value)` already updates the lines with `setLine` and saves
the result with `writeFileAtomic`, ending every line, including the last, with
`"\n"`. But it never reads the file: it starts from empty `data`, so every
other setting is wiped. Read `data` from the file. A missing file counts as
empty; return any other read error.

```go
// app.env holds "# api\nLOG_LEVEL=info\n"
SetKey("app.env", "LOG_LEVEL", "debug")
// app.env holds "# api\nLOG_LEVEL=debug\n"
SetKey("missing.env", "PORT", "9090")
// missing.env holds "PORT=9090\n"
```

The tests check that an existing file is *replaced* rather than rewritten in
place (`!os.SameFile(before, after)`), so keep saving with `writeFileAtomic`,
not `os.WriteFile`. Add the imports you need (`errors`, `io/fs`).

### 3. Fix the leaking helper

`writeFileAtomic` returns errors correctly, but it never removes its temp file
when a step fails. Fix it so that after any failure the directory holds only
what it held before. The test makes `os.Rename` fail by passing a path that is
an existing, non-empty directory.

## Learn more

- [os.ReadFile](https://pkg.go.dev/os#ReadFile)
- [strings.HasPrefix](https://pkg.go.dev/strings#HasPrefix)
- [systemd EnvironmentFile](https://www.freedesktop.org/software/systemd/man/latest/systemd.exec.html#EnvironmentFile=)
