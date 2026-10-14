## Task 1: Write the Shout function

Look at `Greet` in the README's "parts of a file" example. `Shout` has the
same shape: one string parameter and a string result. Only the names and the
body change.

### Closer

The signature line is `func Shout(word string) string {`. In the body, return
the parameter joined with the fixed text `"!"`, and close with `}`.

### Answer

```go
// Shout returns word followed by an exclamation mark.
func Shout(word string) string {
	return word + "!"
}
```

## Task 2: Announce a deploy

Two things to change. Take `service` out of the quotes so you use its value.
Then pass the whole text into `Shout` instead of returning it directly.

### Closer

`"deploy " + service` builds `deploy api`. Put that inside the brackets of a
call: `Shout(...)`. A call runs the function and gives back its result.

### Answer

```go
return Shout("deploy " + service)
```

## Task 3: Add a version to the banner

Constants live at the top level, outside any function, just like `Tool`.

### Closer

Declare it with `const Version = "v1.0.0"`. Then, in `Banner`, join three
pieces: the constant `Tool`, the fixed text `" "`, and the constant `Version`.

### Answer

```go
const Version = "v1.0.0"

func Banner() string {
	return Tool + " " + Version
}
```
