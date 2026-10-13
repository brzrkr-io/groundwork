## Task 1: one rename

Change `func NewQuota(` to `func New(` and fix the doc comment to start with
`New`.

## Task 2: field first, then methods

After renaming the field to `limit`, the constructor must set `limit: limit`.
Then add:

```go
// Limit returns how many deploys the quota allows.
func (q *Quota) Limit() int {
	return q.limit
}
```

`Used` follows the same pattern with `q.used`.

## Task 3: check before counting

```go
func (q *Quota) Take() error {
	if q.used >= q.limit {
		return ErrExhausted
	}
	q.used++
	return nil
}
```
