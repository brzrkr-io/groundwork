## Task 1: Parse a Deployment

This is `json.Unmarshal` with a different import. `ReadType` below already
shows a working `yaml.Unmarshal` call you can copy the shape of.

### Closer

Declare `var d appsv1.Deployment`, pass `&d` to `yaml.Unmarshal`, and return
`&d` when there is no error. Wrap errors with
`fmt.Errorf("parse deployment: %w", err)` (import `"fmt"`).

### Answer

```go
var d appsv1.Deployment
if err := yaml.Unmarshal(data, &d); err != nil {
	return nil, fmt.Errorf("parse deployment: %w", err)
}
return &d, nil
```

`yaml.Unmarshal` fills `d` through the pointer, and you return its address.

## Task 2: Read the type

There are two different errors: one when unmarshalling fails, and one when
it works but `kind` is missing.

### Closer

Replace `return tm, err` with a return that wraps the error using
`"read type: %w"`. After the `if`, check `tm.Kind == ""` and return
`errors.New(...)` (import `"errors"`).

### Answer

```go
if err := yaml.Unmarshal(data, &tm); err != nil {
	return tm, fmt.Errorf("read type: %w", err)
}
if tm.Kind == "" {
	return tm, errors.New("read type: missing kind")
}
return tm, nil
```

A missing key just leaves the field empty, so you have to check it yourself.

## Task 3: Split a stream into documents

`current` holds the lines of the document you are reading. Go through the
stream line by line and decide, for each line, whether it ends a document.

### Closer

Loop with `for line := range strings.SplitSeq(stream, "\n")`. When the line is
`"---"`, hand `current` to `addDocument` and reset it to `nil`. Otherwise
append the line.

### Answer

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

Comparing the whole line with `"---"` means `a---b` stays ordinary text.
