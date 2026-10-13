## Task 1: look at the comment

Compare the line above `schemaVersion` with the one above `migrations`.
A directive starts with `//go:embed`, with no space.

## Task 2: a glob pattern

Replace the folder name with a pattern that only matches SQL files:
`//go:embed migrations/*.sql`. `Names` doesn't need to change.

## Task 3: skip, then read

```go
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
```
