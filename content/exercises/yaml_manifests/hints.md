## Task 1: it's json.Unmarshal with a different import

`ReadType` shows the pattern. Declare `var d appsv1.Deployment`, pass `&d`
to `yaml.Unmarshal`, and return `&d` when there is no error. Wrap errors
with `fmt.Errorf("parse deployment: %w", err)` (import `"fmt"`).

## Task 2: two different errors

Replace `return tm, err` with a return that wraps the error using
`"read type: %w"`. After the `if`, check `tm.Kind == ""` and return
`errors.New("read type: missing kind")` (import `"errors"`).

## Task 3: the loop

`current` holds the lines of the document you are reading. Loop with
`for line := range strings.SplitSeq(stream, "\n")`. When the line is `"---"`,
hand `current` to `addDocument` and reset it to `nil`. Otherwise append the
line.

## Task 3: almost the answer

```go
for line := range strings.SplitSeq(stream, "\n") {
	if line == "---" {
		docs = addDocument(docs, current)
		current = nil
		continue
	}
	current = append(current, line)
}
return addDocument(docs, current)
```
