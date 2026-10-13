## Task 1: Respect replacements in DepVersion

When you find the dependency, check whether it was replaced before you return its version.

### Closer

Inside the `if dep.Path == path` block, return `dep.Replace.Version` when `dep.Replace != nil`, and `dep.Version` otherwise.

### Answer

```go
if dep.Path == path {
	if dep.Replace != nil {
		return dep.Replace.Version
	}
	return dep.Version
}
```

## Task 2: Format the build time

Find the `vcs.time` setting, turn its text into a `time.Time`, and turn that back into text in the new layout. Anything that goes wrong gives `"unknown"`.

### Closer

Check `info == nil` first. When `s.Key == "vcs.time"`, use `time.Parse(time.RFC3339, s.Value)`. Then `t.UTC().Format("2006-01-02 15:04")` and add `" UTC"`.

### Answer

```go
func BuildTime(info *debug.BuildInfo) string {
	if info == nil {
		return "unknown"
	}
	for _, s := range info.Settings {
		if s.Key != "vcs.time" {
			continue
		}
		t, err := time.Parse(time.RFC3339, s.Value)
		if err != nil {
			return "unknown"
		}
		return t.UTC().Format("2006-01-02 15:04") + " UTC"
	}
	return "unknown"
}
```

## Task 3: Write the report

Build one line per path, in a loop, using `DepVersion`.

### Closer

Start with `var lines []string` and `append` `path+" "+DepVersion(info, path)` for each path.

### Answer

```go
func Report(info *debug.BuildInfo, paths []string) []string {
	var lines []string
	for _, path := range paths {
		lines = append(lines, path+" "+DepVersion(info, path))
	}
	return lines
}
```
