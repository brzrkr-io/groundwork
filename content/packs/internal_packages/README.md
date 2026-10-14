# internal packages

Exported names are visible to every package, including packages written by
other teams. Sometimes you want a helper that *your* packages share but
nobody else can import. Go has a folder name for exactly that: `internal`.

**New here:** `internal` folders, public API, standard module layout.

## Learn

### The internal rule

A package with a folder called `internal` in its import path can only be
imported by code inside the folder that *contains* that `internal` folder.

For a module called `example.com/shipit`, with a helper in
`internal/retry/`:

| Importing package | May import `example.com/shipit/internal/retry`? |
|---|---|
| `example.com/shipit/client` | yes, it is inside the module root |
| `example.com/shipit/cmd/shipit` | yes |
| `github.com/other-team/tool` | no, compile error |

When the other team tries anyway, the build stops with:

```text
use of internal package example.com/shipit/internal/retry not allowed
```

> **Why is it written like this?**
> - **Go requires** it: the `go` command enforces the rule, so nobody can
>   skip it by accident. `internal` is the only folder name with a special
>   meaning like this.
> - The folder must be called exactly `internal`, lower case.

In this exercise every package sits under the module root, so you won't see
that error. *internal packages: the boundary* puts `internal` deeper in the
tree, where you'll hit it for real.

### Why it exists

Your module's *public API* is everything another module can import. Every
exported name in it is a promise: rename it and someone's build breaks. An
`internal` package can export as much as it likes, because only your own
code can reach it. You keep the freedom to change it.

Most Go tools look roughly like this:

```text
.
├── go.mod
├── cmd/mytool/main.go     the program
├── internal/…             helpers only this module may use
└── other folders          packages you want others to import (optional)
```

> **Why is it written like this?**
> - **Convention:** `cmd/<name>/` for programs and `internal/` for private
>   helpers. Go doesn't require this layout, but most projects use it, so
>   readers know where to look.
> - **Author's choice:** what goes in `internal`. Start there and move code
>   out only when another module truly needs it.

### Using an internal package

Inside the module it is an ordinary package. Import it by its full path
(module path plus folders) and call its exported names:

```go
import "example.com/shipit/internal/retry"

retry.Times(3)
```

The package is called `retry`, not `internal`. `internal` is just a folder
along the path. Its functions still need capital letters, because the
importing package (`client`) is a different package.

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

**Goal:** `internal/slug/slug.go` has a function `Make` that does exactly what `Slugify` does.

It's called `Make` because callers write `slug.Make`, which reads better
than `slug.Slugify`. Its behaviour stays the same:

- It lower-cases the text.
- Every character that isn't `a`–`z` or `0`–`9` becomes `-`.

```go
slug.Make("Payments API")    // => "payments-api"
slug.Make("checkout-v1.4.2") // => "checkout-v1-4-2"
```

Leave `Slugify` in `deploykit.go` for now. You remove it in task 3.

**Think it through:**
- Copy the doc comment, `func` line and body. Which part of the `func` line changes?
- Which package does the body use? Does `slug.go` import it yet?

### 2. Make the names Kubernetes-safe

**Goal:** `Make` returns names of at most 63 characters, with no `-` at the start or end.

Kubernetes object names must be at most 63 characters and must start and
end with a letter or digit. Only your module calls `Make`, so you can change
it freely. After the existing steps, `Make` should, in this order:

1. Keep only the first 63 characters, if there are more.
2. Then remove any `-` from the start and the end. Dashes in the middle stay.

```go
slug.Make("  API  ")               // => "api"
slug.Make("checkout (canary)")     // => "checkout--canary"
slug.Make(strings.Repeat("a", 70)) // => 63 a's
```

`TestMakeLength` checks these rules, including a name whose 63rd
character is a dash.

**Think it through:**
- Right now `Make` returns the builder's result straight away. Where can you keep it so you can change it first?
- Why must cutting come *before* trimming? Think about a name that is cut on a dash.
- Look in the `strings` package for a function that trims a character from both ends.

### 3. Switch deploykit over and delete Slugify

**Goal:** `ReleaseName` uses `slug.Make`, and `Slugify` is gone from `deploykit.go`.

`ReleaseName` joins the service and version with `-` and turns the result
into a slug:

```go
deploykit.ReleaseName("Payments API", "v1.4.2")
// => "payments-api-v1-4-2"
```

`TestReleaseName` includes a very long service name. It only passes once
`ReleaseName` uses your 63-character `Make`.

**Think it through:**
- What is the import path of `internal/slug`? Start from the module path in `go.mod`.
- In `ReleaseName`, only the function being called changes.
- After deleting `Slugify`, which import in `deploykit.go` is unused?

## Learn more

- [Go command: Internal directories](https://pkg.go.dev/cmd/go#hdr-Internal_Directories)
- [Organizing a Go module](https://go.dev/doc/modules/layout)
- [Kubernetes: Object names](https://kubernetes.io/docs/concepts/overview/working-with-objects/names/)
