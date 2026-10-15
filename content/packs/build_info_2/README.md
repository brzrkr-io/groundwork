# Build info: what's inside the binary

A security advisory lands: "versions of `golang.org/x/net` before v0.38.0 are
affected". Which of your binaries are? Every Go binary records the exact
version of every module it was built with, and you can read that list from
inside the program.

**New here:** `info.Deps`, `debug.Module.Replace`, the `vcs.time` setting.

## Learn

### Recap

As in *Version info and ldflags*: `debug.ReadBuildInfo()` returns a
`*debug.BuildInfo` (or `nil`), with the main module and `Settings` such as
`vcs.revision`. Functions take the `*debug.BuildInfo` as a parameter, so
tests can build their own. Check `info == nil` before reading a field.

### Packages, modules and dependencies

A *package* is one folder of Go code. A *module* is a set of packages
released together, with one `go.mod` at the top and one version number. Your
program's `go.mod` lists the other modules it needs: its *dependencies*.

`info.Deps` lists every module compiled into the binary. It's a
`[]*debug.Module`, and each module has:

| Field | Example |
|---|---|
| `Path` | `"k8s.io/client-go"` |
| `Version` | `"v0.37.1"` |
| `Replace` | `nil`, or a `*debug.Module` that was used instead |

> **Why is it written like this?** Versions belong to modules, not packages,
> so Go **requires** that you ask about a module path (`golang.org/x/net`),
> not a package inside it.

### Replacements

A `replace` line in `go.mod` swaps a dependency for another version or a
fork. When that happens, `Replace` points at the module that was really
built in, and *its* `Version` is the one that matters.

```go
dep := info.Deps[0]
fmt.Println(dep.Path) // the module go.mod asked for
if dep.Replace != nil {
	fmt.Println("built from", dep.Replace.Path)
}
```

`dep.Replace.Path` can even be a different module, such as your own fork.

You can see the same data from the shell, without running the program:

```bash
go version -m ./depcheck
```

### Build time

Builds inside a git checkout also record `vcs.time`: the time of the commit,
in RFC 3339 format, such as `2026-10-01T14:03:27Z`. As in *Time and
durations*:

| Step | Call |
|---|---|
| text → `time.Time` | `time.Parse(time.RFC3339, value)` (returns an error for bad text) |
| any zone → UTC | `t.UTC()` |
| `time.Time` → text | `t.Format(layout)`, where the layout is written using the date `2006-01-02 15:04:05` |

```go
t.UTC().Format("15:04 on 02 Jan") // => "14:03 on 01 Oct"
```

## Your task

`depcheck` prints when it was built and which versions of a few important
libraries it contains. The functions are half-written.

### 1. Respect replacements in DepVersion

**Goal:** make `DepVersion(info, path)` return the replacement's version
when the dependency was replaced.

It already finds a dependency by module path and returns its `Version`, or
`"not found"` (also when `info` is nil). It ignores `Replace`, so it reports
the wrong version for replaced modules.

```go
// Deps: k8s.io/client-go v0.37.1
//       golang.org/x/net v0.30.0, replaced by golang.org/x/net v0.38.0
//       sigs.k8s.io/yaml v1.6.0, replaced by example.com/yaml-fork v1.6.1-fixed
DepVersion(info, "k8s.io/client-go")  // => "v0.37.1"
DepVersion(info, "golang.org/x/net")  // => "v0.38.0"
DepVersion(info, "sigs.k8s.io/yaml")  // => "v1.6.1-fixed"
DepVersion(info, "github.com/nope/x") // => "not found"
```

**Think it through:**
- Look at the `TODO(1)` line inside the loop. What must you check before the
  existing `return`?

### 2. Format the build time

**Goal:** finish `BuildTime(info)` so it returns the `vcs.time` setting in
UTC, as `2006-01-02 15:04 UTC`.

```go
// Settings: vcs.time=2026-10-01T14:03:27Z
BuildTime(info) // => "2026-10-01 14:03 UTC"
// Settings: vcs.time=2026-10-01T16:03:27+02:00
BuildTime(info) // => "2026-10-01 14:03 UTC"
// Settings: vcs.time=2026-03-02T01:30:00+03:00
BuildTime(info) // => "2026-03-01 22:30 UTC"
```

- Return `"unknown"` when `info` is nil, when there is no `vcs.time`
  setting, or when its value doesn't parse (such as `"yesterday"`).
- Add `"time"` to the imports.

**Think it through:**
- Follow the three steps in the *Build time* table, in order. Which of them
  can fail?
- The word `UTC` at the end is fixed text. Does it belong in the layout or
  after it?

### 3. Write the report

**Goal:** finish `Report(info, paths)` so it returns one line per path, in
the same order: the path, a space, then `DepVersion`.

```go
Report(info, []string{"golang.org/x/net", "github.com/nope/x"})
// => []string{"golang.org/x/net v0.38.0", "github.com/nope/x not found"}
```

An empty `paths` gives an empty result (`nil` is fine).

**Think it through:**
- Which part of each line is the loop variable, which is fixed text, and
  which is a function call?

When it's done, `go run .` prints the report for this exercise. It has no
dependencies, so both libraries are `not found`.

## Learn more

- [runtime/debug.Module](https://pkg.go.dev/runtime/debug#Module)
- [go version -m](https://pkg.go.dev/cmd/go#hdr-Print_Go_version)
- [Go vulnerability management (govulncheck)](https://go.dev/doc/security/vuln/)
