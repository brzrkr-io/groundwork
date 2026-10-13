# internal packages

Exported names are visible to every package, including packages written by
other teams. Sometimes you want a helper that *your* packages share but
nobody else can import. Go has a folder name for exactly that: `internal`.

## Learn

### The internal rule

A package whose import path contains a folder called `internal` can only be
imported by code inside the folder that *contains* `internal`.

For this exercise's module, `gopherlings.local/internal_packages`:

| Importing package | Imports `…/internal/slug`? |
|---|---|
| `gopherlings.local/internal_packages/deploykit` | yes, it lives under the module root |
| `gopherlings.local/internal_packages/cmd/release` | yes |
| `github.com/other-team/tool` | no, compile error |

When the other team tries anyway, the build stops with:

```text
use of internal package gopherlings.local/internal_packages/internal/slug not allowed
```

The go command enforces this, so it can't be skipped by accident. This
exercise can't show you that error, because here every package sits under
the module root and is allowed in. The next exercise puts an `internal`
folder deeper in the tree, where you'll hit the error for real.

### Why it exists

Your module's *public API* is everything another module can import. Every
exported name in it is a promise: rename it and someone's build breaks. An
`internal` package can export as much as it likes, because only your own
code can reach it. You keep the freedom to change it.

Most Go tools in the wild look like this:

```text
.
├── go.mod
├── cmd/mytool/main.go     the program
├── internal/…             helpers only this module may use
└── pkg or top-level dirs  packages you want others to import (optional)
```

### Using an internal package

Inside the module it is an ordinary package. Import it by its full path and
call its exported names:

```go
import "gopherlings.local/internal_packages/internal/slug"

slug.Make("Payments API") // => "payments-api"
```

The package is called `slug`, not `internal`. `internal` is just a folder
along the path.

## Your task

`deploykit` is a library that other teams import to name their releases. It
also exports a helper, `Slugify`, and two teams have started calling it
directly. Before more teams do, move the helper somewhere they can't reach.

```text
.
├── go.mod
├── deploykit/
│   ├── deploykit.go          public API: ReleaseName (and Slugify, for now)
│   └── deploykit_test.go
└── internal/slug/
    ├── slug.go               package slug: where the helper belongs
    └── slug_test.go
```

The `slug` tests don't compile until task 1 is done.

### 1. Move Slugify into internal/slug as Make

Move `Slugify` from `deploykit/deploykit.go` into `internal/slug/slug.go` and
rename it `Make`. `slug.Make` is easier to read than `slug.Slugify`. Its
behaviour stays the same:

- It lower-cases the text.
- Every character that isn't `a`–`z` or `0`–`9` becomes `-`.

```go
slug.Make("Payments API")    // => "payments-api"
slug.Make("checkout-v1.4.2") // => "checkout-v1-4-2"
```

Leave `Slugify` in `deploykit.go` for now. You remove it in task 3.

### 2. Make the names Kubernetes-safe

Kubernetes object names must be at most 63 characters, and must start and
end with a letter or digit. Now that only your code calls `Make`, you can
change it freely. After the two steps above, Make should:

- Keep only the first 63 characters, if it has more.
- Then remove any `-` from the start and the end.

```go
slug.Make("  API  ")               // => "api"
slug.Make("checkout (canary)")    // => "checkout--canary"
slug.Make(strings.Repeat("a", 70)) // => 63 a's
```

`TestMakeLength` checks these rules.

### 3. Switch deploykit over and delete Slugify

In `deploykit/deploykit.go`:

1. Import `gopherlings.local/internal_packages/internal/slug`.
2. Make `ReleaseName` call `slug.Make`.
3. Delete `Slugify`, and the `strings` import it used.

`ReleaseName` joins the service and version with `-` and turns the result
into a slug:

```go
deploykit.ReleaseName("Payments API", "v1.4.2")
// => "payments-api-v1-4-2"
```

`TestReleaseName` includes a very long service name. It only passes once
`ReleaseName` uses your 63-character `Make`.

## Learn more

- [Go command: Internal directories](https://pkg.go.dev/cmd/go#hdr-Internal_Directories)
- [Organizing a Go module](https://go.dev/doc/modules/layout)
- [Kubernetes: Object names](https://kubernetes.io/docs/concepts/overview/working-with-objects/names/)
