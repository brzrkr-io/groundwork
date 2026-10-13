# internal packages: the boundary

An `internal` folder doesn't have to sit at the top of a module. Put it deeper
and it protects one corner of your code from the rest of the same module.
This time you'll meet the compiler error for real.

## Learn

### Recap

Packages under a folder called `internal` can only be imported by code inside
the folder that contains `internal`. Everyone else gets
`use of internal package … not allowed`. That keeps helpers out of your
public API, so you can change them freely.

### internal deeper in the tree

The rule is measured from the parent of `internal`, wherever it is. In this
exercise the helper lives at `deploy/internal/naming`, so its boundary is the
`deploy` folder:

| Package | May import `deploy/internal/naming`? |
|---|---|
| `deploy` | yes, it is the parent |
| `deploy/rollout` | yes, it is inside `deploy/` |
| `report` | no, it is outside `deploy/`, even though it's in the same module |

### What to do when you hit the boundary

The error means the package's owner decided not to share that code. You have
two honest fixes:

1. Use the public API that the owner *does* export. Here, `deploy` exports
   `ReleaseName`, which calls the internal helper for you.
2. If several parts of the module really need the helper, move the
   `internal` folder higher up, so the boundary covers all of them.

Prefer the first. It keeps the owner free to change their internals.

## Your task

```text
.
├── go.mod
├── deploy/
│   ├── deploy.go                 package deploy: ReleaseName
│   ├── internal/naming/
│   │   ├── naming.go             package naming: Release
│   │   └── naming_test.go
│   └── rollout/
│       ├── rollout.go            package rollout: Steps
│       └── rollout_test.go
└── report/
    ├── report.go                 package report: Line
    └── report_test.go
```

The `report` package reaches into `deploy/internal`, so it can't even be
built and none of its tests run. The other packages are tested as usual.

### 1. Use deploy's public API in report

`go test` reports:

```text
report/report.go:…: use of internal package gopherlings.local/internal_packages_2/deploy/internal/naming not allowed
```

In `report/report.go`, import `gopherlings.local/internal_packages_2/deploy`
instead, and call `deploy.ReleaseName` where the code now calls
`naming.Release`. `Line` builds one line of the deploy status report:

```go
report.Line("payments", "v1.4.2", true)  // => "payments-v1-4-2: healthy"
report.Line("checkout", "v2.0.0", false) // => "checkout-v2-0-0: DOWN"
```

### 2. Fix the bug inside naming

Release names must be lower case, but `naming.Release` keeps capital letters.
Change `Release` in `deploy/internal/naming/naming.go` so it lower-cases the
result. Because nobody outside `deploy/` calls `naming` directly, you can fix
it without asking other teams.

```go
naming.Release("Payments", "V1.4.2") // => "payments-v1-4-2"
naming.Release("checkout", "v2.0.0") // => "checkout-v2-0-0"
```

### 3. Write rollout.Steps

`deploy/rollout` is inside `deploy/`, so it *may* import
`gopherlings.local/internal_packages_2/deploy/internal/naming`. Finish
`Steps` in `deploy/rollout/rollout.go`. It returns one line per host, in
order, using `naming.Release` for the name:

```go
rollout.Steps("payments", "v1.4.2", []string{"web-1", "web-2"})
// => []string{"deploy payments-v1-4-2 to web-1", "deploy payments-v1-4-2 to web-2"}
```

No hosts means an empty result (`nil` is fine).

## Learn more

- [Go command: Internal directories](https://pkg.go.dev/cmd/go#hdr-Internal_Directories)
- [Go 1.4 release notes: Internal packages](https://go.dev/doc/go1.4#internalpackages)
- [Organizing a Go module](https://go.dev/doc/modules/layout)
