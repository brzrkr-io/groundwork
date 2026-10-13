## Task 1: check in order

Return early as soon as one rule matches:

```go
if stamped != "" {
	return stamped
}
if info != nil && info.Main.Version != "" && info.Main.Version != "(devel)" {
	return info.Main.Version
}
return "dev"
```

## Task 2: one loop, two variables

Loop over `info.Settings` once and remember what you find:

```go
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
```

Check `info == nil` before the loop. If `revision` is still `""` afterwards,
return `"unknown"`.

## Task 2: shortening safely

`revision[:7]` crashes if the string has fewer than 7 characters. Only slice
when `len(revision) > 7`.

## Task 3: fmt.Sprintf

```go
return fmt.Sprintf("%s %s (commit %s)", name, Version(stamped, info), Revision(info))
```
