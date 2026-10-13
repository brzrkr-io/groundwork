## Task 1: one line

The directive sits on the line right above the variable, with no space
after `//`:

```go
//go:embed defaults.json
var defaultConfig []byte
```

## Task 2: TrimSuffix

Add the directive the same way as in task 1, then wrap the name in the loop:

```go
kinds = append(kinds, strings.TrimSuffix(entry.Name(), ".txt"))
```

`ReadDir` already sorts by name.

## Task 3: ParseFS, then Execute

Build the path with `"templates/" + kind + ".txt"`, pass it to
`template.ParseFS(templates, path)`, and execute the result into a
`strings.Builder`. Return `"", err` as soon as either step fails.
