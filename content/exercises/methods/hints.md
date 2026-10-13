## Use the receiver like a parameter

Inside `func (h Host) Address() string`, the host is called `h`. Read its
fields with `h.IP` and `h.Port`, the same way you would with any struct.
Remember to add `"fmt"` and `"strings"` to the imports.

## Task 1: format a string and an int

`Port` is an `int`, so you can't add it to a string with `+`. `fmt.Sprintf`
with `%s` for the IP and `%d` for the port builds the whole address in one go.

## Task 2: call the other method

`h.Address()` already gives you `"10.0.0.5:8080"`. Put `"http://"` in front
and `path` after it with `+`.

## Task 3: almost the answer

```go
_, tag, found := strings.Cut(d.Image, ":")
if !found {
	return "latest"
}
return tag
```
