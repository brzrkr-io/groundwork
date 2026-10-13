## Task 1: same code, new home

Copy the body of `Slugify` into a new function in `internal/slug/slug.go`:

```go
// Make turns s into a lower-case name made of letters, digits and dashes.
func Make(s string) string {
	// the body of Slugify
}
```

Add `import "strings"` at the top of `slug.go`.

## Task 2: cut, then trim

Store the builder's result in a variable instead of returning it straight
away. `len(name) > 63` tells you when to cut with `name[:63]`, and
`strings.Trim(name, "-")` removes dashes from both ends.

## Task 3: a one-line change plus clean-up

```go
import "gopherlings.local/internal_packages/internal/slug"

func ReleaseName(service, version string) string {
	return slug.Make(service + "-" + version)
}
```

Then delete `Slugify`. Nothing else in `deploykit.go` uses `strings`, so
remove that import too.
