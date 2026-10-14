## Task 1: Pick the version

Check the three rules in order and return as soon as one matches. The last rule is what's left when nothing else matched.

### Closer

Use `if stamped != "" { return stamped }`. Then check `info != nil` first, before reading `info.Main.Version`, and combine the conditions with `&&`.

### Answer

```go
func Version(stamped string, info *debug.BuildInfo) string {
	if stamped != "" {
		return stamped
	}
	if info != nil && info.Main.Version != "" && info.Main.Version != "(devel)" {
		return info.Main.Version
	}
	return "dev"
}
```

## Task 2: Find the commit

Loop over `info.Settings` once and remember the two values you care about. After the loop, decide what to return.

### Closer

Check `info == nil` before the loop. Inside it, `switch s.Key` with cases `"vcs.revision"` and `"vcs.modified"`. `revision[:7]` crashes on short strings, so only slice when `len(revision) > 7`.

### Answer

```go
func Revision(info *debug.BuildInfo) string {
	if info == nil {
		return "unknown"
	}
	var revision string
	dirty := false
	for _, s := range info.Settings {
		switch s.Key {
		case "vcs.revision":
			revision = s.Value
		case "vcs.modified":
			dirty = s.Value == "true"
		}
	}
	if revision == "" {
		return "unknown"
	}
	if len(revision) > 7 {
		revision = revision[:7]
	}
	if dirty {
		revision += "-dirty"
	}
	return revision
}
```

## Task 3: Build the version line

Call your two functions and put their results into one line.

### Closer

`fmt.Sprintf` with three `%s` verbs: `"%s %s (commit %s)"`.

### Answer

```go
return fmt.Sprintf("%s %s (commit %s)", name, Version(stamped, info), Revision(info))
```
