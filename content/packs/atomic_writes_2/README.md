# Atomic writes: safely change one setting

Ops tools rarely write a config from scratch. Usually they change one setting
and leave the rest alone: bump `REPLICAS`, flip `LOG_LEVEL`. Here you build
that *read-modify-write* update for an env file, and fix the atomic writer it
relies on.

**New here:** read-modify-write, changing a slice element by index. The rest is practice.

## Learn

### Recap: atomic writes

Create a temp file next to the target with `os.CreateTemp`, write it, `Chmod`,
`Sync`, `Close`, then `os.Rename` it over the target. Readers see the old file
or the new one, never half of each. On any failure the temp file must be
removed, as in *Atomic writes: temp file and rename*.

### Read, modify, write

Changing one setting takes three steps:

1. **Read** the whole file with `os.ReadFile`. A missing file is not an
   error here: treat it as empty, so the first change creates the file.
2. **Modify** the lines in memory. Change only the line you mean to; keep
   comments, blank lines and other keys exactly as they were.
3. **Write** the result back atomically, so a crash never leaves a config
   with half its keys missing.

Doing step 2 on a `[]string` in its own function makes it testable without
files.

### The env file format

Each setting is one line, `KEY=value`. Lines starting with `#` are comments.

```text
# api service
LOG_LEVEL=info
PORT=8080
```

A line *belongs* to a key when it starts with the key followed by `=`. So
`PORT=8080` belongs to `PORT`, but `PORT_ADMIN=9000` and `# PORT=80` don't.

`strings.HasPrefix(s, prefix)` checks how a string starts. The prefix can be
built with `+`:

```go
strings.HasPrefix("web01.prod", "web01"+".") // true
strings.HasPrefix("web011.prod", "web01"+".") // false
```

The provided `splitLines(s)` helper splits text into lines with the trailing
`"\n"` removed, and returns no lines for an empty string.

### Changing a slice element in a loop

`for i, x := range s` gives you the index `i` as well as the value `x`. To
change the slice itself, assign to `s[i]`; assigning to `x` only changes a
copy.

```go
for i, host := range hosts {
	if host == "old01" {
		hosts[i] = "new01"
	}
}
```

> **Why is it written like this?** Go requires `s[i] = ...` to change an
> element: `x` is a copy. The names `i` and `x` are your choice; `i` is
> convention for an index.

### "Missing is fine, other errors aren't"

To accept one error and reject the rest, combine both conditions with `&&`
and `!`:

```go
if err != nil && !errors.Is(err, os.ErrDeadlineExceeded) {
	return err
}
```

## Your task

Your deploy tool changes settings in `/etc/default/*`-style env files with
`SetKey`. The file-writing helper it uses has a bug too.

### 1. Update the lines in memory

**Goal:** `setLine(lines, key, value)` replaces the line belonging to `key` with `KEY=value`, or appends `KEY=value` if there is none.

- The new line is `key`, then `=`, then `value`.
- All other lines stay as they are, in order.
- `lines` may be empty or `nil`.
- Right now it always appends, so a file ends up with two `LOG_LEVEL=` lines.

```go
setLine([]string{"# api", "LOG_LEVEL=info", "PORT=8080"}, "LOG_LEVEL", "debug")
// => []string{"# api", "LOG_LEVEL=debug", "PORT=8080"}
setLine([]string{"PORT_ADMIN=9000"}, "PORT", "8080")
// => []string{"PORT_ADMIN=9000", "PORT=8080"}
```

**Think it through:**
- What must a line start with to belong to `key`? (Why isn't `key` alone
  enough?)
- When you find it, change it and stop. When do you append?

### 2. Read, modify, write

**Goal:** `SetKey` reads `data` from the file at `path` instead of starting empty.

- A missing file counts as empty (the file is then created).
- Return any other read error, for example when `path` is a directory.
- Keep saving with `writeFileAtomic`: the tests check that an existing file
  is *replaced*, not rewritten in place.
- Add `"errors"` and `"io/fs"` to the imports.

The rest of `SetKey` already works: it calls `setLine` and ends every line,
including the last, with `"\n"`.

```go
// app.env holds "# api\nLOG_LEVEL=info\n"
SetKey("app.env", "LOG_LEVEL", "debug")
// app.env holds "# api\nLOG_LEVEL=debug\n"
SetKey("missing.env", "PORT", "9090")
// missing.env holds "PORT=9090\n"
```

**Think it through:**
- Which line declares `data` right now? That's the line to replace.
- Which error must *not* stop `SetKey`?

### 3. Fix the leaking helper

**Goal:** `writeFileAtomic` removes its temp file whenever a step fails.

It already returns errors correctly. After any failure, the directory must
hold only what it held before. The test makes `os.Rename` fail by passing a
path that is an existing, non-empty directory.

**Think it through:**
- After which line does a temp file exist?
- The last line returns `os.Rename`'s result directly. How will your cleanup
  know whether the rename worked?

## Learn more

- [os.ReadFile](https://pkg.go.dev/os#ReadFile)
- [strings.HasPrefix](https://pkg.go.dev/strings#HasPrefix)
- [systemd EnvironmentFile](https://man7.org/linux/man-pages/man5/systemd.exec.5.html)
