# Version info and ldflags

During an incident, the first question is "what exactly is running?". A
good tool answers with `deployctl --version`: the release number and the
exact commit it was built from. Go can bake both into the binary at build
time.

## Learn

### Stamping a version with -ldflags

The Go *linker* can overwrite a package-level string variable while it
builds the binary. You pass it a flag through `-ldflags`:

```go
package main

var version string // empty unless the build sets it
```

```bash
go build -ldflags "-X main.version=1.4.2" -o deployctl .
./deployctl
# deployctl 1.4.2 (commit 3f2a9c1)
```

`-X main.version=1.4.2` means "in package `main`, set the variable `version`
to `1.4.2`". Release pipelines usually pass the git tag, for example
`-X main.version=$(git describe --tags)`. It only works on a `var` of type
`string`, not on a `const`. Without the flag, `version` stays `""`.

### Build info the toolchain records for you

Every Go binary also carries build information. Read it at run time with
`runtime/debug`:

```go
info, ok := debug.ReadBuildInfo()
if !ok {
	// no build info: info is nil
}
```

`info` is a `*debug.BuildInfo`. The parts you need:

| Field | Example | Meaning |
|---|---|---|
| `info.GoVersion` | `"go1.26.1"` | the Go version that built it |
| `info.Main.Path` | `"example.com/deployctl"` | the main module path |
| `info.Main.Version` | `"v1.4.2"` or `"(devel)"` | the module version; `(devel)` means "no version known" |
| `info.Settings` | slice of `debug.BuildSetting` | key/value build details |

`go install example.com/deployctl@v1.4.2` records `v1.4.2` as the module
version. A `go build` outside a tagged git checkout records `(devel)`.

When you build inside a git repository, `Settings` includes version-control
details:

| `Key` | `Value` example |
|---|---|
| `vcs.revision` | `3f2a9c1e8b7d6a5f4e3d2c1b0a9f8e7d6c5b4a39` (the full commit hash) |
| `vcs.modified` | `"true"` if there were uncommitted changes, else `"false"` |

`Settings` is a slice of `debug.BuildSetting` structs with `Key` and `Value`
fields, so you loop over it to find a key.

### Making it testable

You can't control `debug.ReadBuildInfo` from a test. So `main` reads it and
passes the result to ordinary functions, and the tests build their own
values instead:

```go
info := &debug.BuildInfo{
	Main:     debug.Module{Version: "v1.4.2"},
	Settings: []debug.BuildSetting{{Key: "vcs.revision", Value: "3f2a9c1e8b7d"}},
}
```

Your functions must also handle `info == nil`. Reading `info.Main` through a
nil pointer would crash the program.

## Your task

`deployctl` should print a version line like
`deployctl 1.4.2 (commit 3f2a9c1)`. `main` is already written. You write the
three functions it relies on.

### 1. Pick the version

Finish `Version(stamped string, info *debug.BuildInfo) string`. Use the
first of these that applies:

1. `stamped` is not empty: return `stamped`.
2. `info` is not nil and `info.Main.Version` is neither `""` nor `"(devel)"`:
   return `info.Main.Version`.
3. Otherwise return `"dev"`.

```go
Version("1.4.2", nil)                                                 // => "1.4.2"
Version("", &debug.BuildInfo{Main: debug.Module{Version: "v2.0.0"}})  // => "v2.0.0"
Version("", &debug.BuildInfo{Main: debug.Module{Version: "(devel)"}}) // => "dev"
Version("", nil)                                                      // => "dev"
```

### 2. Find the commit

Finish `Revision(info *debug.BuildInfo) string`:

- Find the setting with key `vcs.revision`. Return its first 7 characters,
  or the whole value if it is shorter.
- If the setting `vcs.modified` is `"true"`, add `-dirty` to the end.
- If `info` is nil, or there is no `vcs.revision` setting, return
  `"unknown"`.

```go
// Settings: vcs.revision=3f2a9c1e8b7d6a5f, vcs.modified=false
Revision(info) // => "3f2a9c1"
// Settings: vcs.revision=3f2a9c1e8b7d6a5f, vcs.modified=true
Revision(info) // => "3f2a9c1-dirty"
Revision(nil)  // => "unknown"
```

### 3. Build the version line

Finish `VersionLine(name, stamped string, info *debug.BuildInfo) string`.
It returns `<name> <version> (commit <revision>)`, using your two functions:

```go
VersionLine("deployctl", "1.4.2", info) // => "deployctl 1.4.2 (commit 3f2a9c1)"
VersionLine("deployctl", "", nil)       // => "deployctl dev (commit unknown)"
```

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
