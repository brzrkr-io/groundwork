## Task 1: capital letters

`encoding/json` can't see lowercase fields. Rename `at` to `At`, `user` to
`User`, and so on, and add a tag to each:

```go
At time.Time `json:"time"`
```

Then update the struct literal: `auditRecord{At: at, User: user, ...}`.

## Task 2: read each tag slowly

Look for a misspelled key, a space after `json:`, and a field that should
use `,omitempty`.

## Task 3: almost the answer

```go
enc := json.NewEncoder(w)
for _, a := range alerts {
	if err := enc.Encode(a); err != nil {
		return err
	}
}
return nil
```
