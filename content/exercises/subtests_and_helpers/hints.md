## Task 1: an if and an Errorf

The helper is a normal check, as in *Your first test*: compare, and call
`t.Errorf` if the values differ. The only new parts are `t.Helper()` and
`slices.Equal`.

## Task 1, nearly done

```go
import (
	"slices"
	"testing"
)

func assertHosts(t *testing.T, got, want []string) {
	t.Helper()
	if !slices.Equal(got, want) {
		t.Errorf("got %q, want %q", got, want)
	}
}
```

## Task 2: copy the example row

Each new case is a row like the `"one host per line"` row. You don't need to
touch the loop: it runs every row as a subtest.

## Task 2, nearly done

```go
{
	name: "skips comments",
	text: "# web tier\nweb-1\n# database\ndb-1",
	want: []string{"web-1", "db-1"},
},
```

Add `"skips blank lines"` and `"trims spaces"` the same way, with the values
from the README table.
