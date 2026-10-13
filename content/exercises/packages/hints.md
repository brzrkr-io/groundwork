## Task 1: one folder, one name

Every `.go` file in the `inventory` folder must start with the same
`package` line. Look at the first line of `inventory/inventory_test.go` and
make `inventory/inventory.go` match it.

## Task 2: module path plus folder

Open `go.mod` and copy the module path. Add `/inventory` to the end. That is
the import path:

```go
import (
	"strings"

	"gopherlings.local/packages/inventory"
)
```

## Task 3: move the function and its import

After the move, `inventory/inventory.go` needs both imports:

```go
import (
	"fmt"
	"strings"
)
```

`deploy/deploy.go` keeps only the inventory import, and the loop becomes:

```go
targets = append(targets, inventory.Address(inventory.NormalizeHost(host), port))
```
