## Task 1: Move Slugify into internal/slug as Make

The code doesn't change, only where it lives and what it's called. Copy the function into `internal/slug/slug.go` and rename it.

### Closer

Copy the body of `Slugify` into `func Make(s string) string` in `slug.go`. It uses `strings.Builder` and `strings.ToLower`, so add `import "strings"` to that file.

### Answer

```go
import "strings"

// Make turns s into a lower-case name made of letters, digits and dashes.
func Make(s string) string {
	var b strings.Builder
	for _, r := range strings.ToLower(s) {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			b.WriteRune(r)
		} else {
			b.WriteByte('-')
		}
	}
	return b.String()
}
```

## Task 2: Make the names Kubernetes-safe

Keep the builder's result in a variable instead of returning it straight away. Then shorten it if needed, and trim the dashes, in that order.

### Closer

`name := b.String()`. If `len(name) > 63`, cut it with `name[:63]`. Then `strings.Trim(name, "-")` removes dashes from both ends.

### Answer

```go
	name := b.String()
	if len(name) > 63 {
		name = name[:63]
	}
	return strings.Trim(name, "-")
}
```

The reference solution names the number with `const maxLen = 63`.

## Task 3: Switch deploykit over and delete Slugify

`deploykit` is in the same module as `internal/slug`, so it may import it. Point `ReleaseName` at `slug.Make`, then remove what's no longer used.

### Closer

The import path is the module path plus the folder: `groundwork.local/internal_packages/internal/slug`. After deleting `Slugify`, nothing in `deploykit.go` uses `strings`.

### Answer

```go
import "groundwork.local/internal_packages/internal/slug"

func ReleaseName(service, version string) string {
	return slug.Make(service + "-" + version)
}
```

Delete `Slugify` and the `"strings"` import.
