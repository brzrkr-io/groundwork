# Packages and imports

Real tools are split into packages: one for the inventory, one for planning
deploys, one for the command line. This exercise shows how Go finds your
packages and how one package calls another.

**New here:** packages, modules, `go.mod`, import paths, calling another package.

## Learn

### A package is a folder

A *package* is a folder of `.go` files that Go compiles together as one unit.
Every file in the folder starts with the same *package clause*:

```go
package billing
```

All the files in one folder can use each other's names freely, as if they
were one big file.

> **Why is it written like this?**
> - **Go requires** the package clause as the first line of code in every
>   `.go` file (comments may come before it).
> - **Go requires** one package per folder. Two files in one folder that
>   disagree about the name won't compile. (Test files get one exception,
>   which you'll meet in *Exported and unexported names*.)
> - **Convention:** the package name is short, lower case, one word, and the
>   same as the folder name.

### A module is a tree of packages

The `go.mod` file at the top of this exercise marks a *module*: a tree of
folders (packages) that are built and versioned together. Its first line
gives the module its name, the *module path*:

```text
module example.com/shop
```

To get a package's *import path*, add its folder to the module path. For a
module called `example.com/shop`:

| Folder | Import path | Package clause |
|---|---|---|
| `billing/` | `example.com/shop/billing` | `package billing` |
| `cart/` | `example.com/shop/cart` | `package cart` |
| `cart/discount/` | `example.com/shop/cart/discount` | `package discount` |

> **Why is it written like this?**
> - **Go requires** the full path for your own packages. Short paths like
>   `"strings"` are reserved for the standard library, which ships with Go.
>   Write just `"billing"` and Go looks in the standard library and gives up:
>   `package billing is not in std`.
> - **Author's choice:** the module path. Real projects use their repository
>   address, like `github.com/acme/deploy-tool`. This course uses
>   `gopherlings.local/<exercise>`. Look in `go.mod` to see this one.

### Where the name comes from

The import path says *where* the package lives. The package clause says what
you *call* it in code. You import `"net/http"` and then write
`http.Get(url)`: the last part of the path and the package name match.

> **Why is it written like this?**
> - **Convention:** keep the folder and the package name the same, so nobody
>   has to open the file to learn what to type.

### Calling another package

Import the package, then write its name and a dot in front of what you use:

```go
package cart

import (
	"strings"

	"example.com/shop/billing"
)

func Total(code string) string {
	return strings.ToUpper(billing.Format(code, 42))
}
```

You can only use names that start with a capital letter, like `Format`.
*Exported and unexported names* covers why.

> **Why is it written like this?**
> - **Go requires** the quotes around an import path, and `name.` in front
>   of everything from another package.
> - **Convention:** `gofmt` puts standard library imports in the first group
>   and your own packages in a second group, with a blank line between.
> - **Go requires** the round brackets only when you import more than one
>   package. With one, `import "strings"` is enough.

### Imports are strict

Go refuses to compile a file that imports a package it doesn't use, or uses
a package it doesn't import:

```text
"strings" imported and not used
undefined: strings
```

Imports belong to one *file*, not the whole package. When you move code from
one file to another, its imports have to move with it.

### Tests per package

A test file sits in the folder of the package it tests and uses the same
package clause. `go test ./...` tests every package, and each one passes or
fails on its own. So you may see one error for `deploy` and a different one
for `inventory` at the same time.

## Your task

The deploy planner was split into two packages in a hurry, and now nothing
compiles. Each task fixes one compiler error. Until all three are fixed,
Gopherlings shows compiler errors instead of a test checklist. That's
expected.

```text
.
├── go.mod
├── deploy/
│   ├── deploy.go           package deploy
│   └── deploy_test.go
└── inventory/
    ├── inventory.go        package inventory (wrong right now)
    └── inventory_test.go
```

### 1. Fix the package clause

**Goal:** make the first line of `inventory/inventory.go` agree with the folder it is in.

The file was copied from the `deploy` folder and still has deploy's package
clause. The test file next to it disagrees, so Go reports:

```text
found packages deploy (inventory.go) and inventory (inventory_test.go) in …/inventory
```

Once it compiles, `TestAddress` checks the existing `Address` function:

```go
inventory.Address("web-1", 8080)
// => "web-1:8080"
```

**Think it through:**
- Compare the first line of `inventory.go` with the first line of `inventory_test.go`.
- Only the package clause changes. Leave the doc comment above it alone.

### 2. Fix the import path

**Goal:** make `deploy/deploy.go` import the inventory package by its full import path.

Right now it imports `"inventory"`, which fails with
`package inventory is not in std`. Once it compiles, `TestTargets` checks
`deploy.Targets`, which cleans up host names and turns them into addresses:

```go
deploy.Targets([]string{" Web-1", "web-2 "}, 8080)
// => []string{"web-1:8080", "web-2:8080"}
```

**Think it through:**
- Open `go.mod`. What is the module path?
- The import path is the module path, a `/`, then the folder name.
- Keep the `"strings"` import for now: `deploy.go` still uses it until task 3.

### 3. Move NormalizeHost into inventory

**Goal:** `NormalizeHost` lives in `inventory/inventory.go`, and `deploy` calls it from there.

Cleaning up host names belongs with the inventory code. `TestNormalizeHost`
runs inside the `inventory` package, so it fails with
`undefined: NormalizeHost` until the function is there:

```go
inventory.NormalizeHost("  DB-2.Internal ")
// => "db-2.internal"
```

The function itself is already correct. You only move it. Four steps:

1. Cut `NormalizeHost` (with its doc comment) out of `deploy/deploy.go` and
   paste it at the end of `inventory/inventory.go`.
2. Fix the imports in `inventory/inventory.go`.
3. Fix the call inside `Targets` in `deploy/deploy.go`.
4. Fix the imports in `deploy/deploy.go`.

**Think it through:**
- Which package does `NormalizeHost` use? Which file uses it now?
- From inside `deploy`, how do you call a function that lives in another package?
- After the move, which import in `deploy.go` is no longer used?

All three tests pass once both packages compile.

## Learn more

- [How to Write Go Code](https://go.dev/doc/code)
- [Go spec: Packages](https://go.dev/ref/spec#Packages)
- [Go spec: Import declarations](https://go.dev/ref/spec#Import_declarations)
- [Go Modules Reference: go.mod files](https://go.dev/ref/mod#go-mod-file)
