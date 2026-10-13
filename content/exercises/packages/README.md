# Packages and imports

Real tools are split into packages: one for talking to the inventory, one for
planning deploys, one for the command line. This exercise shows how Go finds
your packages and how one package calls another.

## Learn

### A package is a folder

A *package* is a folder of `.go` files that are compiled together. Every file
in the folder starts with the same *package clause*:

```go
package inventory
```

Two files in one folder that disagree about the package name won't compile.
Files in different folders are always different packages, even if they use
the same name.

### A module is a tree of packages

The `go.mod` file at the top of this exercise marks a *module*: a tree of
packages that are built and versioned together. Its first line names the
module:

```text
module gopherlings.local/packages
```

That *module path* is the start of every *import path* inside the module.
You get a package's import path by adding its folder to the module path:

| Folder | Import path | Package clause |
|---|---|---|
| `inventory/` | `gopherlings.local/packages/inventory` | `package inventory` |
| `deploy/` | `gopherlings.local/packages/deploy` | `package deploy` |

Standard library packages have short paths like `"strings"` because they ship
with Go. Your own packages always need the full path. If you write only
`"inventory"`, Go looks in the standard library and gives up:

```text
package inventory is not in std
```

### Package name and folder name

The import path says *where* the package lives. The package clause says what
you *call* it in code. By convention they match the last part of the path:
you import `"net/http"` and then call `http.Get(url)`. Keep them the same in
your own code and nobody has to guess.

### Calling another package

Import the package, then put its name and a dot in front of what you use:

```go
package deploy

import (
	"strings"

	"gopherlings.local/packages/inventory"
)

func Label(host string) string {
	return strings.ToUpper(inventory.Address(host, 443))
}
```

`gofmt` keeps standard library imports in the first group and your own
packages in a second group, separated by a blank line.

You can only use names that start with a capital letter, like `Address`.
*Exported and unexported names* covers this in detail.

### Imports are strict

Go refuses to compile a file that imports a package it doesn't use, or uses a
package it doesn't import:

```text
"strings" imported and not used
undefined: strings
```

When you move code between files, its imports have to move with it.

### Tests per package

A test file sits in the folder of the package it tests and uses the same
package clause. `go test ./...` tests every package in the module, and each
package passes or fails on its own. So you may see an error for `deploy` and
a different one for `inventory` at the same time.

## Your task

The deploy planner was split into two packages in a hurry, and now nothing
compiles. Each task fixes one of the compiler errors below. Until all three
are fixed, Gopherlings shows compiler errors instead of a test checklist.
That's expected.

```text
.
├── go.mod
├── deploy/
│   ├── deploy.go           package deploy
│   └── deploy_test.go
└── inventory/
    ├── inventory.go        package inventory
    └── inventory_test.go
```

### 1. Fix the package clause

`inventory/inventory.go` was copied from the `deploy` folder and still says
`package deploy`. The test file next to it says `package inventory`, so Go
reports:

```text
found packages deploy (inventory.go) and inventory (inventory_test.go) in …/inventory
```

Change the package clause in `inventory/inventory.go` to `package inventory`.
`TestAddress` checks the existing `Address` function:

```go
inventory.Address("web-1", 8080)
// => "web-1:8080"
```

### 2. Fix the import path

`deploy/deploy.go` imports `"inventory"`, which fails with
`package inventory is not in std`. Change it to the full import path:

```text
gopherlings.local/packages/inventory
```

`TestTargets` then checks `deploy.Targets`, which turns host names into
addresses:

```go
deploy.Targets([]string{" Web-1", "web-2 "}, 8080)
// => []string{"web-1:8080", "web-2:8080"}
```

### 3. Move NormalizeHost into inventory

Cleaning up host names belongs with the rest of the inventory code.
`TestNormalizeHost` in `inventory/inventory_test.go` calls `NormalizeHost` from
inside the `inventory` package, so it fails with `undefined: NormalizeHost`.

1. Cut `NormalizeHost` (with its doc comment) out of `deploy/deploy.go` and
   paste it at the end of `inventory/inventory.go`.
2. It uses `strings`, so add `"strings"` to the imports in
   `inventory/inventory.go`.
3. In `deploy/deploy.go`, call it as `inventory.NormalizeHost(host)`.
4. `deploy/deploy.go` no longer uses `strings`, so remove that import.

```go
inventory.NormalizeHost("  DB-2.Internal ")
// => "db-2.internal"
```

All three tests pass once both packages compile.

## Learn more

- [How to Write Go Code](https://go.dev/doc/code)
- [Go spec: Packages](https://go.dev/ref/spec#Packages)
- [Go spec: Import declarations](https://go.dev/ref/spec#Import_declarations)
- [Go Modules Reference: go.mod files](https://go.dev/ref/mod#go-mod-file)
