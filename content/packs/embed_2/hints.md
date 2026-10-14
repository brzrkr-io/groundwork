## Task 1: Fix the version directive

Compare the comment above `schemaVersion` with the one above `migrations`, character by character. Go only treats a comment as a directive if it's written exactly right.

### Closer

A directive starts with `//go:embed`, with no space between `//` and `go`. With a space, it's just an ordinary comment.

### Answer

```go
//go:embed version.txt
var schemaVersion string
```

## Task 2: Embed only the SQL files

Right now the directive embeds the whole folder. A pattern can pick out only the files you want.

### Closer

Replace the folder name with a glob pattern: `migrations/` followed by `*` and the `.sql` ending. `Names` doesn't need to change.

### Answer

```go
//go:embed migrations/*.sql
var migrations embed.FS
```

## Task 3: List the pending SQL

Get the sorted names from `Names`, skip the first `applied` of them, and read each remaining file.

### Closer

Loop with `for i, name := range names` and `continue` while `i < applied`. Read with `migrations.ReadFile("migrations/" + name)` and append `string(data)`. Return any error straight away.

### Answer

```go
func Pending(applied int) ([]string, error) {
	names, err := Names()
	if err != nil {
		return nil, err
	}
	var pending []string
	for i, name := range names {
		if i < applied {
			continue
		}
		data, err := migrations.ReadFile("migrations/" + name)
		if err != nil {
			return nil, err
		}
		pending = append(pending, string(data))
	}
	return pending, nil
}
```

When `applied` is 3 or more, every name is skipped and `pending` stays `nil`.
