## Task 1: Build the address

Inside `func (h Host) Address() string`, the host is called `h`, so read its
fields with `h.IP` and `h.Port`. `Port` is an `int`, so you can't join it to a
string with `+`.

### Closer

`fmt.Sprintf` builds a string from a pattern: `%s` stands for a string and
`%d` for an int, like `fmt.Sprintf("%s:%d", ...)`. Remember to add `"fmt"` and
`"strings"` to the imports.

### Answer

```go
return fmt.Sprintf("%s:%d", h.IP, h.Port)
```

## Task 2: Build a URL

`h.Address()` already gives you `"10.0.0.5:8080"`. You only need to put text
before and after it.

### Closer

Join three pieces with `+`: the fixed text `"http://"`, the result of
`h.Address()`, and `path`.

### Answer

```go
return "http://" + h.Address() + path
```

## Task 3: Read the image tag

Split `d.Image` around the colon and keep the part after it. If there was no
colon at all, the answer is `"latest"`.

### Closer

`strings.Cut(d.Image, ":")` returns before, after and a `found` bool. You
don't need the part before, so use `_` for it: `_, tag, found := ...`.

### Answer

```go
_, tag, found := strings.Cut(d.Image, ":")
if !found {
	return "latest"
}
return tag
```
