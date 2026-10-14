# Version info and ldflags

During an incident, the first question is "what exactly is running?". A good
tool answers with `deployctl --version`: the release number and the exact
commit it was built from. Go can bake both into the binary at build time.

**New here:** `-ldflags "-X ..."`, `runtime/debug.ReadBuildInfo`,
`debug.BuildInfo`.

## Learn

### Stamping a version with -ldflags

The Go *linker* is the last step of `go build`: it joins your compiled
packages into one binary. It can also overwrite a package-level `string`
variable while it does so. You give it the instruction with `-ldflags`:

```go
package main

var release string // empty unless the build sets it
```

```bash
go build -ldflags "-X main.release=2025.10" -o mytool .
```

`-X main.release=2025.10` means "in package `main`, set the variable
`release` to `2025.10`". Release pipelines usually pass the git tag, for
example `-X main.release=$(git describe --tags)`.

> **Why is it written like this?**
> - **Go requires:** `-X` only works on a package-level `var` of type
>   `string`. A `const` can't be changed, and a variable inside a function
>   isn't visible to the linker.
> - **Go requires:** the name is *package path* + `.` + *variable*. For a
>   program the package is always `main`; for any other package you write its
>   full import path, such as `example.com/tool/internal/version.release`.
> - **Author's choice:** the variable's name. If you misspell it in the flag,
>   the build still succeeds and the variable stays `""`.

### Build info the toolchain records for you

Every Go binary also carries build information. Read it while the program
runs with `runtime/debug`:

```go
info, ok := debug.ReadBuildInfo()
if !ok {
	// no build info: info is nil
}
```

`info` is a `*debug.BuildInfo`, a pointer to a struct. The parts you need:

| Field | Example | Meaning |
|---|---|---|
| `info.GoVersion` | `"go1.26.1"` | the Go version that built it |
| `info.Main.Path` | `"example.com/deployctl"` | the main module path |
| `info.Main.Version` | `"v1.4.2"` or `"(devel)"` | the module version; `(devel)` means "no version known" |
| `info.Settings` | slice of `debug.BuildSetting` | key/value build details |

`go install example.com/deployctl@v1.4.2` records `v1.4.2` as the module
version. A `go build` outside a tagged git checkout records `(devel)`.

### Settings

`Settings` is a `[]debug.BuildSetting`. Each one has a `Key` and a `Value`,
both strings. There's no lookup by key, so you loop over the slice and
compare each `Key`. Inside a git repository, the build records:

| `Key` | `Value` example |
|---|---|
| `vcs.revision` | `3f2a9c1e8b7d6a5f4e3d2c1b0a9f8e7d6c5b4a39` (the full commit hash) |
| `vcs.modified` | `"true"` if there were uncommitted changes, else `"false"` |

Either setting may be missing.

### Making it testable

A test can't control what `debug.ReadBuildInfo` returns. So `main` reads it
once and passes the result to ordinary functions. Tests build their own
values instead:

```go
info := &debug.BuildInfo{
	Main:     debug.Module{Version: "v0.3.0"},
	Settings: []debug.BuildSetting{{Key: "vcs.modified", Value: "false"}},
}
```

Your functions must also handle `info == nil`. Reading `info.Main` through a
nil pointer crashes the program, so check for `nil` *before* you read a field.

> **Why is it written like this?** Passing `info` in as a parameter is the
> author's **choice**, and a common Go habit: the function only uses its
> inputs, so a test can feed it anything. Checking `nil` first is
> **required** in practice: Go doesn't stop you, but the program panics.

### Slicing a string

`s[:n]` is the first `n` bytes of `s`. If `s` is shorter than `n`, it
crashes, so check `len(s)` first:

```go
id := "a81c5f"
short := id
if len(short) > 4 {
	short = short[:4] // "a81c"
}
```

## Your task

`deployctl` should print a version line like
`deployctl 1.4.2 (commit 3f2a9c1)`. `main` is already written. You write the
three functions it relies on.

### 1. Pick the version

**Goal:** finish `Version(stamped, info)` so it returns the best version it
can find, or `"dev"`.

Use the first rule that applies:

1. `stamped` is not empty: return `stamped`.
2. `info` is not nil, and `info.Main.Version` is neither `""` nor
   `"(devel)"`: return `info.Main.Version`.
3. Otherwise return `"dev"`.

```go
Version("1.4.2", nil)                                                 // => "1.4.2"
Version("", &debug.BuildInfo{Main: debug.Module{Version: "v2.0.0"}})  // => "v2.0.0"
Version("", &debug.BuildInfo{Main: debug.Module{Version: "(devel)"}}) // => "dev"
Version("", nil)                                                      // => "dev"
```

A stamped version wins even when `info` has a module version.

**Think it through:**
- Rule 2 has three conditions that must all be true. Which one must come
  first so the program can't crash?

### 2. Find the commit

**Goal:** finish `Revision(info)` so it returns the short commit hash, with
`-dirty` added when there were uncommitted changes.

- Find the setting with key `vcs.revision`. Use its first 7 characters, or
  the whole value if it's shorter.
- If the setting `vcs.modified` is `"true"`, add `-dirty` to the end.
- If `info` is nil, or there is no `vcs.revision` setting, return
  `"unknown"`.

The two settings can come in any order.

```go
// Settings: vcs.revision=3f2a9c1e8b7d6a5f, vcs.modified=false
Revision(info) // => "3f2a9c1"
// Settings: vcs.modified=true, vcs.revision=3f2a9c1e8b7d6a5f
Revision(info) // => "3f2a9c1-dirty"
// Settings: vcs.revision=abc
Revision(info) // => "abc"
Revision(nil)  // => "unknown"
```

**Think it through:**
- Since the order can vary, can you decide the answer inside the loop, or do
  you need to remember what you saw and decide after it?
- Which check from *Slicing a string* stops a short hash from crashing?

### 3. Build the version line

**Goal:** finish `VersionLine(name, stamped, info)` so it returns
`<name> <version> (commit <revision>)`.

```go
VersionLine("deployctl", "1.4.2", info) // => "deployctl 1.4.2 (commit 3f2a9c1)"
VersionLine("deployctl", "", nil)       // => "deployctl dev (commit unknown)"
```

**Think it through:**
- `name` is a parameter. The version and revision come from your two
  functions. Which parts are fixed text (spaces, `(commit `, `)`)?

Then try it from the exercise folder:

```bash
go run -ldflags "-X main.version=1.4.2" .
go run .
```

The first prints `deployctl 1.4.2 (commit unknown)`. The second has no
stamped version, so it prints `dev`, unless the folder is inside a tagged git
repository. The commit is `unknown` unless the folder is inside a git
repository.

## Learn more

- [runtime/debug.BuildInfo](https://pkg.go.dev/runtime/debug#BuildInfo)
- [Go command: Build flags (-ldflags)](https://pkg.go.dev/cmd/go#hdr-Compile_packages_and_dependencies)
- [Linker flags (-X)](https://pkg.go.dev/cmd/link)
