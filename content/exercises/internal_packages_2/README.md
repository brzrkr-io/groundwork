# internal packages: the boundary

An `internal` folder doesn't have to sit at the top of a module. Put it
deeper and it protects one corner of your code from the rest of the same
module. This time you'll meet the compiler error for real.

**New here:** `internal` deeper in the tree, fixing a boundary error.

## Learn

### Recap

Packages under a folder called `internal` can only be imported by code
inside the folder that contains `internal`. Everyone else gets
`use of internal package … not allowed`. That keeps helpers out of your
public API, so you can change them freely.

### internal deeper in the tree

The rule is measured from the *parent* of `internal`, wherever it is. Take a
module with a helper at `billing/internal/tax`. Its boundary is the
`billing/` folder:

| Package | May import `billing/internal/tax`? |
|---|---|
| `billing` | yes, it is the parent |
| `billing/invoice` | yes, it is inside `billing/` |
| `shipping` | no, it is outside `billing/`, even though it's in the same module |

> **Why is it written like this?**
> - **Go requires** it: the boundary is the folder that holds `internal`,
>   not the module. Being in the same module isn't enough.
> - **Author's choice:** where you put `internal`. Deeper means fewer
>   packages may use it.

### What to do when you hit the boundary

The error means the package's owner decided not to share that code. Don't
copy the helper. You have two honest fixes:

1. Use the public API that the owner *does* export. Often an exported
   function in the parent package calls the internal helper for you.
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

**Goal:** `report/report.go` uses the public `deploy` package instead of `deploy/internal/naming`.

`go test` reports:

```text
report/report.go:…: use of internal package gopherlings.local/internal_packages_2/deploy/internal/naming not allowed
```

`deploy` exports `ReleaseName`, which takes the same arguments and calls the
internal helper for you. `Line` builds one line of the deploy status report:

```go
report.Line("payments", "v1.4.2", true)  // => "payments-v1-4-2: healthy"
report.Line("checkout", "v2.0.0", false) // => "checkout-v2-0-0: DOWN"
```

**Think it through:**
- Two things change: the import path and the function call. What is the import path of the `deploy` folder?
- The status text and the `": "` already work. Leave them alone.

### 2. Fix the bug inside naming

**Goal:** `naming.Release` in `deploy/internal/naming/naming.go` returns a lower-case name.

It already joins service and version with `-` and turns dots into `-`. It
just keeps capital letters. Nobody outside `deploy/` calls `naming`
directly, so you can fix it without asking other teams.

```go
naming.Release("Payments", "V1.4.2") // => "payments-v1-4-2"
naming.Release("checkout", "v2.0.0") // => "checkout-v2-0-0"
```

**Think it through:**
- Which function in the `strings` package lower-cases a whole string?
- You can wrap the existing expression instead of rewriting it.

### 3. Write rollout.Steps

**Goal:** `Steps` in `deploy/rollout/rollout.go` returns one `deploy <release> to <host>` line per host, in order.

`rollout` is inside `deploy/`, so it *may* import the internal `naming`
package. The release name comes from `naming.Release`:

```go
rollout.Steps("payments", "v1.4.2", []string{"web-1", "web-2"})
// => []string{"deploy payments-v1-4-2 to web-1", "deploy payments-v1-4-2 to web-2"}
```

No hosts means an empty result (`nil` is fine).

**Think it through:**
- `rollout.go` has no imports yet. What is the full import path of `deploy/internal/naming`?
- The release name is the same for every host. Work it out once, before the loop.
- In each line, which words are fixed text and which come from variables? Watch the spaces.

## Learn more

- [Go command: Internal directories](https://pkg.go.dev/cmd/go#hdr-Internal_Directories)
- [Go 1.4 release notes: Internal packages](https://go.dev/doc/go1.4#internalpackages)
- [Organizing a Go module](https://go.dev/doc/modules/layout)
