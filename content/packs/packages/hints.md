## Task 1: Fix the package clause

Every `.go` file in one folder must start with the same `package` line. Compare the first line of `inventory/inventory.go` with the first line of `inventory/inventory_test.go`.

### Closer

The folder is called `inventory` and its test file says `package inventory`. Change only the first line of `inventory/inventory.go`.

### Answer

```go
package inventory
```

## Task 2: Fix the import path

An import path for your own package is the module path from `go.mod`, then the folder name.

### Closer

`go.mod` says `module groundwork.local/packages`. Add `/inventory` to the end and use that in the import in `deploy/deploy.go`.

### Answer

```go
import (
	"strings"

	"groundwork.local/packages/inventory"
)
```

`strings` is still used until Task 3.

## Task 3: Move NormalizeHost into inventory

Move the whole function, with its doc comment, from `deploy/deploy.go` to the end of `inventory/inventory.go`. Then fix the imports in both files and the call in `deploy`.

### Closer

`inventory/inventory.go` now needs `"fmt"` and `"strings"`. In `deploy/deploy.go`, the call becomes `inventory.NormalizeHost(host)`, and the only import left is the inventory one.

### Answer

In `inventory/inventory.go`:

```go
import (
	"fmt"
	"strings"
)

// NormalizeHost trims spaces from both ends of a host name and lower-cases it.
func NormalizeHost(name string) string {
	return strings.ToLower(strings.TrimSpace(name))
}
```

In `deploy/deploy.go`:

```go
import "groundwork.local/packages/inventory"

// in the loop in Targets:
targets = append(targets, inventory.Address(inventory.NormalizeHost(host), port))
```
