## Task 1: Finish the helper

The helper is a normal check, like in *Your first test*: compare `got` with `want`, and call `t.Errorf` if they differ. The new parts are `t.Helper()` and `slices.Equal`.

### Closer

You can't compare slices with `!=`. Use `!slices.Equal(got, want)` instead, and add `"slices"` to the imports. Put `t.Helper()` on the first line, so failures point at the caller.

### Answer

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

## Task 2: Add the cases

Each new case is a row like the `"one host per line"` row. You don't need to touch the loop, because it already runs every row as a subtest.

### Closer

Copy the example row, with its `name:`, `text:` and `want:` fields, and fill in the values from the README table.

### Answer

```go
{
	name: "skips comments",
	text: "# web tier\nweb-1\n# database\ndb-1",
	want: []string{"web-1", "db-1"},
},
{
	name: "skips blank lines",
	text: "web-1\n\nweb-2\n\n",
	want: []string{"web-1", "web-2"},
},
{
	name: "trims spaces",
	text: "  web-1  \n\tdb-1",
	want: []string{"web-1", "db-1"},
},
```
