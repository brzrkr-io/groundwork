# Build info: what's inside the binary

A security advisory lands: "versions of `golang.org/x/net` before v0.38.0 are
affected". Which of your binaries are? Every Go binary records the exact
version of every module it was built with, and you can read that list from
inside the program.

## Learn

### Recap

`-ldflags "-X main.version=1.4.2"` sets a string variable at build time.
`debug.ReadBuildInfo()` returns a `*debug.BuildInfo` (or `nil`, `false`)
with the Go version, the main module and `Settings` such as
`vcs.revision`. Functions take the `*debug.BuildInfo` as a parameter so
tests can build their own.

### Dependencies

`info.Deps` lists every module compiled into the binary. It is a
`[]*debug.Module`, and each module has:

| Field | Example |
|---|---|
| `Path` | `"k8s.io/client-go"` |
| `Version` | `"v0.37.1"` |
| `Replace` | `nil`, or a `*debug.Module` that was used instead |

A `replace` line in `go.mod` swaps a dependency for another version or a
fork. When that happens, `Replace` points at the module that was actually
built in, and *its* `Version` is the one that matters.

```go
dep := info.Deps[0]
if dep.Replace != nil {
	fmt.Println("really built with", dep.Replace.Version)
}
```

You can see the same data from the shell, without running the program:

```bash
go version -m ./depcheck
```

### Build time

Builds inside a git checkout also record `vcs.time`: the time of the commit,
in RFC 3339 format, such as `2026-10-01T14:03:27Z`. Parse it with
`time.Parse(time.RFC3339, value)`, as in *Time and durations*.

## Your task

`depcheck` prints when it was built and which versions of a few important
libraries it contains. The functions are half-written.

### 1. Respect replacements in DepVersion

`DepVersion(info, path)` finds a dependency by its module path and returns
its version, or `"not found"`. It ignores `Replace`, so it reports the wrong
version for replaced modules. Fix it: when the dependency's `Replace` is not
`nil`, return `Replace.Version` instead.

```go
// Deps: k8s.io/client-go v0.37.1
//       golang.org/x/net v0.30.0, replaced by golang.org/x/net v0.38.0
DepVersion(info, "k8s.io/client-go")  // => "v0.37.1"
DepVersion(info, "golang.org/x/net")  // => "v0.38.0"
DepVersion(info, "github.com/nope/x") // => "not found"
```

### 2. Format the build time

Finish `BuildTime(info) string`. Find the `vcs.time` setting, parse it with
`time.RFC3339`, and format it in UTC as `2006-01-02 15:04 UTC`:

```go
// Settings: vcs.time=2026-10-01T14:03:27Z
BuildTime(info) // => "2026-10-01 14:03 UTC"
// Settings: vcs.time=2026-10-01T16:03:27+02:00
BuildTime(info) // => "2026-10-01 14:03 UTC"
```

Return `"unknown"` when `info` is nil, when there is no `vcs.time`, or when
it doesn't parse. Add `"time"` to the imports.

### 3. Write the report

Finish `Report(info, paths) []string`. It returns one line per path, in the
same order, made of the path, a space and `DepVersion`:

```go
Report(info, []string{"golang.org/x/net", "github.com/nope/x"})
// => []string{"golang.org/x/net v0.38.0", "github.com/nope/x not found"}
```

An empty `paths` gives an empty result (`nil` is fine).

When it's done, `go run .` prints the report for this exercise. It has no
dependencies, so both libraries are `not found`.

## Learn more

- [runtime/debug.Module](https://pkg.go.dev/runtime/debug#Module)
- [go version -m](https://pkg.go.dev/cmd/go#hdr-Print_Go_version)
- [Go vulnerability management (govulncheck)](https://go.dev/doc/security/vuln/)
