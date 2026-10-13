## Task 1: Fix the empty audit entries

`encoding/json` can only see exported fields, and a field is exported when its
name starts with a capital letter. Right now it sees none of them.

### Closer

Rename `at` to `At`, `user` to `User`, and so on, and give each a tag such as
``At time.Time `json:"time"` ``. Then update the struct literal in
`AuditJSON`: `auditRecord{At: at, User: user, ...}`.

### Answer

```go
type auditRecord struct {
	At     time.Time `json:"time"`
	User   string    `json:"user"`
	Action string    `json:"action"`
	Target string    `json:"target"`
}

rec := auditRecord{At: at, User: user, Action: action, Target: target}
```

The tags set the lowercase keys, while the capitals make the fields visible.

## Task 2: Fix the alert tags

Read each tag slowly, character by character. One key is misspelled, one tag
has a stray space, and one field should be left out when it's `false`.

### Closer

Tags allow no space after `json:`. `sevrity` should be `severity`, and
`silenced` needs `,omitempty` after its key.

### Answer

```go
Severity string `json:"severity"`
Region   string `json:"region"`
Silenced bool   `json:"silenced,omitempty"`
```

A malformed tag like `json: "region"` is silently ignored, so the key falls
back to the field name `Region`.

## Task 3: Write JSON Lines

A `json.Encoder` writes straight to an `io.Writer` and ends each value with a
newline. Make one encoder before the loop and use it for every alert.

### Closer

`enc := json.NewEncoder(w)`, then inside `for _, a := range alerts` call
`enc.Encode(a)` and return its error if it isn't `nil`.

### Answer

```go
enc := json.NewEncoder(w)
for _, a := range alerts {
	if err := enc.Encode(a); err != nil {
		return err
	}
}
return nil
```

`Encode` adds the `"\n"` itself, so each alert lands on its own line.
