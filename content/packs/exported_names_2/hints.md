## Task 1: Rename the constructor

The package is already called `quota`, so callers write `quota.NewQuota`, which repeats itself. Rename the function so calls read `quota.New`.

### Closer

Change `func NewQuota(` to `func New(`, and make the doc comment start with `New`.

### Answer

```go
// New returns a Quota that allows limit deploys.
func New(limit int) *Quota {
	return &Quota{Limit: limit}
}
```

## Task 2: Make the fields private and add getters

First make the `Limit` field lower case, so callers can't change it. Then add small methods that only return the values, so callers can still read them.

### Closer

After the rename, the constructor sets `limit: limit`. A getter looks like `func (q *Quota) Limit() int { return q.limit }`. `Used` follows the same pattern with `q.used`.

### Answer

```go
type Quota struct {
	limit int
	used  int
}

func New(limit int) *Quota {
	return &Quota{limit: limit}
}

// Limit returns how many deploys the quota allows.
func (q *Quota) Limit() int {
	return q.limit
}

// Used returns how many deploys have been taken.
func (q *Quota) Used() int {
	return q.used
}
```

## Task 3: Enforce the limit in Take

Check before you count. If the quota is already used up, say no and leave the count alone.

### Closer

`if q.used >= q.limit { return ErrExhausted }`, then `q.used++` and `return nil`. Using `>=` also handles a limit of 0.

### Answer

```go
func (q *Quota) Take() error {
	if q.used >= q.limit {
		return ErrExhausted
	}
	q.used++
	return nil
}
```
