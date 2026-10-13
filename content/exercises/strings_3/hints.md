## Task 1: build the stars first

`strings.Repeat("*", len(secret))` gives as many stars as the secret has
characters. Pass that to `strings.ReplaceAll` as the replacement.

## Task 2: one step per line

A variable keeps it readable:

```go
host := strings.TrimPrefix(url, "https://")
host = strings.TrimSuffix(host, "/")
```

Then return `host` in lowercase.

## Task 3: join, replace twice, upper-case

Join the parts with `service + "_" + setting`. Then call `strings.ReplaceAll`
once for `"-"` and once for `" "`, and finish with `strings.ToUpper`.

## Almost the answer

```go
func EnvVarName(service, setting string) string {
	name := service + "_" + setting
	name = strings.ReplaceAll(name, "-", "_")
	name = strings.ReplaceAll(name, " ", "_")
	return strings.ToUpper(name)
}
```
