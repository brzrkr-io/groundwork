## Task 1: Embed the default config

A `//go:embed` comment tells Go to put a file's contents into the variable on the very next line, when it builds the program.

### Closer

Write the directive on the line right above `var defaultConfig []byte`, with no space after `//`: `//go:embed <file name>`.

### Answer

```go
//go:embed defaults.json
var defaultConfig []byte
```

## Task 2: Embed the templates folder

First embed the folder the same way as in Task 1. Then, in `Kinds`, each name needs its `.txt` ending removed before you add it to the list.

### Closer

`//go:embed templates` goes above `var templates embed.FS`. In the loop, `strings.TrimSuffix(entry.Name(), ".txt")` drops the ending; add `"strings"` to the imports. `ReadDir` already sorts by name.

### Answer

```go
//go:embed templates
var templates embed.FS
```

```go
for _, entry := range entries {
	kinds = append(kinds, strings.TrimSuffix(entry.Name(), ".txt"))
}
return kinds, nil
```

## Task 3: Render a notification

There are two steps: load the template file from the embedded files, then run it with `d` and collect the text. Either step can fail.

### Closer

Build the path with `"templates/" + kind + ".txt"` and pass it to `template.ParseFS(templates, path)`. Run it with `tmpl.Execute(&b, d)`, where `b` is a `strings.Builder`. Return `"", err` as soon as a step fails.

### Answer

```go
func Render(kind string, d Deploy) (string, error) {
	tmpl, err := template.ParseFS(templates, "templates/"+kind+".txt")
	if err != nil {
		return "", err
	}
	var b strings.Builder
	if err := tmpl.Execute(&b, d); err != nil {
		return "", err
	}
	return b.String(), nil
}
```
